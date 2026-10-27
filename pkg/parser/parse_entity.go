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
		p.recordDiagnostic(p.newDiagnosticError(err, identTok.Span, ent))
		p.syncToBlockStart()
		return
	}

	if _, hasAlias := p.stream.TryConsumeKW(keyword.Alias); hasAlias {
		secondIdentTok := p.stream.Emit()
		if err := p.setAliasAndName(ent, secondIdentTok); err != nil {
			p.recordDiagnostic(p.newDiagnosticError(err, secondIdentTok.Span, ent))
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
			p.recordDiagnostic(p.newDiagnosticError(err, span, ent))
			p.syncToBlockStart()
			return
		}
		ent.Generic = gen
	}

	preTags := p.tryReadTags()
	stereo, span, err := p.tryParseStereotype()
	if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
		p.recordDiagnostic(p.newDiagnosticError(err, span, ent))
		p.syncToBlockStart()
		return
	}
	ent.Stereotype = stereo
	postTags := p.tryReadTags()
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
		// No body return entity as is
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
			p.recordDiagnostic(p.newDiagnosticError(unclosedBlockErr, lbraceTok.Span, ent))
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

func (p *Parser) routeEntityMember(tok tokenizer.Token) ast.Member {
	collector := &tokenizer.TokenCollector{}
	detach := p.stream.Attach(collector)
	defer detach()

	leadingTrivia := p.dumpTrivia()

	memberMark := p.Mark(tok)
	member, err := p.parseEntityMember()
	if err == nil {
		if member == nil {
			return nil
		}
		detach()
		p.stream.EmitCommentToks()
		member.SetMetadata(p.Span(memberMark), leadingTrivia, p.dumpTrivia())
		return member
	}

	p.syncToLineOrBlockEnd()
	if collector.Len() == 0 && !p.stream.AssertType(tokenizer.EOF) && !p.isEndMarker() && !p.stream.AssertType(tokenizer.RBRACE) {
		p.stream.Emit()
	}
	detach()
	badToks := collector.Tokens()
	span := p.Span(memberMark)

	diagst := &ast.Diagnostic{
		ID:      len(p.ast.Diagnostics),
		Message: err.Error(),
		Span:    span,
		Err:     err,
	}

	p.stream.EmitCommentToks()
	badMember := &ast.BadMember{
		Diagnostic: diagst,
		Tokens:     badToks,
		BaseNode: ast.BaseNode{
			NodeSpan:       span,
			LeadingTrivia:  leadingTrivia,
			TrailingTrivia: p.dumpTrivia(),
		},
	}

	diagst.Node = badMember
	p.recordDiagnostic(diagst)
	return badMember
}

func (p *Parser) parseEntityMember() (ast.Member, error) {
	var member ast.Member
	var err error

	// Modifiers and separators precede the switch to not mistake -- separator and '-' for visibility
	// PLAN: rewrite the class separator to use the same logic as members,
	// just make them assert on extracted token slice
	if member, err = p.tryReadClassSeparator(); err == nil {
		return member, nil
	}

	vis := ast.VisibilityUnknown
	if mod, err := p.tryReadModifier(); err == nil {
		// Handle scope modifiers
		member, err = p.parseFieldOrMethod(&mod, vis, unamb(tokenizer.LBRACE))
		if err != nil {
			return nil, err
		}
		return member, nil
	} else if err != tokenizer.ErrStartMarkerNotFound {
		// Ignore the particular error because of the reasons above
		return nil, err
	}

	tok := p.stream.Emit()
	switch tok.Type {
	case tokenizer.RBRACE:
		// It shouldn't arrive here, but just in case
		return nil, nil
	case tokenizer.DASH, tokenizer.TILDE, tokenizer.HASH, tokenizer.PLUS:
		vis = p.mapTokenToVisibility(tok.Type)
	case tokenizer.IDENTIFIER:
		// simply to not error out
	case tokenizer.NEWLINE:
		// We encountered a comment
		// We should retry the whole loop
	default:
		return nil, NewParserError(tok, "Unexpected token in entity body")
	}
	return p.parseFieldOrMethod(nil, vis, tok)
}

