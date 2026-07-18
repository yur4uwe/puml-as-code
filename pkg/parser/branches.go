// Package parser implements a parser for PlantUML files
package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseVisibilityCommand(tok tokenizer.Token) (ast.VisibilityCommand, error) {
	cmd := ast.VisibilityCommand{
		Kind: ast.Unknown,
	}
	switch keyword.Classify(tok.Literal) {
	case keyword.Hide:
		cmd.Kind = ast.Hide
	case keyword.Show:
		cmd.Kind = ast.Show
	case keyword.Remove:
		cmd.Kind = ast.Remove
	case keyword.Restore:
		cmd.Kind = ast.Restore
	}
	cmd.Target = p.stream.ReadUntilNewline()
	return cmd, nil
}

func (p *Parser) parseDiagDirection(tok tokenizer.Token) (ast.DirectionCommand, error) {
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
			Type:    tokenizer.IDENTIFIER,
			Literal: to,
		},
		{
			Type:    tokenizer.IDENTIFIER,
			Literal: "direction",
		},
	}

	for _, token := range expectedSeq {
		if _, ok := p.stream.TryConsume(token); !ok {
			return ast.DirectionCommand{}, NewParserError("Unexpected diagram direction modifier", token.Pos)
		}
	}
	switch tok.Literal {
	case "left":
		return ast.DirectionCommand{Direction: ast.LeftToRightDirection}, nil
	case "top":
		return ast.DirectionCommand{Direction: ast.TopToBottomDirection}, nil
	}
	return ast.DirectionCommand{}, nil
}

func (p *Parser) parseSkinparam() error {
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
	return p.ast.Skinparam.SetAndDecodeWithContext(ast.SkinparamKey{Stereotype: stereo}, name, value)
}

func (p *Parser) parseScale() (ast.ScaleCommand, error) {
	cmd := ast.ScaleCommand{}
	if _, ok := p.stream.TryConsume(tokenizer.Token{Type: tokenizer.IDENTIFIER, Literal: "max"}); ok {
		cmd.IsMax = true
	}

	tok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
	if !ok {
		// Try to handle 200x300 which is tokenized as an IDENTIFIER
		if tok, ok = p.stream.TryConsumeType(tokenizer.IDENTIFIER); !ok {
			return ast.ScaleCommand{}, NewParserError("Expected number after scale", p.stream.PeekTokenAt(0).Pos)
		}
		if !strings.Contains(tok.Literal, "x") {
			return ast.ScaleCommand{}, NewParserError("Expected number after scale", tok.Pos)
		}
		parts := strings.Split(tok.Literal, "x")
		if len(parts) == 2 {
			w, errW := strconv.Atoi(parts[0])
			h, errH := strconv.Atoi(parts[1])
			if errW == nil && errH == nil {
				cmd.Width = w
				cmd.Height = h
				p.ast.Statements = append(p.ast.Statements, cmd)
				return ast.ScaleCommand{}, nil
			}
		}
	}

	val, err := strconv.ParseFloat(tok.Literal, 64)
	if err != nil {
		return ast.ScaleCommand{}, NewParserError(fmt.Sprintf("Invalid number: %s", err.Error()), tok.Pos)
	}

	tok = p.stream.Emit()
	switch tok.Type {
	case tokenizer.IDENTIFIER:
		switch tok.Literal {
		case "width":
			cmd.Width, ok = toInteger(val)
			if !ok {
				return ast.ScaleCommand{}, NewParserError("Expected width to be an integer", tok.Pos)
			}
		case "height":
			cmd.Height, ok = toInteger(val)
			if !ok {
				return ast.ScaleCommand{}, NewParserError("Expected height to be an integer", tok.Pos)
			}
		case "x":
			cmd.Width, ok = toInteger(val)
			if !ok {
				return ast.ScaleCommand{}, NewParserError("Expected width to be an integer", tok.Pos)
			}
			heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
			if !ok {
				return ast.ScaleCommand{}, NewParserError("Expected height after 'x'", p.stream.PeekTokenAt(0).Pos)
			}
			h, err := strconv.Atoi(heightTok.Literal)
			if err != nil {
				return ast.ScaleCommand{}, NewParserError("Invalid height", heightTok.Pos)
			}
			cmd.Height = h
		default:
			return ast.ScaleCommand{}, NewParserError(fmt.Sprintf("Unexpected identifier: %s", tok.Literal), tok.Pos)
		}
	case tokenizer.ASTERISK:
		cmd.Width, ok = toInteger(val)
		if !ok {
			return ast.ScaleCommand{}, NewParserError("Expected width to be an integer", tok.Pos)
		}
		heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return ast.ScaleCommand{}, NewParserError("Expected height after '*'", p.stream.PeekTokenAt(0).Pos)
		}
		h, err := strconv.Atoi(heightTok.Literal)
		if err != nil {
			return ast.ScaleCommand{}, NewParserError("Invalid height", heightTok.Pos)
		}
		cmd.Height = h
	case tokenizer.SLASH:
		numer, ok := toInteger(val)
		if !ok {
			return ast.ScaleCommand{}, NewParserError("Expected numerator to be an integer", tok.Pos)
		}
		denomTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return ast.ScaleCommand{}, NewParserError("Expected denominator after '/'", p.stream.PeekTokenAt(0).Pos)
		}
		denom, err := strconv.Atoi(denomTok.Literal)
		if err != nil {
			return ast.ScaleCommand{}, NewParserError("Invalid denominator", denomTok.Pos)
		}
		if denom == 0 {
			return ast.ScaleCommand{}, NewParserError("Denominator cannot be zero", denomTok.Pos)
		}
		cmd.Scale = float64(numer) / float64(denom)
	case tokenizer.NEWLINE, tokenizer.EOF:
		// EOF handling is a special case for a test case
		cmd.Scale = val
	default:
		return ast.ScaleCommand{}, NewParserError(fmt.Sprintf("Unexpected token: %s(%s)", tok.Type.String(), tok.Literal), tok.Pos)
	}

	return cmd, nil
}

