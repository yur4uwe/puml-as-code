package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) isStyleTagEnd() bool {
	return p.stream.AssertSeq([]tokenizer.Token{
		{Type: tokenizer.LANGLE},
		{Type: tokenizer.SLASH},
		{Type: tokenizer.IDENTIFIER, Literal: "style"},
		{Type: tokenizer.RANGLE},
	})
}

func (p *Parser) parseStyleBlock() *ast.StyleBlock {
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
			p.ast.RecordError(
				unclBlkErr,
				blockMarkerSpan,
				styleBlock,
			)
			break
		} else if tok.Type == tokenizer.NEWLINE || tok.Type == tokenizer.SEMICOLON {
			continue
		}

		ruleMark := p.Mark(tok)
		leadingTrivia := p.dumpTrivia()

		var stmnt ast.Statement
		switch tok.Type {
		case tokenizer.EXCLAMATION:
			stmnt = p.parseDirective(tok)
		default:
			// Must be a StyleRule (e.g. "root { ... }" or "classDiagram { ... }")
			stmnt = p.parseStyleRule(tok)
		}
		if stmnt != nil {
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

	return styleBlock
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
				p.ast.RecordError(err, t.Span, rule)
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
		p.ast.RecordError(err, lastTok.Span, rule)
	}
	rule.Selectors = append(rule.Selectors, selectorBuilder.String())
	selectorBuilder.Reset()
}

func (p *Parser) parseStyleRule(startTok tokenizer.Token) *ast.StyleRule {
	currentRule := &ast.StyleRule{}

	p.setSelectorsFor(currentRule, startTok)

	lbraceTok, ok := p.stream.TryConsumeType(tokenizer.LBRACE)
	if !ok {
		nextTok := p.stream.PeekTokenAt(0)
		err := NewParserError(nextTok, "Expected opening brace after style rule")
		p.ast.RecordError(err, nextTok.Span, currentRule)
		return currentRule
	}
	currentRule.LBrace = lbraceTok.Span.Start

	err := ast.NewUnclosedBlockError("style rule", "}")

	for {
		if p.isStyleTagEnd() || p.isEndMarker() || p.stream.AssertType(tokenizer.EOF) {
			p.ast.RecordError(err, lbraceTok.Span, currentRule)
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
			dir := p.parseDirective(tok)
			if dir != nil {
				currentRule.Statements = append(currentRule.Statements, dir)
			}
			continue
		}

		if p.hasLBraceOnLine() {
			nestedRule := p.parseStyleRule(tok)
			currentRule.Statements = append(currentRule.Statements, nestedRule)
			continue
		}

		// By elimination: it is a style declaration
		if tok.Type != tokenizer.IDENTIFIER {
			err := NewParserError(tok, "Expected style property name or rule selector")
			p.ast.RecordError(err, tok.Span, currentRule)
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
			p.ast.RecordError(err, nextTok.Span, currentRule)
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
	return currentRule
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
