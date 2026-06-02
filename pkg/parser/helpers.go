package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func unamb(tok tokenizer.TokenType) tokenizer.Token {
	return tokenizer.Token{Type: tok}
}

func amb(tok tokenizer.TokenType, literal string) tokenizer.Token {
	return tokenizer.Token{Type: tok, Literal: literal}
}

func (p *Parser) parseCommand(tok tokenizer.Token) (any, error) {
	switch tok.Type {
	case tokenizer.SET_CMD:
	case tokenizer.HIDE, tokenizer.SHOW, tokenizer.REMOVE, tokenizer.RESTORE:
	case tokenizer.SCALE:
	case tokenizer.EXCLAMATION:
	}
	return nil, nil
}

func (p *Parser) parseTitle() error {
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		// We are in luck and ints single line title
		p.ast.Title = p.stream.ReadRawUntilNewline()
		return nil
	}
	if _, ok := p.stream.ConsumeType(tokenizer.NEWLINE); !ok {
		return NewParserError("Expected newline after multiline title", p.stream.PeekTokenAt(0).Pos)
	}

	// Otherwise we are inside a multiline title block
	title, err := p.stream.ReadBlock(unamb(tokenizer.END_BLOCK), unamb(tokenizer.TITLE))
	if err != nil {
		return NewParserError("Expected title block to end", p.stream.PeekTokenAt(0).Pos)
	}

	p.ast.Title = title
	return nil
}

func (p *Parser) parseStyles(tok tokenizer.Token) error {
	if tok.Type != tokenizer.LANGLE {
		// <style> token sequence for now simply read it as a string
		// and ignore it
		// I probebly should actually assert internal value of the IDENTIFIER token
		seq := []tokenizer.Token{unamb(tokenizer.LANGLE), unamb(tokenizer.SLASH), amb(tokenizer.IDENTIFIER, "style"), unamb(tokenizer.RANGLE)}
		styles, err := p.stream.ReadBlock(seq...)
		if err != nil {
			return NewParserError("Expected style block to end", p.stream.PeekTokenAt(0).Pos)
		}
		p.styles = append(p.styles, styles)
		return nil
	}
	return nil
}

func (p *Parser) parseDiagDirection(tok tokenizer.Token) error {
	buildExpSeq := func(from string) []tokenizer.Token {
		var to string
		switch from {
		case "left":
			to = "right"
		case "top":
			to = "bottom"
		}
		return []tokenizer.Token{
			{
				Type:    tokenizer.IDENTIFIER,
				Literal: "to",
			},
			{
				Type:    tokenizer.DIRECTION,
				Literal: to,
			},
			{
				Type:    tokenizer.IDENTIFIER,
				Literal: "direction",
			},
		}
	}
	expectedSeq := buildExpSeq(tok.Literal)
	for _, token := range expectedSeq {
		if _, ok := p.stream.Consume(token); !ok {
			return NewParserError("Unexpected diagram direction modifier", token.Pos)
		}
	}
	switch tok.Literal {
	case "left":
		p.ast.Statements = append(p.ast.Statements, ast.DirectionCommand{Direction: ast.LeftToRightDirection})
	case "top":
		p.ast.Statements = append(p.ast.Statements, ast.DirectionCommand{Direction: ast.TopToBottomDirection})
	}
	return nil
}

func (p *Parser) parseVisibilityCommand(tok tokenizer.Token) error {
	cmd := ast.VisibilityCommand{
		Kind: ast.Unknown,
	}
	switch tok.Type {
	case tokenizer.HIDE:
		cmd.Kind = ast.Hide
	case tokenizer.SHOW:
		cmd.Kind = ast.Show
	case tokenizer.REMOVE:
		cmd.Kind = ast.Remove
	case tokenizer.RESTORE:
		cmd.Kind = ast.Restore
	}
	cmd.Target = p.stream.ReadUntilNewline()
	p.ast.Statements = append(p.ast.Statements, cmd)
	return nil
}