func setAliasAndName(ent *ast.Entity, nameOrAlias tokenizer.Token) error {
	switch nameOrAlias.Type {
	case tokenizer.STRING:
		if ent.Alias != "" {
			return NewParserError("Entity alias already set", nameOrAlias.Pos)
		}
		ent.Alias = nameOrAlias.Literal
	case tokenizer.IDENTIFIER:
		if ent.Identifier != "" {
			return NewParserError("Entity name already set", nameOrAlias.Pos)
		}
		ent.Identifier = nameOrAlias.Literal
	default:
		return NewParserError("Expected token for entity identifier or alias", nameOrAlias.Pos)
	}
	return nil
}

// tok is the kind of an entity (class, interface, struct, enum, etc.)
func (p *Parser) parseEntity(tok tokenizer.Token) (*ast.Entity, error) {
	ent := &ast.Entity{
		Kind: p.mapTokenToEntityKind(tok),
	}

	// Unexpectedly, class and other entity definitions have very strict syntax:
	// class <entity name> as <entity alias> <generics> <stereotype> <styles> <body>

	if err := setAliasAndName(ent, p.stream.Emit()); err != nil {
		return nil, err
	}

	if _, hasAlias := p.stream.TryConsumeKW(keyword.Alias); hasAlias {
		if err := setAliasAndName(ent, p.stream.Emit()); err != nil {
			return nil, err
		}
	}

	if !p.stream.AssertSeq([]tokenizer.Token{{Type: tokenizer.LANGLE}, {Type: tokenizer.LANGLE}}) {
		gen, err := p.stream.TryReadGeneric()
		if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
			return nil, err
		}
		ent.Generic = gen
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
		for {
			if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
				break
			}
		}

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
	} else if err == tokenizer.ErrUnexpectedEOF {
		return nil, err
	}

	vis := ast.UnknownVisibility
	if mod, err := p.stream.TryReadModifier(); err == nil {
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
	default:
		return nil, NewParserError("Unexpected token in entity body", p.stream.PeekTokenAt(0).Pos)
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
	var mods []string

	if mod != nil {
		switch *mod {
		case "method":
			mustBeMethod = true
		case "field":
			mustBeField = true
		}
		mods = append(mods, *mod)
	}

	if vis != ast.UnknownVisibility {
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
			mods = append(mods, mod)
			continue
		} else if err != tokenizer.ErrStartMarkerNotFound {
			return nil, err
		}

		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.EOF:
			return nil, tokenizer.ErrUnexpectedEOF
		case tokenizer.NEWLINE:
			break outer
		case tokenizer.HASH, tokenizer.PLUS, tokenizer.DASH, tokenizer.TILDE:
			if canEncounterVisibility {
				entry = append(entry, tok)
				continue
			}
			vis = p.mapTokenToVisibility(tok.Type)
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

	if isMethod {
		return p.dialect.ParseMethod(entry, vis, mods)
	} else {
		return p.dialect.ParseField(entry, vis, mods)
	}
}

