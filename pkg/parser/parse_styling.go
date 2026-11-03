package parser

import (
	"errors"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseSkinparamRoot(tok tokenizer.Token) (ast.Statement, error) {
	leadingTrivia := p.dumpTrivia()
	m := p.Mark(tok)
	skinparam, err := p.parseSkinparamStatement(tok, true)
	if err != nil {
		return nil, err
	}
	p.stream.EmitCommentToks()
	skinparam.SetMetadata(p.Span(m), leadingTrivia, p.dumpTrivia())
	return skinparam, err
}

func (p *Parser) parseSkinparamTarget(firstTok tokenizer.Token) ([]tokenizer.Token, *ast.Stereotype, error) {
	var tokens []tokenizer.Token
	var stereo *ast.Stereotype

	switch firstTok.Type {
	case tokenizer.DOT, tokenizer.UNDERSCORE, tokenizer.DASH, tokenizer.IDENTIFIER,
		tokenizer.NUMBER:
		tokens = append(tokens, firstTok)
	default:
		return nil, nil, NewParserError(firstTok, "expected valid skinparam target name")
	}
	prevTokEnd := firstTok.EndOffset()

	// Collect contiguous key tokens (e.g. "class.BackgroundColor")
	for {
		nextTok := p.stream.PeekTokenAt(0)
		if nextTok.Span.Start.Offset != prevTokEnd {
			break
		}
		switch nextTok.Type {
		case tokenizer.NEWLINE, tokenizer.SEMICOLON, tokenizer.EOF, tokenizer.LBRACE:
			return tokens, stereo, nil
		case tokenizer.DOT, tokenizer.UNDERSCORE, tokenizer.DASH, tokenizer.IDENTIFIER, tokenizer.NUMBER:
			tokens = append(tokens, p.stream.Emit())
			prevTokEnd = nextTok.EndOffset()
		case tokenizer.LANGLE:
			if stereo != nil {
				return tokens, stereo, NewParserError(nextTok, "multiple stereotypes")
			}
			var err error
			stereo, _, err = p.tryParseStereotype()
			if errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
				return tokens, stereo, NewParserError(nextTok, "unexpected token in stereotype target declaration")
			}
			if err != nil {
				return tokens, stereo, err
			}
			prevTokEnd = stereo.FullSpan.End.Offset
		default:
			return tokens, stereo, NewParserError(nextTok, "unexpected token in stereotype target declaration")
		}
	}

	if stereo != nil && p.stream.AssertTypeSeq(tokenizer.LANGLE, tokenizer.LANGLE) {
		return tokens, stereo, NewParserError(p.stream.PeekTokenAt(0), "multiple stereotypes")
	}
	if stereo == nil {
		var err error
		stereo, _, err = p.tryParseStereotype()
		if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
			return tokens, nil, err
		}
	}

	return tokens, stereo, nil
}

func (p *Parser) parseSkinparamStatement(tok tokenizer.Token, isRoot bool) (ast.Statement, error) {
	var keyword string
	var targetTokens []tokenizer.Token
	var stereo *ast.Stereotype
	var err error
	var nameStartTok tokenizer.Token

	if isRoot {
		keyword = tok.Literal

		// Check for anonymous block: `skinparam { ... }`
		if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); ok {
			p.stream.EmitCommentToks()
			trailingTrivia := p.dumpTrivia()
			block := &ast.SkinparamBlock{
				BaseNode: ast.BaseNode{
					TrailingTrivia: trailingTrivia,
				},
				Keyword: keyword,
			}
			err := p.parseSkinparamBlockMembers(block)
			return block, err
		}

		nameStartTok = p.stream.Emit()
		if nameStartTok.Type == tokenizer.NEWLINE || nameStartTok.Type == tokenizer.EOF {
			return nil, NewParserError(nameStartTok, "Expected target or parameter after skinparam")
		}
	} else {
		nameStartTok = tok
	}

	targetTokens, stereo, err = p.parseSkinparamTarget(nameStartTok)
	if err != nil {
		return nil, err
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); ok {
		p.stream.EmitCommentToks()
		trailingTrivia := p.dumpTrivia()
		block := &ast.SkinparamBlock{
			BaseNode: ast.BaseNode{
				TrailingTrivia: trailingTrivia,
			},
			Keyword:    keyword,
			Name:       ast.IdentFromTokens(targetTokens...),
			Stereotype: stereo,
		}
		err := p.parseSkinparamBlockMembers(block)
		return block, err
	}

	toks := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.SEMICOLON)
	value := p.stream.SliceInputEnclosingTokens(toks...)
	p.stream.TryConsumeType(tokenizer.SEMICOLON)
	return &ast.SkinparamSetting{
		Keyword:    keyword,
		Name:       ast.IdentFromTokens(targetTokens...),
		Value:      value,
		Stereotype: stereo,
	}, nil
}

