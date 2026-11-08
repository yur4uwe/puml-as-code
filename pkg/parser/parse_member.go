package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) routeEntityMember(tok tokenizer.Token) ast.Member {
	collector := &stream.TokenCollector{}
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
	p.ast.RecordDiagnostic(diagst)
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
	} else if err != stream.ErrStartMarkerNotFound {
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
		} else if err != stream.ErrStartMarkerNotFound {
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
	targetRef := p.parseQualifiedName(firstTok)
	im := &ast.InlineMember{
		Target: targetRef,
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.COLON); !ok {
		nextTok := p.stream.PeekTokenAt(0)
		err := NewParserError(nextTok, "Expected ':' after entity identifier for inline member declaration")
		p.ast.RecordError(
			err, nextTok.Span, im,
		)
		p.syncToLineEnd()
		return im, nil
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
		if err == nil {
			mod = &m
			break
		}
		fallthrough
	case tokenizer.IDENTIFIER:
		p.stream.Emit()
	}
	memberMark := p.Mark(entryTok)
	member, err := p.parseFieldOrMethod(mod, vis, entryTok)
	if err != nil {
		p.ast.RecordError(
			err, p.Span(memberMark), im,
		)
		p.syncToLineEnd()
		return im, nil
	}
	p.stream.EmitCommentToks()
	member.SetMetadata(p.Span(memberMark), p.dumpTrivia(), p.dumpTrivia())
	im.Member = member

	return im, nil
}