func (p *Parser) parseContainer(tok tokenizer.Token) (ast.Container, error) {
	containerClass := keyword.Classify(tok.Literal)
	container := ast.Container{
		Kind: p.mapKeywordToContainerKind(containerClass),
	}

	// Handle container name, alias, stereotype, color
	// Name rules:
	// - Name and alias must be a string or Identifier
	// - Name and alias CAN be of the same type
	// - if Name and alias are of the same type, alias is on the left
	// and Name is on the right of 'as' keyword

	if containerClass != keyword.Together {
		var err error
		container.Alias, container.Identifier, err = p.parseContainerIdentAndAlias()
		if err != nil {
			return container, err
		}
		container.Stereotype, err = p.stream.TryReadStereotype()
		if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
			return container, err
		}

		container.Color = p.tryParseColor()
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		if _, ok = p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
			return container, NewParserError("Expected container body to end", p.stream.PeekTokenAt(0).Pos)
		}
		// if there is no body we can return, though the uml diagram drawer
		// spits an error if the container is empty.
		// As i understand, error is related to the fact that empty containers
		// are related to other uml diagram types, thus producing a mixed diagram
		// which is an error
		return container, nil
	}

	for tok := p.stream.Emit(); tok.Type != tokenizer.RBRACE; tok = p.stream.Emit() {
		if tok.Type == tokenizer.EOF {
			return container, tokenizer.ErrUnexpectedEOF
		}

		// Only parse statements allowed in containers
		stmt, err := p.parseContainerStatement(tok)
		if err != nil {
			return container, err
		}
		if stmt == nil {
			return container, NewParserError("Expected a statement in a container body", tok.Pos)
		}
		container.Statements = append(container.Statements, stmt)
	}
	return container, nil
}

// parseContainerIdentAndAlias parses the container name and alias
//
// Returns
// - Alias
// - Identifier
// - error
func (p *Parser) parseContainerIdentAndAlias() (string, string, error) {
	var lhs tokenizer.Token
	if p.stream.AssertAnyType(tokenizer.STRING, tokenizer.IDENTIFIER) {
		lhs = p.stream.Emit()
	} else {
		return "", "", NewParserError("Expected container name or alias", p.stream.PeekTokenAt(0).Pos)
	}
	if _, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		return "", lhs.Literal, nil
	}
	var rhs tokenizer.Token
	if p.stream.AssertAnyType(tokenizer.STRING, tokenizer.IDENTIFIER) {
		rhs = p.stream.Emit()
	} else {
		return "", "", NewParserError("Expected container name or alias", p.stream.PeekTokenAt(0).Pos)
	}
	if lhs.Type == rhs.Type {
		return lhs.Literal, rhs.Literal, nil
	}
	switch lhs.Type {
	case tokenizer.STRING:
		return lhs.Literal, rhs.Literal, nil
	case tokenizer.IDENTIFIER:
		return rhs.Literal, lhs.Literal, nil
	default:
		panic("unreachable")
	}
}

func (p *Parser) parseNote(tok tokenizer.Token) (ast.Note, error) {
	// tok is a keyword 'note'
	tok = p.stream.Emit()
	if tok.Type == tokenizer.STRING {
		return p.parseInlineAliasNote(tok)
	}
	class := keyword.Classify(tok.Literal)
	switch class {
	case keyword.Direction:
		return p.parseReltiveNote(tok)
	case keyword.Position:
		return p.parseLinkNote(tok)
	case keyword.Alias:
		return p.parseMultilineAliasNote()
	default:
		return ast.Note{}, WrapParserError(fmt.Errorf("expected direction, string, note position or alias after 'note', got %s", class.String()), tok.Pos)
	}
}