func (p *Parser) parseSkinparamBlockMembers(block *ast.SkinparamBlock) error {
	for tok := p.stream.Emit(); tok.Type != tokenizer.RBRACE; tok = p.stream.Emit() {
		switch tok.Type {
		case tokenizer.NEWLINE, tokenizer.SEMICOLON:
			continue
		case tokenizer.EOF:
			return NewParserError(tok, "Unexpected EOF in skinparam block")
		}

		memberMark := p.Mark(tok)
		leadingTrivia := p.dumpTrivia()

		var member ast.Statement
		var err error
		switch tok.Type {
		case tokenizer.EXCLAMATION:
			member, err = p.parseDirective(tok)
		default:
			member, err = p.parseSkinparamStatement(tok, false)
		}
		if err != nil {
			p.recordDiagnostic(&ast.Diagnostic{
				ID:      len(p.ast.Diagnostics),
				Message: err.Error(),
				Span:    p.Span(memberMark),
				Err:     err,
			})
			continue
		}
		p.stream.EmitCommentToks()
		member.SetMetadata(p.Span(memberMark), leadingTrivia, p.dumpTrivia())
		block.Statements = append(block.Statements, member)
	}

	return nil
}

func (p *Parser) isStyleTagEnd() bool {
	return p.stream.AssertSeq([]tokenizer.Token{
		{Type: tokenizer.LANGLE},
		{Type: tokenizer.SLASH},
		{Type: tokenizer.IDENTIFIER, Literal: "style"},
		{Type: tokenizer.RANGLE},
	})
}

func (p *Parser) parseStyleBlock() (*ast.StyleBlock, error) {
	langle := p.stream.LastSemanticToken() // get the '<'
	p.stream.Emit()                        // consume 'style'
	rangle := p.stream.Emit()              // consume '>'
	blockMarkerSpan := tokenizer.SpanEnclosing(langle, rangle)

	styleBlock := &ast.StyleBlock{
		RAngle: rangle.Span.Start,
	}

	unclBlkErr := ast.NewUnclosedBlockError("<style>", "</style>")

	for !p.isStyleTagEnd() {
		tok := p.stream.Emit()
		if p.isEndMarker() || tok.Type == tokenizer.EOF {
			p.recordDiagnostic(p.newDiagnosticError(
				unclBlkErr,
				blockMarkerSpan,
				styleBlock,
			))
			break
		} else if tok.Type == tokenizer.NEWLINE || tok.Type == tokenizer.SEMICOLON {
			continue
		}

		ruleMark := p.Mark(tok)
		leadingTrivia := p.dumpTrivia()

		var stmnt ast.Statement
		var err error
		switch tok.Type {
		case tokenizer.EXCLAMATION:
			stmnt, err = p.parseDirective(tok)
		default:
			// Must be a StyleRule (e.g. "root { ... }" or "classDiagram { ... }")
			stmnt, err = p.parseStyleRule(tok)
		}
		if err != nil {
			p.recordDiagnostic(p.newDiagnosticError(
				err, tokenizer.SourceSpan{}, styleBlock,
			))
		} else {
			p.stream.EmitCommentToks()
			stmnt.SetMetadata(p.Span(ruleMark), leadingTrivia, p.dumpTrivia())
			styleBlock.Statements = append(styleBlock.Statements, stmnt)
		}
	}

	// Consume '</style>'
	if p.isStyleTagEnd() {
		if ltok, ok := p.stream.TryConsumeType(tokenizer.LANGLE); ok {
			styleBlock.LAngle = ltok.Span.Start
		}
		p.stream.TryConsumeType(tokenizer.SLASH)
		p.stream.TryConsume(tokenizer.Token{
			Type:    tokenizer.IDENTIFIER,
			Literal: "style",
		})
		p.stream.TryConsumeType(tokenizer.RANGLE)
	}

	return styleBlock, nil
}

