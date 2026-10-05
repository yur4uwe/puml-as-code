package parser

import (
	"errors"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseSkinparamRoot(tok tokenizer.Token) (ast.Statement, error) {
	return p.parseSkinparamStatement(tok, true)
}

func (p *Parser) parseSkinparamStatement(tok tokenizer.Token, isRoot bool) (ast.Statement, error) {
	mark := p.Mark(tok)
	leadingTrivia := p.stream.DumpCollectedTrivia()

	var keyword string
	var name string

	if isRoot {
		keyword = tok.Literal

		// Check for anonymous block: `skinparam { ... }`
		if !p.stream.AssertType(tokenizer.LBRACE) {
			nameTok := p.stream.Emit()
			if nameTok.Type == tokenizer.NEWLINE || nameTok.Type == tokenizer.EOF {
				return nil, NewParserError("Expected target or parameter after skinparam", nameTok)
			}
			name = nameTok.Literal
		}
	} else {
		name = tok.Literal
	}

	stereo, err := p.tryReadStereotype()
	if errors.Is(err, tokenizer.ErrUnexpectedEOF) {
		// EOF will be the next token
		return nil, WrapParserError(err, p.stream.PeekTokenAt(0))
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); ok {
		p.stream.EmitCommentToks()
		block := ast.SkinparamBlock{
			Keyword:    keyword,
			Name:       name,
			Stereotype: stereo,
			BaseNode: ast.BaseNode{
				LeadingTrivia:  leadingTrivia,
				TrailingTrivia: p.stream.DumpCollectedTrivia(),
			},
		}
		err := p.parseSkinparamBlockMembers(&block)
		if err != nil {
			return nil, err
		}
		block.NodeSpan = p.Span(mark)
		p.stream.EmitCommentToks()
		p.stream.TryConsumeType(tokenizer.NEWLINE)
		block.TrailingTrivia = append(block.TrailingTrivia, p.stream.DumpCollectedTrivia()...)
		return block, nil
	}

	// read contiguous tokens as a key
	prevTokEnd := p.stream.LastSemanticToken().Span.End.Offset
	for {
		nextTok := p.stream.PeekTokenAt(0)
		if nextTok.Span.Start.Offset != prevTokEnd ||
			nextTok.Type == tokenizer.NEWLINE ||
			nextTok.Type == tokenizer.EOF ||
			nextTok.Type == tokenizer.SEMICOLON {
			break
		}

		switch nextTok.Type {
		case tokenizer.DOT, tokenizer.UNDERSCORE, tokenizer.IDENTIFIER:
			// It's a key.
			nameContinues := p.stream.Emit()
			name += nameContinues.Literal
			prevTokEnd = nextTok.Span.End.Offset
		default:
			return nil, NewParserError("Unexpected token in skinparam block", nextTok)
		}
	}

	toks := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.SEMICOLON)
	value := p.stream.SliceInputEnclosingTokens(toks...)
	span := p.Span(mark)
	p.stream.TryConsumeType(tokenizer.SEMICOLON)
	p.stream.EmitCommentToks()
	p.stream.TryConsumeType(tokenizer.NEWLINE)
	return ast.SkinparamSetting{
		Keyword:    keyword,
		Name:       name,
		Value:      value,
		Stereotype: stereo,
		BaseNode: ast.BaseNode{
			NodeSpan:       span,
			LeadingTrivia:  leadingTrivia,
			TrailingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}, nil
}

func (p *Parser) parseSkinparamBlockMembers(block *ast.SkinparamBlock) error {
	for tok := p.stream.Emit(); tok.Type != tokenizer.RBRACE; tok = p.stream.Emit() {
		var member ast.Statement
		var err error
		switch tok.Type {
		case tokenizer.NEWLINE:
			continue
		case tokenizer.EOF:
			return NewParserError("Unexpected EOF in skinparam block", tok)
		case tokenizer.SEMICOLON:
			continue
		case tokenizer.EXCLAMATION:
			member, err = p.parseDirective(tok)
		default:
			member, err = p.parseSkinparamStatement(tok, false)
		}
		if err != nil {
			return err
		}
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

func (p *Parser) parseStyleBlock(startTok tokenizer.Token) (ast.Statement, error) {
	if !p.stream.AssertSeq(
		[]tokenizer.Token{
			amb(tokenizer.IDENTIFIER, "style"),
			unamb(tokenizer.RANGLE),
		},
	) {
		return nil, NewParserError("Expected <style> opening tag", startTok)
	}
	p.stream.Emit() // consume 'style'
	p.stream.Emit() // consume '>'

	m := p.Mark(startTok)

	styleBlock := ast.StyleBlock{
		BaseNode: ast.BaseNode{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}

	for !p.isStyleTagEnd() {
		var stmnt ast.Statement
		var err error
		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.NEWLINE, tokenizer.SEMICOLON:
			continue
		case tokenizer.EOF:
			return nil, NewParserError("Unexpected EOF in style block", tok)
		case tokenizer.EXCLAMATION:
			stmnt, err = p.parseDirective(tok)
		default:
			// Must be a StyleRule (e.g. "root { ... }" or "classDiagram { ... }")
			stmnt, err = p.parseStyleRule(tok)
		}
		if err != nil {
			return nil, err
		}
		styleBlock.Statements = append(styleBlock.Statements, stmnt)
	}

	// Consume '</style>'
	for range 4 {
		p.stream.Emit()
	}

	styleBlock.NodeSpan = p.Span(m)
	p.stream.EmitCommentToks()
	styleBlock.TrailingTrivia = p.stream.DumpCollectedTrivia()
	return styleBlock, nil
}

func (p *Parser) getSelectors(startTok tokenizer.Token) ([]string, error) {
	// LBRACE is canonocal syntax but leave newline for error recovery
	toks := p.stream.ConsumeUntilType(tokenizer.LBRACE, tokenizer.NEWLINE)
	toks = append([]tokenizer.Token{startTok}, toks...)
	headerText := p.stream.SliceInputEnclosingTokens(toks...)

	var selectors []string
	for _, part := range strings.Split(headerText, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, NewParserError("Empty selector in style rule", toks[0])
		}
		selectors = append(selectors, trimmed)
	}
	return selectors, nil
}

func (p *Parser) parseStyleRule(startTok tokenizer.Token) (ast.StyleRule, error) {
	m := p.Mark(startTok)
	currentRule := ast.StyleRule{
		BaseNode: ast.BaseNode{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}

	var err error
	currentRule.Selectors, err = p.getSelectors(startTok)
	if err != nil {
		return currentRule, err
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		return currentRule, NewParserError("Expected opening brace after style rule", p.stream.PeekTokenAt(0))
	}

	for !p.isStyleTagEnd() && !p.stream.AssertType(tokenizer.RBRACE) && !p.stream.AssertType(tokenizer.EOF) {

		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.NEWLINE, tokenizer.SEMICOLON:
			continue
		case tokenizer.EXCLAMATION:
			dir, err := p.parseDirective(tok)
			if err != nil {
				return currentRule, err
			}
			currentRule.Statements = append(currentRule.Statements, dir)
			continue
		}

		if p.hasLBraceOnLine() {
			nestedRule, err := p.parseStyleRule(tok)
			if err != nil {
				return currentRule, err
			}
			currentRule.Statements = append(currentRule.Statements, nestedRule)
			continue
		}

		// By elimination: it is a style declaration
		if tok.Type != tokenizer.IDENTIFIER {
			return currentRule, NewParserError("Expected style property name or rule selector", tok)
		}

		p.stream.TryConsumeType(tokenizer.COLON)
		leadingTrivia := p.stream.DumpCollectedTrivia()
		toks := p.stream.ConsumeUntilType(tokenizer.SEMICOLON, tokenizer.NEWLINE)
		if len(toks) == 0 {
			return currentRule, NewParserError("Expected value after style declaration", p.stream.PeekRawTokenAt(0))
		}

		val := p.stream.SliceInputEnclosingTokens(toks...)
		semicolonTok, hasSemicolon := p.stream.TryConsumeType(tokenizer.SEMICOLON)
		var span tokenizer.SourceSpan
		if hasSemicolon {
			span = tokenizer.SpanEnclosing(tok, semicolonTok)
		} else {
			span = tokenizer.SpanEnclosing(tok, toks[len(toks)-1])
		}
		p.stream.EmitCommentToks()
		declaration := ast.StyleDeclaration{
			BaseNode: ast.BaseNode{
				NodeSpan:       span,
				LeadingTrivia:  leadingTrivia,
				TrailingTrivia: p.stream.DumpCollectedTrivia(),
			},
			Property: tok.Literal,
			Value:    val,
		}
		currentRule.Statements = append(currentRule.Statements, declaration)
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.RBRACE); !ok {
		return currentRule, NewParserError("Expected closing brace '}' after style rule", p.stream.PeekTokenAt(0))
	}

	currentRule.NodeSpan = p.Span(m)
	p.stream.EmitCommentToks()
	currentRule.TrailingTrivia = p.stream.DumpCollectedTrivia()
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