func (p *Parser) parseReltiveNote(dirTok tokenizer.Token) (ast.Note, error) {
	var note ast.Note
	note.Direction = p.mapTokenToDirection(dirTok)
	if relativeTok, ok := p.stream.TryConsumeKW(keyword.Position); ok {
		targetTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
		if !ok {
			return note, NewParserError("Expected identifier for a note target", targetTok.Pos)
		}
		if strings.ToLower(targetTok.Literal) != "link" && relativeTok.Literal == "on" {
			return note, NewParserError("Unexpected identifier for a note link target", targetTok.Pos)
		}
		note.Target = targetTok.Literal
	} else if relativeTok.Type == tokenizer.IDENTIFIER {
		tok := p.stream.Emit()
		return note, NewParserError("Unexpected identifier after direction", tok.Pos)
	}
	p.tryParseColor()
	txt, err := p.parseNoteBody()
	if err != nil {
		return note, err
	}
	note.Text = txt
	return note, nil
}

func (p *Parser) parseInlineAliasNote(stringTok tokenizer.Token) (ast.Note, error) {
	var note ast.Note
	note.Text = stringTok.Literal
	if aliasTok, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		return note, NewParserError("Expected alias keyword after note text", aliasTok.Pos)
	}
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return note, NewParserError("Expected identifier after alias keyword", tok.Pos)
	}
	note.Alias = tok.Literal
	p.tryParseColor()
	return note, nil
}

func (p *Parser) parseMultilineAliasNote() (ast.Note, error) {
	var note ast.Note
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return note, NewParserError("Expected identifier after alias keyword", tok.Pos)
	}
	p.tryParseColor()
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		return note, NewParserError("Expected newline after alias keyword", tok.Pos)
	}
	txt, err := p.parseNoteBody()
	if err != nil {
		return note, err
	}
	note.Text = txt
	return note, nil
}

func (p *Parser) parseLinkNote(onTok tokenizer.Token) (ast.Note, error) {
	var note ast.Note
	if onTok.Literal != "on" {
		return note, NewParserError("Unexpected identifier after 'note'", onTok.Pos)
	}
	if _, ok := p.stream.TryConsume(amb(tokenizer.IDENTIFIER, "link")); !ok {
		return note, NewParserError("Expected 'link' after 'note on'", onTok.Pos)
	}
	txt, err := p.parseNoteBody()
	if err != nil {
		return note, err
	}
	note.Text = txt
	return note, nil
}

func (p *Parser) tryParseColor() string {
	if _, ok := p.stream.TryConsumeType(tokenizer.HASH); !ok {
		return ""
	}
	collector := tokenizer.TokenCollector{}
	detach := p.stream.Attach(&collector)
	p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.COLON)
	detach()
	return p.stream.TokensToString(collector.Tokens())
}

func (p *Parser) parseNoteBody() (string, error) {
	tok := p.stream.Emit()

	var noteEndSequence string
	switch tok.Type {
	case tokenizer.COLON:
		noteEndSequence = "\n"
	case tokenizer.NEWLINE:
		noteEndSequence = "end note"
	default:
		return "", NewParserError("Expected ':' or newline after note definition", tok.Pos)
	}

	p.stream.SetRawMode(noteEndSequence)
	bodyTok := p.stream.MustConsumeType(tokenizer.STRING)
	return bodyTok.Literal, nil
}

