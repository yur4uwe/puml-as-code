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

func (p *Parser) parseSkinparamBlock(prefixKey ast.SkinparamKey) error {
	// Next should be an opening brace
	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		return NewParserError("Expected opening brace after skinparam target", p.stream.PeekTokenAt(0).Pos)
	}

	for {
		// Consume leading newlines
		for {
			if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
				break
			}
		}

		tok := p.stream.Emit()
		if tok.Type == tokenizer.RBRACE {
			break
		}
		if tok.Type == tokenizer.EOF {
			return NewParserError("Unexpected EOF in skinparam block", tok.Pos)
		}

		// Read target or param
		name := tok.Literal
		stereo, _ := p.stream.TryReadStereotype()

		if p.stream.AssertType(tokenizer.LBRACE) {
			// Recursive block
			newKey := prefixKey
			if newKey.SubTarget == "" {
				newKey.SubTarget = name
			} else {
				return NewParserError("Skinparam nesting too deep", tok.Pos)
			}
			newKey.Stereotype = stereo
			if err := p.parseSkinparamBlock(newKey); err != nil {
				return err
			}
		} else {
			// Inline value
			value := p.stream.ReadUntilNewline()
			newKey := prefixKey
			if stereo != "" {
				newKey.Stereotype = stereo
			}
			if err := p.skinparam.SetAndDecodeWithContext(newKey, name, value); err != nil {
				return NewParserError(fmt.Sprintf("Failed to set skinparam value: %s", err.Error()), tok.Pos)
			}
		}
	}
	return nil
}

func (p *Parser) parseStyles(tok tokenizer.Token) error {
	if tok.Type != tokenizer.SKINPARAM {
		// <style> token sequence for now simply read it as a string
		// and ignore it
		styles, err := p.stream.ReadBlock(unamb(tokenizer.LANGLE), unamb(tokenizer.SLASH), amb(tokenizer.IDENTIFIER, "style"), unamb(tokenizer.RANGLE))
		if err != nil {
			return NewParserError("Expected style block to end", p.stream.PeekTokenAt(0).Pos)
		}
		p.styles = append(p.styles, styles)
		return nil
	}

	// Consume skinparam keyword (already done if tok.Type == SKINPARAM)

	// Peek paramTok to see if it's a target or a block
	paramTok := p.stream.Emit()
	if paramTok.Type == tokenizer.NEWLINE || paramTok.Type == tokenizer.EOF {
		return NewParserError("Expected target or parameter after skinparam", paramTok.Pos)
	}

	name := paramTok.Literal
	stereo, _ := p.stream.TryReadStereotype()

	if p.stream.AssertType(tokenizer.LBRACE) {
		// skinparam target { ... }
		return p.parseSkinparamBlock(ast.SkinparamKey{MainTarget: name, Stereotype: stereo})
	}

	// skinparam combinedName value
	value := p.stream.ReadUntilNewline()
	return p.skinparam.SetAndDecode(name, stereo, value)
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

func toInteger(f float64) (int, bool) {
	i := int(f)
	if float64(i) == f {
		return i, true
	}
	return 0, false
}

func (p *Parser) parseScale() error {
	cmd := ast.ScaleCommand{}
	if _, ok := p.stream.TryConsume(tokenizer.Token{Type: tokenizer.IDENTIFIER, Literal: "max"}); ok {
		cmd.IsMax = true
	}

	tok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
	if !ok {
		// Try to handle 200x300 which is tokenized as an IDENTIFIER
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

	tok = p.stream.Emit()
	switch tok.Type {
	case tokenizer.IDENTIFIER:
		switch tok.Literal {
		case "width":
			cmd.Width, ok = toInteger(val)
			if !ok {
				return NewParserError("Expected width to be an integer", tok.Pos)
			}
		case "height":
			cmd.Height, ok = toInteger(val)
			if !ok {
				return NewParserError("Expected height to be an integer", tok.Pos)
			}
		case "x":
			cmd.Width, ok = toInteger(val)
			if !ok {
				return NewParserError("Expected width to be an integer", tok.Pos)
			}
			heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
			if !ok {
				return NewParserError("Expected height after 'x'", p.stream.PeekTokenAt(0).Pos)
			}
			h, err := strconv.Atoi(heightTok.Literal)
			if err != nil {
				return NewParserError("Invalid height", heightTok.Pos)
			}
			cmd.Height = h
		default:
			return NewParserError(fmt.Sprintf("Unexpected identifier: %s", tok.Literal), tok.Pos)
		}
	case tokenizer.ASTERISK:
		cmd.Width, ok = toInteger(val)
		if !ok {
			return NewParserError("Expected width to be an integer", tok.Pos)
		}
		heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return NewParserError("Expected height after '*'", p.stream.PeekTokenAt(0).Pos)
		}
		h, err := strconv.Atoi(heightTok.Literal)
		if err != nil {
			return NewParserError("Invalid height", heightTok.Pos)
		}
		cmd.Height = h
	case tokenizer.SLASH:
		numer, ok := toInteger(val)
		if !ok {
			return NewParserError("Expected numerator to be an integer", tok.Pos)
		}
		denomTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return NewParserError("Expected denominator after '/'", p.stream.PeekTokenAt(0).Pos)
		}
		denom, err := strconv.Atoi(denomTok.Literal)
		if err != nil {
			return NewParserError("Invalid denominator", denomTok.Pos)
		}
		if denom == 0 {
			return NewParserError("Denominator cannot be zero", denomTok.Pos)
		}
		cmd.Scale = float64(numer) / float64(denom)
	case tokenizer.NEWLINE, tokenizer.EOF:
		// EOF handling is a special case for a test case
		cmd.Scale = val
	default:
		return NewParserError(fmt.Sprintf("Unexpected token: %s(%s)", tok.Type.String(), tok.Literal), tok.Pos)
	}

	p.ast.Statements = append(p.ast.Statements, cmd)
	return nil
}
