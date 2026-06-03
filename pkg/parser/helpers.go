package parser

import (
	"fmt"
	"strconv"
	"strings"
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
	if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
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

func (p *Parser) parseSkinparamBlock(tok tokenizer.Token) error {
	// This function should be passed the keyword that names the
	// target of the skinparam block
	target := tok.Literal
	// Next should be an opening brace
	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		return NewParserError("Expected opening brace after skinparam keyword", tok.Pos)
	}
	// We parse lines until we encounter a closing brace
	for {
		tok := p.stream.Emit()
		if tok.Type != tokenizer.NEWLINE {
			return NewParserError("Expected to have only one skinparam style per line", tok.Pos)
		}
		tok = p.stream.Emit()
		if tok.Type == tokenizer.RBRACE {
			break
		}
		styleName := tok.Literal
		stereoName, _ := p.stream.TryReadStereotype() // We can igrnore error as the stereotype is optional

		// Here we can encounter hex for colors, identifier for colors, or even values for some styles
		value := p.stream.ReadUntilNewline()
		if err := p.skinparam.Set(target, styleName, stereoName, value); err != nil {
			return NewParserError(fmt.Sprintf("Failed to set skinparam value: %s", err.Error()), tok.Pos)
		}
	}
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
	// Otherwise its a skinparam keyword
	if tok.Type != tokenizer.SKINPARAM {
		return NewParserError("Expected skinparam keyword", tok.Pos)
	}

	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		// We must have encountered a keyword and can assume skinparam block like:
		// skinparam class {}
		return p.parseSkinparamBlock(tok)
	}

	target := tok.Literal // of the form xxxParamX
	// I dont really know whether inline skinparam can have a stereotype so i won't handle it
	value := p.stream.ReadUntilNewline()

	return p.skinparam.SetAndDecode(target, "", value)
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
		if _, ok := p.stream.TryConsume(token); !ok {
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

func (p *Parser) parseScale() error {
	cmd := ast.ScaleCommand{}
	if _, ok := p.stream.TryConsume(tokenizer.Token{Type: tokenizer.IDENTIFIER, Literal: "max"}); ok {
		cmd.IsMax = true
	}

	tok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
	if !ok {
		// Try to handle 200x300 which might be tokenized as an IDENTIFIER
		if tok, ok = p.stream.TryConsumeType(tokenizer.IDENTIFIER); !ok {
			return NewParserError("Expected number after scale", p.stream.PeekTokenAt(0).Pos)
		}
		if !strings.Contains(tok.Literal, "x") {
			return NewParserError("Expected number after scale", tok.Pos)
		}
		parts := strings.Split(tok.Literal, "x")
		if len(parts) == 2 {
			w, errW := strconv.Atoi(parts[0])
			h, errH := strconv.Atoi(parts[1])
			if errW == nil && errH == nil {
				cmd.Width = w
				cmd.Height = h
				p.ast.Statements = append(p.ast.Statements, cmd)
				return nil
			}
		}
	}

	val, err := strconv.ParseFloat(tok.Literal, 64)
	if err != nil {
		return NewParserError(fmt.Sprintf("Invalid number: %s", err.Error()), tok.Pos)
	}

	if p.stream.AssertType(tokenizer.IDENTIFIER) {
		next := p.stream.PeekTokenAt(0)
		switch next.Literal {
		case "width":
			p.stream.Emit()
			cmd.Width = int(val)
		case "height":
			p.stream.Emit()
			cmd.Height = int(val)
		default:
			cmd.Scale = val
		}
	} else if _, ok := p.stream.TryConsumeType(tokenizer.ASTERISK); ok {
		cmd.Width = int(val)
		heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return NewParserError("Expected height after '*'", p.stream.PeekTokenAt(0).Pos)
		}
		h, err := strconv.Atoi(heightTok.Literal)
		if err != nil {
			return NewParserError("Invalid height", heightTok.Pos)
		}
		cmd.Height = h
	} else {
		cmd.Scale = val
	}

	p.ast.Statements = append(p.ast.Statements, cmd)
	return nil
}