func (p *Parser) mapTokenToDirection(tok tokenizer.Token) ast.DirectionKind {
	switch tok.Literal {
	case "left":
		return ast.Left
	case "right":
		return ast.Right
	case "top":
		return ast.Top
	case "bottom":
		return ast.Bottom
	default:
		return ast.UnknownDirectionKind
	}
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
		if err := p.ast.Skinparam.SetAndDecodeWithContext(newKey, name, value); err != nil {
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

	if isBlockTitle && !p.stream.AssertKWSeq(keyword.End, keyword.Title) {
		return NewParserError("Expected title block to end with \"end title\"", p.stream.PeekTokenAt(0).Pos)
	}
	return nil
}

func (p *Parser) parseRelationshipTarget(entityNameTok tokenizer.Token) (string, error) {
	target := entityNameTok.Literal
	if _, ok := p.stream.TryConsumeType(tokenizer.COLON); ok {
		// can be field or method reference
		_, ok := p.stream.TryConsumeType(tokenizer.COLON)
		if !ok {
			return "", NewParserError("Expected '::' to reference a class member", p.stream.PeekTokenAt(0).Pos)
		}

		tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
		if !ok {
			return "", NewParserError("Expected identifier after '::'", p.stream.PeekTokenAt(0).Pos)
		}
		target = target + "::" + tok.Literal
	}
	return target, nil
}

func (p *Parser) parseRelationship(firstTargetTok tokenizer.Token) (ast.Relationship, error) {
	// Entry token is supposedly the first identifier
	var err error
	var rel ast.Relationship
	rel.LHS, err = p.parseRelationshipTarget(firstTargetTok)
	if err != nil {
		return rel, err
	}

	if multTok, ok := p.stream.TryConsumeType(tokenizer.STRING); ok {
		rel.MultLHS, err = ast.ParseCardinality(multTok.Literal)
		if err != nil {
			return rel, WrapParserError(err, multTok.Pos)
		}
	}

	p.parseArrowTokens(&rel)

	if multTok, ok := p.stream.TryConsumeType(tokenizer.STRING); ok {
		rel.MultRHS, err = ast.ParseCardinality(multTok.Literal)
		if err != nil {
			return rel, WrapParserError(err, multTok.Pos)
		}
	}

	if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.STRING) {
		return rel, NewParserError("Expected identifier or string after relationship", p.stream.PeekTokenAt(0).Pos)
	}
	rel.RHS, err = p.parseRelationshipTarget(p.stream.Emit())
	if err != nil {
		return rel, err
	}

	endingToken := p.stream.Emit()
	switch endingToken.Type {
	case tokenizer.NEWLINE, tokenizer.EOF:
		break
	case tokenizer.COLON:
		p.stream.SetRawMode("\n")
		rel.Label = p.stream.MustConsumeType(tokenizer.STRING).Literal
		if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
			return rel, NewParserError("Expected newline after relationship label", endingToken.Pos)
		}
	default:
		return rel, NewParserError("Expected newline or colon after relationship", endingToken.Pos)
	}

	return rel, nil
}