func (p *Parser) setSelectorsFor(rule *ast.StyleRule, startTok tokenizer.Token) {
	// LBRACE is canonocal syntax but leave newline for error recovery
	toks := p.stream.ConsumeUntilType(tokenizer.LBRACE, tokenizer.NEWLINE)
	toks = append([]tokenizer.Token{startTok}, toks...)

	var selectorBuilder strings.Builder
	for _, t := range toks {
		if t.Type == tokenizer.COMMA {
			if selectorBuilder.Len() == 0 {
				err := NewParserError(t, "empty selector in style rule")
				p.recordDiagnostic(p.newDiagnosticError(err, t.Span, rule))
			}
			rule.Selectors = append(rule.Selectors, selectorBuilder.String())
			selectorBuilder.Reset()
		} else {
			selectorBuilder.WriteString(t.Literal)
		}
	}

	lastTok := toks[len(toks)-1]
	if selectorBuilder.Len() == 0 {
		err := NewParserError(lastTok, "empty selector in style rule")
		p.recordDiagnostic(p.newDiagnosticError(err, lastTok.Span, rule))
	}
	rule.Selectors = append(rule.Selectors, selectorBuilder.String())
	selectorBuilder.Reset()
}

func (p *Parser) parseStyleRule(startTok tokenizer.Token) (*ast.StyleRule, error) {
	currentRule := &ast.StyleRule{}

	p.setSelectorsFor(currentRule, startTok)

	lbraceTok, ok := p.stream.TryConsumeType(tokenizer.LBRACE)
	if !ok {
		nextTok := p.stream.PeekTokenAt(0)
		err := NewParserError(nextTok, "Expected opening brace after style rule")
		p.recordDiagnostic(p.newDiagnosticError(err, nextTok.Span, currentRule))
		return currentRule, nil
	}
	currentRule.LBrace = lbraceTok.Span.Start

	err := ast.NewUnclosedBlockError("style rule", "}")

	for {
		if p.isStyleTagEnd() || p.isEndMarker() || p.stream.AssertType(tokenizer.EOF) {
			p.recordDiagnostic(p.newDiagnosticError(err, lbraceTok.Span, currentRule))
			break
		}
		if p.stream.AssertType(tokenizer.RBRACE) {
			break
		}

		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.NEWLINE, tokenizer.SEMICOLON:
			continue
		case tokenizer.EXCLAMATION:
			dir, err := p.parseDirective(tok)
			if err != nil {
				return nil, err
			}
			currentRule.Statements = append(currentRule.Statements, dir)
			continue
		}

		if p.hasLBraceOnLine() {
			nestedRule, _ := p.parseStyleRule(tok)
			currentRule.Statements = append(currentRule.Statements, nestedRule)
			continue
		}

		// By elimination: it is a style declaration
		if tok.Type != tokenizer.IDENTIFIER {
			err := NewParserError(tok, "Expected style property name or rule selector")
			p.recordDiagnostic(p.newDiagnosticError(err, tok.Span, currentRule))
			p.syncToLineEnd()
			continue
		}

		declMark := p.Mark(tok)
		p.stream.TryConsumeType(tokenizer.COLON)
		leadingTrivia := p.dumpTrivia()
		toks := p.stream.ConsumeUntilType(tokenizer.SEMICOLON, tokenizer.NEWLINE)
		if len(toks) == 0 {
			nextTok := p.stream.PeekTokenAt(0)
			err := NewParserError(nextTok, "expected value after style declaration")
			p.recordDiagnostic(p.newDiagnosticError(err, nextTok.Span, currentRule))
			p.syncToLineEnd()
			continue
		}

		val := p.stream.SliceInputEnclosingTokens(toks...)
		p.stream.TryConsumeType(tokenizer.SEMICOLON)
		p.stream.EmitCommentToks()
		declaration := &ast.StyleDeclaration{
			Property: tok.Literal,
			Value:    val,
		}
		declaration.SetMetadata(p.Span(declMark), leadingTrivia, p.dumpTrivia())
		currentRule.Statements = append(currentRule.Statements, declaration)
	}

	if rbraceTok, ok := p.stream.TryConsumeType(tokenizer.RBRACE); ok {
		currentRule.RBrace = rbraceTok.Span.Start
	}
	return currentRule, nil
}

func (p *Parser) hasLBraceOnLine() bool {
	for i := 0; ; i++ {
		tok := p.stream.PeekTokenAt(i)
		if tok.Type == tokenizer.LBRACE {
			return true
		}
		if tok.Type == tokenizer.NEWLINE || tok.Type == tokenizer.EOF || tok.Type == tokenizer.RBRACE || tok.Type == tokenizer.SEMICOLON {
			return false
		}
	}
}
