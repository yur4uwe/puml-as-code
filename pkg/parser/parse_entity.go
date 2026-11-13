package parser

import (
	"errors"
	"fmt"
	"log"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"
)

func (p *Parser) parseEntityStyle() (*ast.EntityStyle, tokenizer.SourceSpan, error) {
	entStyle := &ast.EntityStyle{}

	tokenAfterHash := p.stream.PeekTokenAt(0)
	if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.NUMBER) {
		goto trailingLine
	}

	// First item: bare background color without a target
	switch p.stream.PeekTokenAt(0).Literal {
	case "back", "header", "line", "text":
	default:
		tok := p.stream.Emit()
		bg, err := p.parseColorExpr(tok)
		if err != nil {
			return entStyle, p.stream.LastSemanticToken().Span, err
		}
		entStyle.Background = bg
		if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); !ok {
			goto trailingLine
		}
	}

	// Remaining items: strictly target:color pairs
	for {
		tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
		if !ok {
			return entStyle, tok.Span, fmt.Errorf("unexpected token describing entity colorable target: %s(%s)", tok.Type, tok.Literal)
		}

		targetKind := ""
		switch tok.Literal {
		case "line":
			targetKind = tok.Literal
			if _, ok = p.stream.TryConsumeType(tokenizer.DOT); !ok {
				break // no dot — fall through to colon:color
			}

			lineStyleTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
			if !ok {
				return entStyle, lineStyleTok.Span, fmt.Errorf("unexpected token describing line style: %s(%s)", lineStyleTok.Type, lineStyleTok.Literal)
			}

			ls, err := parseLineStyle(lineStyleTok)
			if err != nil {
				return entStyle, lineStyleTok.Span, err
			}
			entStyle.LineStyle = ls

			if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); ok {
				continue
			}
			// no semicolon — fall through to parse optional :color
		case "back", "header", "text":
			targetKind = tok.Literal
		default:
			return entStyle, tok.Span, fmt.Errorf("unexpected value for color target specifier: %s", tok.Literal)
		}

		if _, ok := p.stream.TryConsumeType(tokenizer.COLON); !ok {
			peekTok := p.stream.PeekTokenAt(0)
			return entStyle, peekTok.Span, errors.New("missing colon after specified color target")
		}

		colorTok := p.stream.Emit()
		expr, err := p.parseColorExpr(colorTok)
		if err != nil {
			return entStyle, p.stream.LastSemanticToken().Span, err
		}

		switch targetKind {
		case "header":
			entStyle.Header = expr
		case "text":
			entStyle.TextColor = expr
		case "line":
			entStyle.LineColor = expr
		case "back":
			entStyle.Background = expr
		}

		if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); !ok {
			break
		}
	}

trailingLine:
	// Trailing ##[lineStyle]lineColor after #color or composite styles
	if p.stream.AssertTypeSeq(tokenizer.HASH, tokenizer.HASH) {
		p.stream.Emit() // consume first #
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.HASH); ok {
		if _, ok := p.stream.TryConsumeType(tokenizer.LBRACKET); ok {
			lineStyleTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
			if !ok {
				return entStyle, lineStyleTok.Span, fmt.Errorf("unexpected token describing line style: %s(%s)", lineStyleTok.Type, lineStyleTok.Literal)
			}

			ls, err := parseLineStyle(lineStyleTok)
			if err != nil {
				return entStyle, lineStyleTok.Span, err
			}
			entStyle.LineStyle = ls

			if _, ok := p.stream.TryConsumeType(tokenizer.RBRACKET); !ok {
				peekTok := p.stream.PeekTokenAt(0)
				return entStyle, peekTok.Span, fmt.Errorf("unexpected token in line style syntax: %s(%s)", peekTok.Type, peekTok.Literal)
			}
		}

		lineColorTok := p.stream.Emit()
		if lineColorTok.Type != tokenizer.IDENTIFIER && lineColorTok.Type != tokenizer.NUMBER {
			return entStyle, lineColorTok.Span, fmt.Errorf("unexpected token describing line color: %s(%s)", lineColorTok.Type, lineColorTok.Literal)
		}

		entStyle.LineColor = ast.FromToken(lineColorTok)
		return entStyle, tokenizer.SourceSpan{}, nil
	}

	switch tokenAfterHash.Type {
	case tokenizer.IDENTIFIER, tokenizer.NUMBER, tokenizer.HASH:
	default:
		return entStyle, tokenAfterHash.Span, fmt.Errorf("unexpected token in style declaration: %s(%s)", tokenAfterHash.Type, tokenAfterHash.Literal)
	}

	return entStyle, tokenizer.SourceSpan{}, nil
}

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
			return NewParserErrorAtIdent(targetRef.Name, "entity name contains unexpected member")
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
		if err != nil && !errors.Is(err, stream.ErrStartMarkerNotFound) {
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
	if err != nil && !errors.Is(err, stream.ErrStartMarkerNotFound) {
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

	if hashTok, ok := p.stream.TryConsumeType(tokenizer.HASH); ok {
		mark := p.Mark(hashTok)
		style, errSpan, err := p.parseEntityStyle()
		style.Span = p.Span(mark)
		ent.Style = style
		if err != nil {
			p.ast.RecordError(err, errSpan, ent)
			p.syncToBlockStart()
			return
		}
	}
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