func (p *Parser) parseArrowTokens(rel *ast.Relationship) error {
	var sawDirection, sawAttrs, isLolipop bool
	tok := p.stream.Emit()
	switch tok.Type {
	case tokenizer.LANGLE:
		if pipeTok, ok := p.stream.TryConsumeType(tokenizer.PIPE); ok {
			tok = pipeTok
		}
		fallthrough
	case tokenizer.RBRACE:
		rel.LArrow = rune(tok.Literal[0])
	// lolipop interface
	case tokenizer.LPAREN:
		p.stream.MustConsumeType(tokenizer.RPAREN)
		isLolipop = true
		rel.LArrow = rune(tok.Literal[0])
	// position INDEPENDENT start and end tokens
	case tokenizer.IDENTIFIER:
		// special case for 'x' and 'o' in relationship
		if tok.Literal != "x" && tok.Literal != "o" {
			return NewParserError("Unexpected identifier in relationship definition", tok.Pos)
		}
		fallthrough
	case tokenizer.HASH, tokenizer.ASTERISK, tokenizer.PLUS, tokenizer.CARET:
		rel.LArrow = rune(tok.Literal[0])
	case tokenizer.DOT, tokenizer.DASH: // so that encountering them doesn't cause an error
	default:
		return NewParserError("Unexpected token at the start of relationship definition", tok.Pos)
	}

	if tok.Type != tokenizer.DOT && tok.Type != tokenizer.DASH {
		tok = p.stream.Emit() // consume the asserted start token, if any
	}

	bodyTokType := tok.Type
	var oppositeBodyTokType tokenizer.TokenType
	switch bodyTokType {
	case tokenizer.DOT:
		oppositeBodyTokType = tokenizer.DASH
	case tokenizer.DASH:
		oppositeBodyTokType = tokenizer.DOT
	default:
		return NewParserError("Unexpected token as the relationship body", tok.Pos)
	}
	var ok bool
	rel.Body = rune(tok.Literal[0])
	for tok.Type != tokenizer.EOF && tok.Type != tokenizer.NEWLINE {
		if tok, ok = p.stream.TryConsumeType(bodyTokType); ok {
			continue
		} else if tok, ok = p.stream.TryConsumeType(oppositeBodyTokType); ok {
			// Simply convenient error message
			return NewParserError("Different body type runes in relationship", tok.Pos)
		}
		if !p.stream.AssertType(tokenizer.LBRACKET) && !p.stream.AssertKW(keyword.Direction) {
			break
		} else if sawAttrs || sawDirection {
			return NewParserError("Cannot separate direction and attributes with a body token", tok.Pos)
		}

		if isLolipop {
			return NewParserError("Lolipop interface cannot contain attributes or direction", tok.Pos)
		}

		// We check these cases in order to be able to assert this squence:
		// -[attrs]->
		//    or
		// -dir->
		//    or
		// -[attrs]dir-> / -dir[attrs]->
		for range 2 {
			if !sawDirection {
				if tok, ok := p.stream.TryConsumeKW(keyword.Direction); ok {
					sawDirection = true
					var dir ast.DirectionKind
					switch tok.Literal {
					case "left", "l", "le":
						dir = ast.Left
					case "right", "r", "ri":
						dir = ast.Right
					case "up", "u":
						dir = ast.Top
					case "down", "d", "do":
						dir = ast.Bottom
					default:
						return NewParserError("Unexpected direction in relationship", tok.Pos)
					}
					rel.Direction = dir
				}
			}
			if !sawAttrs {
				if _, ok := p.stream.TryConsumeType(tokenizer.LBRACKET); ok {
					sawAttrs = true
					// We have matched attribute container start
					// after the body, now - parse the attributes
					var attrSB strings.Builder
					for tok = p.stream.Emit(); tok.Type != tokenizer.RBRACKET; tok = p.stream.Emit() {
						switch tok.Type {
						case tokenizer.EOF, tokenizer.NEWLINE:
							return NewParserError("Unexpected break in relationship attribute container", tok.Pos)
						case tokenizer.COMMA:
							if attrSB.Len() == 0 {
								return NewParserError("Unexpected comma in relationship attribute container", tok.Pos)
							}
							rel.Attrs = append(rel.Attrs, attrSB.String())
							attrSB.Reset()
							continue // actually useless, added for readability
						default:
							attrSB.WriteString(tok.Literal)
						}
					}
					if attrSB.Len() > 0 {
						rel.Attrs = append(rel.Attrs, attrSB.String())
					}
				}
			}
		}
		// If none matched, we mustn't consume trailing arrow body rune
		if !sawAttrs && !sawDirection {
			continue
		}

		// consume trailing arrow body rune
		if tok, ok = p.stream.TryConsumeType(bodyTokType); !ok {
			return NewParserError("Unexpected token in body relationship definition", tok.Pos)
		}
	}

	tok = p.stream.PeekTokenAt(0)
	switch tok.Type {
	case tokenizer.PIPE:
		// should only be encountered on --|> case as the end of the relationship
		p.stream.Emit()
		if _, ok := p.stream.TryConsumeType(tokenizer.RANGLE); !ok {
			return NewParserError("Expected '|>' after relationship", tok.Pos)
		}
		fallthrough
	case tokenizer.RANGLE, tokenizer.LBRACE:
		p.stream.Emit()
		// will return as this is the end of the relationship
		rel.RArrow = rune(tok.Literal[0])
	// lolipop interface
	case tokenizer.LPAREN:
		if sawDirection || sawAttrs {
			return NewParserError("Lolipop interface cannot contain direction or attributes", tok.Pos)
		}
		if isLolipop {
			return NewParserError("Cannot have double headed lolipop relationship", tok.Pos)
		}
		p.stream.Emit()
		p.stream.MustConsumeType(tokenizer.RPAREN)
		rel.RArrow = rune(tok.Literal[0])
	// position INDEPENDENT start and end tokens
	case tokenizer.IDENTIFIER:
		// special case for 'x' and 'o' in relationship
		if tok.Literal != "x" && tok.Literal != "o" {
			return NewParserError("Unexpected identifier in relationship definition", tok.Pos)
		}
		fallthrough
	case tokenizer.HASH, tokenizer.ASTERISK, tokenizer.PLUS, tokenizer.CARET:
		p.stream.Emit()
		rel.RArrow = rune(tok.Literal[0])
	}
	if rel.Body == 0 {
		return NewParserError("Missing body in the relationship", tok.Pos)
	}
	return nil
}
