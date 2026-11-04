package parser

import (
	"errors"
	"log"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) setAliasAndName(ent *ast.Entity, nameOrAlias tokenizer.Token) error {
	switch nameOrAlias.Type {
	case tokenizer.STRING:
		if !ent.Alias.IsZero() {
			return NewParserError(nameOrAlias, "entity alias already set")
		}
		ent.Alias = ast.IdentFromToken(nameOrAlias)
		return nil
	case tokenizer.IDENTIFIER:
		if !ent.Identifier.Name.IsZero() {
			return NewParserError(nameOrAlias, "entity name already set")
		}
		targetRef, err := p.parseTargetRef(nameOrAlias)
		if err != nil || targetRef.Member != nil {
			return NewParserErrorFromIdent(targetRef.Name, "entity name contains unexpected member")
		}
		ent.Identifier = targetRef
		return nil
	default:
		return NewParserError(nameOrAlias, "expected token for entity identifier or alias")
	}
}

func (p *Parser) parseEntityHeader(ent *ast.Entity) {
	// Unexpectedly, class and other entity definitions have very strict syntax:
	// class <entity name> as <entity alias> <generics> <stereotype> <styles> <body>

	identTok := p.stream.Emit()
	err := p.setAliasAndName(ent, identTok)
	if err != nil {
		p.ast.RecordError(err, identTok.Span, ent)
		p.syncToBlockStart()
		return
	}

	if _, hasAlias := p.stream.TryConsumeKW(keyword.Alias); hasAlias {
		secondIdentTok := p.stream.Emit()
		if err := p.setAliasAndName(ent, secondIdentTok); err != nil {
			p.ast.RecordError(err, secondIdentTok.Span, ent)
			p.syncToBlockStart()
			return
		}
	}

	if ent.Identifier.Name.IsZero() {
		// If no identifier is set, use the alias
		ent.Identifier = ast.TargetRef{Name: ent.Alias}
		ent.Alias = ast.Ident{}
	}

	if !p.stream.AssertTypeSeq(tokenizer.LANGLE, tokenizer.LANGLE) {
		gen, span, err := p.tryReadGeneric()
		if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
			p.ast.RecordError(err, span, ent)
			p.syncToBlockStart()
			return
		}
		ent.Generic = gen
	}

	preTags, span, err := p.tryReadTags()
	if err != nil {
		p.ast.RecordError(err, span, ent)
		p.syncToBlockStart()
		return
	}

	stereo, span, err := p.tryParseStereotype()
	if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
		p.ast.RecordError(err, span, ent)
		p.syncToBlockStart()
		return
	}
	ent.Stereotype = stereo

	postTags, span, err := p.tryReadTags()
	if err != nil {
		p.ast.RecordError(err, span, ent)
		p.syncToBlockStart()
		return
	}

	if len(preTags) > 0 {
		if len(postTags) > 0 {
			log.Printf("warning: tags %v after stereotype on entity %q are ignored in favor of pre-stereotype tags %v\n", postTags, ent.Identifier.EntityName(), preTags)
		}
		ent.Tags = preTags
	} else if len(postTags) > 0 {
		ent.Tags = postTags
	}

	ent.Color = p.tryParseColor()
}

// tok is the kind of an entity (class, interface, struct, enum, etc.)
func (p *Parser) parseEntity(tok tokenizer.Token) *ast.Entity {
	ent := &ast.Entity{
		Kind: p.mapTokenToEntityKind(tok),
	}

	p.parseEntityHeader(ent)

	lbraceTok, ok := p.stream.TryConsumeType(tokenizer.LBRACE)
	if !ok {
		// No body, return entity as is
		return ent
	}

	ent.LBrace = lbraceTok.Span.Start

	p.stream.EmitCommentToks()
	ent.TrailingTrivia = p.dumpTrivia()

	unclosedBlockErr := ast.NewUnclosedBlockError("class", "}")

	// We are inside a body
	for {
		for {
			if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
				break
			}
		}

		if p.stream.AssertType(tokenizer.EOF) || p.isEndMarker() {
			// place unclosed block diagnostic on the opener of the block
			p.ast.RecordError(unclosedBlockErr, lbraceTok.Span, ent)
			break
		} else if p.stream.AssertType(tokenizer.RBRACE) {
			break
		}

		member := p.routeEntityMember(p.stream.PeekTokenAt(0))
		if member == nil {
			break
		}
		ent.Members = append(ent.Members, member)
	}

	rbraceTok, ok := p.stream.TryConsumeType(tokenizer.RBRACE)
	if !ok {
		return ent
	}
	ent.RBrace = rbraceTok.Span.Start

	return ent
}
