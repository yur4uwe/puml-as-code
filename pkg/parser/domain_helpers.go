package parser

import (
	"fmt"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) tryReadModifier() (string, error) {
	start := []tokenizer.Token{{Type: tokenizer.LBRACE}}
	end := []tokenizer.Token{{Type: tokenizer.RBRACE}}
	return p.stream.ReadBetween(start, end)
}

func (p *Parser) tryReadStereotype() (string, error) {
	start := []tokenizer.Token{{Type: tokenizer.LANGLE}, {Type: tokenizer.LANGLE}}
	end := []tokenizer.Token{{Type: tokenizer.RANGLE}, {Type: tokenizer.RANGLE}}
	return p.stream.ReadBetween(start, end)
}

func (p *Parser) tryReadGeneric() (string, error) {
	start := []tokenizer.Token{{Type: tokenizer.LANGLE}}
	end := []tokenizer.Token{{Type: tokenizer.RANGLE}}
	return p.stream.ReadBetween(start, end)
}

func (p *Parser) tryReadClassSeparator() (ast.ClassSeparator, error) {
	var start, end []tokenizer.Token
	var sepChar rune
	switch p.stream.PeekTokenAt(0).Type {
	case tokenizer.DASH:
		start = []tokenizer.Token{{Type: tokenizer.DASH}, {Type: tokenizer.DASH}}
		end = []tokenizer.Token{{Type: tokenizer.DASH}, {Type: tokenizer.DASH}}
		sepChar = '-'
	case tokenizer.DOT:
		start = []tokenizer.Token{{Type: tokenizer.DOT}, {Type: tokenizer.DOT}}
		end = []tokenizer.Token{{Type: tokenizer.DOT}, {Type: tokenizer.DOT}}
		sepChar = '.'
	case tokenizer.EQUALS:
		start = []tokenizer.Token{{Type: tokenizer.EQUALS}, {Type: tokenizer.EQUALS}}
		end = []tokenizer.Token{{Type: tokenizer.EQUALS}, {Type: tokenizer.EQUALS}}
		sepChar = '='
	case tokenizer.UNDERSCORE:
		start = []tokenizer.Token{{Type: tokenizer.UNDERSCORE}, {Type: tokenizer.UNDERSCORE}}
		end = []tokenizer.Token{{Type: tokenizer.UNDERSCORE}, {Type: tokenizer.UNDERSCORE}}
		sepChar = '_'
	default:
		return ast.ClassSeparator{}, fmt.Errorf("unexpected class separator")
	}

	if p.stream.AssertSeq(append(start, tokenizer.Token{Type: tokenizer.NEWLINE})) {
		return ast.ClassSeparator{
			Type: sepChar,
		}, nil
	}

	str, err := p.stream.ReadBetween(start, end)
	if err != nil {
		return ast.ClassSeparator{}, err
	}
	return ast.ClassSeparator{
		Label: str,
		Type:  sepChar,
	}, nil
}

func (p *Parser) tryReadTag() (string, error) {
	if !p.stream.AssertType(tokenizer.DOLLAR) {
		return "", fmt.Errorf("expected dollar sign tag")
	}
	p.stream.Emit() // consume $
	return p.stream.Emit().Literal, nil
}

func (p *Parser) isDiagramBound() bool {
	if !p.stream.AssertType(tokenizer.AT) {
		return false
	}
	tok := p.stream.PeekTokenAt(1)
	if tok.Type != tokenizer.IDENTIFIER {
		return false
	}
	return strings.HasPrefix(tok.Literal, "start") || strings.HasPrefix(tok.Literal, "end")
}

