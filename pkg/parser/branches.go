package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

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

func (p *Parser) parseDiagDirection(tok tokenizer.Token) error {
	var to string
	switch tok.Literal {
	case "left":
		to = "right"
	case "top":
		to = "bottom"
	}

	expectedSeq := []tokenizer.Token{
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
	return p.skinparam.SetAndDecodeWithContext(ast.SkinparamKey{Stereotype: stereo}, name, value)
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

func (p *Parser) parseEntity(tok tokenizer.Token) (*ast.Entity, error) {
	ent := &ast.Entity{
		Kind: p.mapTokenToEntityKind(tok.Type),
	}

	ent.Identifier = p.stream.Emit().Literal
	if _, hasAlias := p.stream.TryConsumeType(tokenizer.ALIAS); hasAlias {
		if !p.stream.AssertAnyType(tokenizer.STRING, tokenizer.IDENTIFIER) {
			return nil, NewParserError("Expected entity alias after 'as'", p.stream.PeekTokenAt(0).Pos)
		}
		ent.Alias = p.stream.Emit().Literal
	}

	stereo, err := p.stream.TryReadStereotype()
	if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
		return nil, err
	}
	ent.Stereotype = stereo

	if _, ok := p.stream.TryConsumeType(tokenizer.HASH); ok {
		collector := tokenizer.TokenCollector{}
		detach := p.stream.Attach(&collector)
		p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.LBRACE)
		detach()
		ent.Color = p.stream.TokensToString(collector.Tokens())
	}

	// It can be there or not, parser doesn't care
	p.stream.TryConsumeType(tokenizer.NEWLINE)

	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		// No body return entity as is
		return ent, nil
	}

	// We are inside a body
	for {
		member, err := p.parseEntityMember()
		if err != nil {
			return ent, err
		}
		if member == nil {
			break
		}

		ent.Members = append(ent.Members, member)
	}

	return ent, nil
}

func (p *Parser) parseEntityMember() (ast.Member, error) {
	var member ast.Member
	var err error

	// Modifiers and separators precede the switch to not mistake -- separator and '-' for visibility
	if member, err = p.stream.TryReadClassSeparator(); err == nil {
		return member, nil
	} else if err != tokenizer.ErrStartMarkerNotFound {
		// Ignore this error as it might be other syntactic constructs
		// Only other error is 'Unexpected EOF' so we just bubble it up
		return nil, err
	}

	field := ast.Field{}
	if mod, err := p.stream.TryReadModifier(); err == nil {
		// Handle scope modifiers
		field.Modifiers = append(field.Modifiers, mod)
		member, err = p.parseFieldOrMethod(tokenizer.Token{Type: tokenizer.LBRACE}, field)
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
		return nil, nil
	case tokenizer.HYPHEN, tokenizer.TILDE, tokenizer.HASH, tokenizer.PLUS:
		field.Visibility = p.mapTokenToVisibility(tok.Type)
	case tokenizer.IDENTIFIER:
		field.Name = tok.Literal
	default:
		return nil, NewParserError("Unexpected token in entity body", p.stream.PeekTokenAt(0).Pos)
	}
	member, err = p.parseFieldOrMethod(tok, field)
	if err != nil {
		return nil, err
	}
	return member, nil
}

// Can be either a field or a method
// but start with a field as either start the same
//
// Returns the parsed field or method, must be asserted to be a field or method
func (p *Parser) parseFieldOrMethod(entryTok tokenizer.Token, field ast.Field) (ast.Member, error) {
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

	if len(field.Modifiers) > 0 {
		switch field.Modifiers[0] {
		case "method":
			mustBeMethod = true
		case "field":
			mustBeField = true
		}
	}

	if p.mapTokenToVisibility(entryTok.Type) != ast.UnknownVisibility {
		canEncounterVisibility = false
	}

	entry := make([]tokenizer.Token, 0, 10)
	if entryTok.Type == tokenizer.IDENTIFIER {
		entry = append(entry, entryTok)
	}

outer:
	for {
		mod, err := p.stream.TryReadModifier()
		if err == nil {
			switch mod {
			case "method":
				mustBeMethod = true
			case "field":
				mustBeField = true
			}
			continue
		}

		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.EOF, tokenizer.NEWLINE:
			break outer
		case tokenizer.HASH, tokenizer.PLUS, tokenizer.HYPHEN, tokenizer.TILDE:
			if canEncounterVisibility {
				entry = append(entry, tok)
				continue
			}
			field.Visibility = p.mapTokenToVisibility(tok.Type)
		case tokenizer.LPAREN:
			containsLParen = true
			fallthrough
		default:
			entry = append(entry, tok)
		}
	}

	if mustBeField && mustBeMethod {
		return nil, NewParserError(
			"Cannot be field and method at the same time",
			entryTok.Pos,
		)
	}

	isMethod := mustBeMethod || (!mustBeField && containsLParen)

	var err error
	if isMethod {
		meth := ast.Method{
			Visibility: field.Visibility,
			Modifiers:  field.Modifiers,
		}

		meth.Name,
			meth.ReturnType,
			meth.Parameters,
			err = p.dialect.ParseMethod(entry)
		if err != nil {
			return nil, err
		}
		return meth, err
	} else {
		var fieldDef ast.Parameter
		fieldDef, err = p.dialect.ParseField(entry)
		if err != nil {
			return nil, err
		}
		field.Name = fieldDef.Name
		field.Type = fieldDef.Type
		return field, err
	}
}

func (p *Parser) parseContainer(mod tokenizer.Token) error {
	panic("unimplemented")
}

func (p *Parser) parseNote(tok tokenizer.Token) (ast.Note, error) {
	panic("unimplemented")
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
			if stereo != "" {
				newKey.Stereotype = stereo
			}
			if err := p.parseSkinparamBlock(newKey); err != nil {
				return err
			}
			continue
		}

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
	return nil
}

func (p *Parser) parseTitle() error {
	var titleEndSequence string
	isBlockTitle := p.stream.AssertType(tokenizer.NEWLINE)
	if !isBlockTitle {
		// We are in luck and ints single line title
		titleEndSequence = "\n"
	} else {
		p.stream.MustConsumeType(tokenizer.NEWLINE)
		titleEndSequence = "end title"
	}
	p.stream.SetRawMode(titleEndSequence)

	titleStrTok := p.stream.MustConsumeType(tokenizer.STRING)
	p.ast.Title = titleStrTok.Literal

	if isBlockTitle && !p.stream.AssertSeqTypes(tokenizer.END_BLOCK, tokenizer.TITLE) {
		return NewParserError("Expected title block to end with \"end title\"", p.stream.PeekTokenAt(0).Pos)
	}
	return nil
}