// Can be either a field or a method
// but start with a field as either start the same
//
// Returns the parsed field or method, must be asserted to be a field or method
func (p *Parser) parseFieldOrMethod(mod *string, vis ast.VisibilityKind, entryTok tokenizer.Token) (ast.Member, error) {
	// Can enter this function with one of:
	// - Name
	// - Visibility
	// - Modifier
	// It makes the parsing of the field or method easier
	// NOTE: Modifier can be in any position, but visibility can only be before the name

	// This flag is used to determine if we can encounter visibility token
	canEncounterVisibility := true
	mustBeField := false
	mustBeMethod := false
	containsLParen := false
	var mods []string = nil

	if mod != nil {
		switch *mod {
		case "method":
			mustBeMethod = true
		case "field":
			mustBeField = true
		}
		mods = append(mods, *mod)
	}

	if vis != ast.VisibilityUnknown {
		canEncounterVisibility = false
	}

	entry := make([]tokenizer.Token, 0, 10)
	if entryTok.Type == tokenizer.IDENTIFIER {
		entry = append(entry, entryTok)
	}

outer:
	for {
		mod, err := p.tryReadModifier()
		if err == nil {
			switch mod {
			case "method":
				mustBeMethod = true
			case "field":
				mustBeField = true
			}
			mods = append(mods, mod)
			continue
		} else if err != tokenizer.ErrStartMarkerNotFound {
			return nil, err
		}

		if p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
			break outer
		}

		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.HASH, tokenizer.PLUS, tokenizer.DASH, tokenizer.TILDE:
			if canEncounterVisibility {
				vis = p.mapTokenToVisibility(tok.Type)
				canEncounterVisibility = false
				continue
			}
			entry = append(entry, tok)
		case tokenizer.LPAREN:
			containsLParen = true
			fallthrough
		default:
			entry = append(entry, tok)
		}
	}

	if mustBeField && mustBeMethod {
		return nil, NewParserError(entryTok, "Cannot be field and method at the same time")
	}

	isMethod := mustBeMethod || (!mustBeField && containsLParen)

	if isMethod {
		method, err := p.Dialect.ParseMethod(entry)
		if err != nil {
			return nil, err
		}
		return &ast.MethodDeclaration{
			Visibility: vis,
			Modifiers:  mods,
			Method:     method,
		}, nil
	} else {
		field, err := p.Dialect.ParseField(entry)
		if err != nil {
			return nil, err
		}
		return &ast.FieldDeclaration{
			Visibility: vis,
			Modifiers:  mods,
			Field:      field,
		}, nil
	}
}

func (p *Parser) parseInlineMember(firstTok tokenizer.Token) (ast.Statement, error) {
	inlineMark := p.Mark(firstTok)
	targetRef, err := p.parseTargetRef(firstTok)
	if err != nil {
		return nil, err
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.COLON); !ok {
		return nil, NewParserError(p.stream.PeekTokenAt(0), "Expected ':' after entity identifier for inline member declaration")
	}

	entryTok := p.stream.PeekTokenAt(0)
	vis := ast.VisibilityUnknown
	var mod *string = nil
	switch entryTok.Type {
	case tokenizer.DASH, tokenizer.TILDE, tokenizer.HASH, tokenizer.PLUS:
		p.stream.Emit()
		vis = p.mapTokenToVisibility(entryTok.Type)
	case tokenizer.LBRACE:
		m, err := p.tryReadModifier()
		if err != nil {
			return nil, err
		}
		mod = &m
	case tokenizer.IDENTIFIER:
		p.stream.Emit()
	}
	memberMark := p.Mark(entryTok)
	member, err := p.parseFieldOrMethod(mod, vis, entryTok)
	if err != nil {
		return nil, err
	}
	p.stream.EmitCommentToks()
	member.SetMetadata(p.Span(memberMark), p.dumpTrivia(), p.dumpTrivia())

	im := &ast.InlineMember{
		Target: targetRef,
		Member: member,
	}
	im.SetMetadata(p.Span(inlineMark), p.dumpTrivia(), p.dumpTrivia())

	return im, nil
}