func (p *Parser) readDiagramBounds() (ast.DiagramBound, error) {
	atTok, consumed := p.stream.TryConsumeType(tokenizer.AT)
	if !consumed {
		return ast.DiagramBound{}, fmt.Errorf("expected @ at diagram bounds start, got %s", atTok.Type)
	}

	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return ast.DiagramBound{}, fmt.Errorf("expected identifier at diagram bounds, got %s", tok.Type)
	}

	if !strings.HasPrefix(tok.Literal, "start") &&
		!strings.HasPrefix(tok.Literal, "end") {
		return ast.DiagramBound{}, fmt.Errorf("invalid bounding marker for diagram, expected something that starts with 'start' or 'end'")
	}

	diag := ast.DiagramBound{
		Opts:          make(map[string]string),
		LeadingTrivia: atTok.LeadingTrivia,
	}

	if typ, found := strings.CutPrefix(tok.Literal, "start"); found {
		diag.IsStart = true
		diag.Type = typ
	} else if typ, found := strings.CutPrefix(tok.Literal, "end"); found {
		diag.IsStart = false
		diag.Type = typ
		return diag, nil
	}

	if !p.stream.AssertType(tokenizer.LPAREN) && !p.stream.AssertType(tokenizer.LBRACE) {
		diag.Name = p.stream.ReadRawUntilNewline()
		return diag, nil
	}

	readKvp := func() error {
		keyTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
		if !ok {
			return fmt.Errorf("expected a key")
		}
		if _, ok := p.stream.TryConsumeType(tokenizer.EQUALS); !ok {
			return fmt.Errorf("expected = after key")
		}
		valTok := p.stream.Emit()
		if keyTok.Literal == "id" {
			diag.ID = valTok.Literal
		} else if _, ok := diag.Opts[keyTok.Literal]; ok {
			return fmt.Errorf("duplicate key in diagram bounds: %s", keyTok.Literal)
		} else {
			diag.Opts[keyTok.Literal] = valTok.Literal
		}
		return nil
	}

	if _, consumed := p.stream.TryConsumeType(tokenizer.LPAREN); consumed {
		for !p.stream.AssertType(tokenizer.RPAREN) && !p.stream.AssertType(tokenizer.EOF) && !p.stream.AssertType(tokenizer.NEWLINE) {
			if err := readKvp(); err != nil {
				return diag, err
			}
			if _, ok := p.stream.TryConsumeType(tokenizer.COMMA); !ok {
				break
			}
		}
		if _, ok := p.stream.TryConsumeType(tokenizer.RPAREN); !ok {
			return diag, fmt.Errorf("expected closing parenthesis in diagram bounds")
		}
	}

	if !p.stream.AssertType(tokenizer.LBRACE) {
		diag.Name = p.stream.ReadRawUntilNewline()
		return diag, nil
	}

	if _, consumed := p.stream.TryConsumeType(tokenizer.LBRACE); consumed {
		tokens := p.stream.ConsumeUntilType(tokenizer.EOF, tokenizer.NEWLINE, tokenizer.COMMA, tokenizer.RBRACE)
		if p.stream.AssertType(tokenizer.EOF) || p.stream.AssertType(tokenizer.NEWLINE) {
			return diag, fmt.Errorf("unexpected EOF or newline")
		}
		diag.Name = p.stream.TokensToString(tokens)

		if p.stream.PeekTokenAt(2).Type != tokenizer.EQUALS {
			p.stream.TryConsumeType(tokenizer.COMMA)
			captionTokens := p.stream.ConsumeUntilType(tokenizer.EOF, tokenizer.NEWLINE, tokenizer.COMMA, tokenizer.RBRACE)
			if p.stream.AssertType(tokenizer.EOF) || p.stream.AssertType(tokenizer.NEWLINE) {
				return diag, fmt.Errorf("unexpected EOF or newline")
			}
			diag.Opts["caption"] = p.stream.TokensToString(captionTokens)
		}

		for p.stream.AssertType(tokenizer.COMMA) {
			p.stream.TryConsumeType(tokenizer.COMMA)
			if err := readKvp(); err != nil {
				return diag, err
			}
		}

		if _, ok := p.stream.TryConsumeType(tokenizer.RBRACE); !ok {
			return diag, fmt.Errorf("expected } at end of diagram bounds options")
		}
	}

	if p.stream.AssertType(tokenizer.LPAREN) {
		return diag, fmt.Errorf("possibly incorrect order of tools and id")
	}

	if !p.stream.AssertType(tokenizer.NEWLINE) {
		return diag, fmt.Errorf("expected newline after diagram bounds")
	}

	return diag, nil
}
