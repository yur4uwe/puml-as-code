// Package parser implements a parser for PlantUML files
package parser

import (
	"errors"
	"fmt"
	"slices"
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

// parseSkinparam parses flat skinparam commands or nested skinparam blocks.
// NOTE: This function appends generated ast.StyleRule statements directly to p.ast.Statements
// instead of returning them, because a single skinparam block can expand into multiple StyleRule
// statements (one per selector hierarchy), whereas the main parser branch dispatch loop expects
// single-statement returns.
func (p *Parser) parseSkinparam() ([]ast.Statement, error) {
	// Peek paramTok to see if it's a target or a block
	paramTok := p.stream.Emit()
	if paramTok.Type == tokenizer.NEWLINE || paramTok.Type == tokenizer.EOF {
		return nil, NewParserError("Expected target or parameter after skinparam", paramTok.Pos)
	}

	name := paramTok.Literal
	stereo, _ := p.tryReadStereotype()

	var stmts []ast.Statement
	if p.stream.AssertType(tokenizer.LBRACE) {
		selectors := []string{name}
		if stereo != "" {
			selectors = append(selectors, stereo)
		}
		rules, err := p.parseSkinparamBlock(selectors)
		if err != nil {
			return nil, err
		}
		for _, r := range rules {
			stmts = append(stmts, r)
		}
	} else {
		// skinparam combinedName value
		value := p.stream.ReadUntilNewline()
		rule := &ast.StyleRule{
			Properties:  make(map[string]string),
			IsSkinparam: true,
		}
		if stereo != "" {
			rule.Selectors = append(rule.Selectors, stereo)
		}
		rule.Properties[name] = value
		stmts = append(stmts, rule)
	}

	return stmts, nil
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

func (p *Parser) setAliasAndName(ent *ast.Entity, nameOrAlias tokenizer.Token) error {
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
		var sb strings.Builder
		sb.WriteString(nameOrAlias.Literal)
		for {
			if sep, ok := p.stream.TryConsumePackageSeparator(); ok {
				sb.WriteString(sep)
				tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
				if !ok {
					return NewParserError("Expected identifier after package separator in entity name", p.stream.PeekTokenAt(0).Pos)
				}
				sb.WriteString(tok.Literal)
				continue
			}
			break
		}
		ent.Identifier = sb.String()
	default:
		return NewParserError("Expected token for entity identifier or alias", nameOrAlias.Pos)
	}
	return nil
}

// tok is the kind of an entity (class, interface, struct, enum, etc.)
func (p *Parser) parseEntity(tok tokenizer.Token) (*ast.Entity, error) {
	ent := &ast.Entity{
		Kind:          p.mapTokenToEntityKind(tok),
		LeadingTrivia: tok.LeadingTrivia,
	}

	// Unexpectedly, class and other entity definitions have very strict syntax:
	// class <entity name> as <entity alias> <generics> <stereotype> <styles> <body>

	if err := p.setAliasAndName(ent, p.stream.Emit()); err != nil {
		return nil, err
	}

	if _, hasAlias := p.stream.TryConsumeKW(keyword.Alias); hasAlias {
		if err := p.setAliasAndName(ent, p.stream.Emit()); err != nil {
			return nil, err
		}
	}

	if ent.Identifier == "" {
		// If no identifier is set, use the alias
		ent.Identifier = ent.Alias
	}

	if !p.stream.AssertSeq([]tokenizer.Token{{Type: tokenizer.LANGLE}, {Type: tokenizer.LANGLE}}) {
		gen, err := p.tryReadGeneric()
		if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
			return nil, err
		}
		ent.Generic = gen
	}

	stereo, err := p.tryReadStereotype()
	if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
		return nil, err
	}
	ent.Stereotype = stereo

	ent.Color = p.tryParseColor()

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
		Kind:          p.mapKeywordToContainerKind(containerClass),
		LeadingTrivia: tok.LeadingTrivia,
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
		container.Stereotype, err = p.tryReadStereotype()
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
		} else if tok.Type == tokenizer.NEWLINE {
			// Skip NEWLINE tokens inside parseContainer() so empty lines within
			// container blocks { ... } do not trigger an erroneous "Expected a statement in a
			// container body" error.
			continue
		}

		// Only parse statements allowed in containers
		stmts, err := p.parseContainerStatement(tok)
		if err != nil {
			return container, err
		}
		if len(stmts) == 0 {
			return container, NewParserError("Expected a statement in a container body", tok.Pos)
		}
		container.Statements = append(container.Statements, stmts...)
	}
	return container, nil
}

func (p *Parser) parseSetDirective() (ast.Statement, error) {
	tok := p.stream.PeekTokenAt(0)
	keyTok := p.stream.Emit() // consume "separator"

	var sb strings.Builder
	for tok = p.stream.Emit(); tok.Type != tokenizer.NEWLINE && tok.Type != tokenizer.EOF; tok = p.stream.Emit() {
		sb.WriteString(tok.Literal)
	}
	sepVal := sb.String()
	if keyTok.Literal == "separator" {
		if sepVal == "none" {
			p.stream.PackageSeparator = ""
		} else {
			p.stream.PackageSeparator = sepVal
		}
	}

	return ast.SetCommand{
		Key:   keyTok.Literal,
		Value: sepVal,
	}, nil
}

func (p *Parser) parseContinerIdent(tok tokenizer.Token) (tokenizer.Token, error) {
	if tok.Type == tokenizer.STRING {
		return tok, nil
	}
	if tok.Type != tokenizer.IDENTIFIER {
		return tokenizer.Token{}, NewParserError("Expected identifier for container name", tok.Pos)
	}

	var sb strings.Builder
	sb.WriteString(tok.Literal)
	lastTok := tok
	for tok = p.stream.PeekTokenAt(0); tok.Type != tokenizer.EOF; tok = p.stream.PeekTokenAt(0) {
		// early exit if we see an alias keyword and not hit the
		// the double identifier case
		if keyword.Classify(tok.Literal) == keyword.Alias {
			return amb(tokenizer.IDENTIFIER, sb.String()), nil
		}
		if sep, ok := p.stream.TryConsumePackageSeparator(); ok {
			sb.WriteString(sep)
			lastTok = tokenizer.Token{Type: tokenizer.DOT}
			continue
		}
		if lastTok.Type == tokenizer.IDENTIFIER && tok.Type == tokenizer.IDENTIFIER {
			return tokenizer.Token{}, NewParserError("Expected container name to be a single identifier", tok.Pos)
		}
		switch tok.Type {
		case tokenizer.LBRACE, tokenizer.LANGLE, tokenizer.HASH, tokenizer.NEWLINE:
			return amb(tokenizer.IDENTIFIER, sb.String()), nil
		default:
			sb.WriteString(tok.Literal)
			p.stream.Emit()
			lastTok = tok
		}
	}

	return tokenizer.Token{}, tokenizer.ErrUnexpectedEOF
}

// parseContainerIdentAndAlias parses the container name and alias
//
// Returns
// - Alias
// - Identifier
// - error
func (p *Parser) parseContainerIdentAndAlias() (string, string, error) {
	lhs, err := p.parseContinerIdent(p.stream.Emit())
	if err != nil {
		return "", "", err
	}
	if _, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		return "", lhs.Literal, nil
	}
	rhs, err := p.parseContinerIdent(p.stream.Emit())
	if err != nil {
		return "", "", err
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
	noteTok := tok
	// tok is a keyword 'note'
	tok = p.stream.Emit()
	var note ast.Note
	var err error
	if tok.Type == tokenizer.STRING {
		note, err = p.parseInlineAliasNote(tok)
	} else {
		class := keyword.Classify(tok.Literal)
		switch class {
		case keyword.Direction:
			note, err = p.parseReltiveNote(tok)
		case keyword.Position:
			note, err = p.parseLinkNote(tok)
		case keyword.Alias:
			note, err = p.parseMultilineAliasNote()
		default:
			return ast.Note{}, WrapParserError(fmt.Errorf("expected direction, string, note position or alias after 'note', got %s", class.String()), tok.Pos)
		}
	}
	if err == nil {
		note.LeadingTrivia = noteTok.LeadingTrivia
	}
	return note, err
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
	} else if tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER); ok {
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
	tokens := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.COLON)
	return p.stream.TokensToString(tokens)
}

func (p *Parser) parseNoteBody() (string, error) {
	tok := p.stream.Emit()

	noteEndSequence := []tokenizer.Token{unamb(tokenizer.NEWLINE)}
	switch tok.Type {
	case tokenizer.COLON:
		// to not error out
	case tokenizer.NEWLINE:
		noteEndSequence = append(noteEndSequence, amb(tokenizer.IDENTIFIER, "end"), amb(tokenizer.IDENTIFIER, "note"))
	default:
		return "", NewParserError("Expected ':' or newline after note definition", tok.Pos)
	}

	return p.stream.ConsumeTextBlock(noteEndSequence)
}

func (p *Parser) mapTokenToDirection(tok tokenizer.Token) ast.DirectionKind {
	switch tok.Literal {
	case "left":
		return ast.DirectionLeft
	case "right":
		return ast.DirectionRight
	case "top":
		return ast.DirectionTop
	case "bottom":
		return ast.DirectionBottom
	default:
		return ast.DirectionUnknown
	}
}

func (p *Parser) parseSkinparamBlock(selectors []string) ([]*ast.StyleRule, error) {
	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		return nil, NewParserError("Expected opening brace after skinparam target", p.stream.PeekTokenAt(0).Pos)
	}

	currentRule := &ast.StyleRule{
		Selectors:   slices.Clone(selectors),
		Properties:  make(map[string]string),
		IsSkinparam: true,
	}
	var rules []*ast.StyleRule

	for tok := p.stream.Emit(); tok.Type != tokenizer.RBRACE; tok = p.stream.Emit() {
		if tok.Type == tokenizer.NEWLINE {
			continue
		}

		if tok.Type == tokenizer.EOF {
			return nil, NewParserError("Unexpected EOF in skinparam block", tok.Pos)
		}

		// Read target or param
		name := tok.Literal
		stereo, _ := p.tryReadStereotype()

		if p.stream.AssertType(tokenizer.LBRACE) {
			// Recursive sub-block with accumulated selectors
			subSelectors := append(slices.Clone(selectors), name)
			if stereo != "" {
				subSelectors = append(subSelectors, stereo)
			}
			subRules, err := p.parseSkinparamBlock(subSelectors)
			if err != nil {
				return nil, err
			}
			rules = append(rules, subRules...)
		} else {
			// Inline value
			value := p.stream.ReadUntilNewline()
			if stereo != "" {
				// Inline stereotype modifier for a property in block
				currentRule.Properties[name+"."+stereo] = value
			} else {
				currentRule.Properties[name] = value
			}
		}
	}

	if len(currentRule.Properties) > 0 {
		rules = append([]*ast.StyleRule{currentRule}, rules...)
	}
	return rules, nil
}

func (p *Parser) isStyleTagEnd() bool {
	return p.stream.AssertSeq([]tokenizer.Token{
		{Type: tokenizer.LANGLE},
		{Type: tokenizer.SLASH},
		{Type: tokenizer.IDENTIFIER, Literal: "style"},
		{Type: tokenizer.RANGLE},
	})
}

func (p *Parser) parseStyleBlock(startTok tokenizer.Token) ([]ast.Statement, error) {
	p.stream.Emit() // consume 'style'
	p.stream.Emit() // consume '>'

	rules, err := p.parseStyleRules([]string{})
	if err != nil {
		return nil, err
	}

	if !p.isStyleTagEnd() {
		return nil, NewParserError("Expected </style> closing tag", p.stream.PeekTokenAt(0).Pos)
	}

	// Consume '</style>'
	for range 4 {
		p.stream.Emit()
	}

	for _, r := range rules {
		r.(*ast.StyleRule).LeadingTrivia = startTok.LeadingTrivia
	}
	return rules, nil
}

func (p *Parser) parseStyleRules(selectors []string) ([]ast.Statement, error) {
	currentRule := &ast.StyleRule{
		Selectors:   slices.Clone(selectors),
		Properties:  make(map[string]string),
		IsSkinparam: false,
	}
	var rules []ast.Statement

	for !p.isStyleTagEnd() && !p.stream.AssertType(tokenizer.RBRACE) && !p.stream.AssertType(tokenizer.EOF) {

		tok := p.stream.Emit()
		if tok.Type == tokenizer.NEWLINE {
			continue
		}

		name := tok.Literal
		stereo, _ := p.tryReadStereotype()

		if p.stream.AssertType(tokenizer.LBRACE) {
			p.stream.Emit() // consume '{'
			subSelectors := append(slices.Clone(selectors), name)
			if stereo != "" {
				subSelectors = append(subSelectors, stereo)
			}
			subRules, err := p.parseStyleRules(subSelectors)
			if err != nil {
				return nil, err
			}
			rules = append(rules, subRules...)
			if p.stream.AssertType(tokenizer.RBRACE) {
				p.stream.Emit() // consume '}'
			}
		} else {
			p.stream.TryConsumeType(tokenizer.COLON) // consume optional ':'
			valTokens := p.stream.ConsumeUntilType(tokenizer.SEMICOLON, tokenizer.NEWLINE, tokenizer.EOF)
			if p.stream.AssertType(tokenizer.SEMICOLON) {
				p.stream.Emit() // consume ';'
			}
			val := strings.TrimSpace(p.stream.TokensToString(valTokens))
			if stereo != "" {
				currentRule.Properties[name+"."+stereo] = val
			} else {
				currentRule.Properties[name] = val
			}
		}
	}

	if len(currentRule.Properties) > 0 {
		rules = append([]ast.Statement{currentRule}, rules...)
	}
	return rules, nil
}

func (p *Parser) parseTitle() (ast.Statement, error) {
	if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); ok {
		// Multi-line title block ending with 'end title'
		titleEndSequence := []tokenizer.Token{
			unamb(tokenizer.NEWLINE),
			amb(tokenizer.IDENTIFIER, "end"),
			amb(tokenizer.IDENTIFIER, "title"),
		}
		var err error
		p.ast.Title, err = p.stream.ConsumeTextBlock(titleEndSequence)
		return ast.TitleDef{Text: p.ast.Title}, err
	}

	// Single-line title
	p.ast.Title = p.stream.ReadUntilNewline()
	return ast.TitleDef{Text: p.ast.Title}, nil
}

func (p *Parser) parseRelationshipTarget(entityNameTok tokenizer.Token) (string, error) {
	var sb strings.Builder
	sb.WriteString(entityNameTok.Literal)

	for {
		if sep, ok := p.stream.TryConsumePackageSeparator(); ok {
			sb.WriteString(sep)
			tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
			if !ok {
				return "", NewParserError("Expected identifier after package separator in relationship target", p.stream.PeekTokenAt(0).Pos)
			}
			sb.WriteString(tok.Literal)
			continue
		}
		break
	}

	if p.stream.PeekTokenAt(0).Type == tokenizer.COLON && p.stream.PeekTokenAt(1).Type == tokenizer.COLON {
		p.stream.Emit() // consume first :
		p.stream.Emit() // consume second :

		tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
		if !ok {
			return "", NewParserError("Expected identifier after '::'", p.stream.PeekTokenAt(0).Pos)
		}
		sb.WriteString("::")
		sb.WriteString(tok.Literal)
	}
	return sb.String(), nil
}

func (p *Parser) parseRelationship(firstTargetTok tokenizer.Token) (ast.Relationship, error) {
	// Entry token is supposedly the first identifier
	var err error
	var rel ast.Relationship
	rel.LeadingTrivia = firstTargetTok.LeadingTrivia
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

	endingToken := p.stream.PeekTokenAt(0)
	switch endingToken.Type {
	case tokenizer.NEWLINE, tokenizer.EOF:
		p.stream.Emit()
	case tokenizer.RBRACE:
		// Do not consume RBRACE so enclosing container block loop sees it
		break
	case tokenizer.COLON:
		p.stream.Emit()
		rel.Label, err = p.stream.ConsumeTextBlock([]tokenizer.Token{unamb(tokenizer.NEWLINE)})
		if err != nil {
			return rel, err
		}
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
		switch rel.LArrow {
		case '<':
			rel.TypeLHS = ast.RelationDependency
		case '|':
			rel.TypeLHS = ast.RelationRealization
		}
	case tokenizer.DASH:
		oppositeBodyTokType = tokenizer.DOT
		switch rel.LArrow {
		case '<':
			rel.TypeLHS = ast.RelationAssociation
		case 'o':
			rel.TypeLHS = ast.RelationAggregation
		case '*':
			rel.TypeLHS = ast.RelationComposition
		case '|':
			rel.TypeLHS = ast.RelationInheritance
		}
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
						dir = ast.DirectionLeft
					case "right", "r", "ri":
						dir = ast.DirectionRight
					case "up", "u":
						dir = ast.DirectionTop
					case "down", "d", "do":
						dir = ast.DirectionBottom
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
		switch rel.Body {
		case '-':
			rel.TypeRHS = ast.RelationInheritance
		case '.':
			rel.TypeRHS = ast.RelationRealization
		}
		rel.RArrow = rune(tok.Literal[0])
	case tokenizer.RANGLE:
		switch rel.Body {
		case '-':
			rel.TypeRHS = ast.RelationAssociation
		case '.':
			rel.TypeRHS = ast.RelationDependency
		}
		fallthrough
	case tokenizer.LBRACE:
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
	case tokenizer.IDENTIFIER, tokenizer.ASTERISK:
		// special case for 'x' and 'o' in relationship
		switch tok.Literal {
		case "*":
			rel.TypeRHS = ast.RelationComposition
		case "x":
		case "o":
			if rel.Body == '-' {
				rel.TypeRHS = ast.RelationAggregation
			}
		default:
			return NewParserError("Unexpected identifier in relationship definition", tok.Pos)
		}
		fallthrough
	case tokenizer.HASH, tokenizer.PLUS, tokenizer.CARET:
		p.stream.Emit()
		rel.RArrow = rune(tok.Literal[0])
	default:
		if rel.TypeLHS == ast.RelationUnknown && rel.Body == '-' {
			rel.TypeLHS = ast.RelationAssociation
			rel.TypeRHS = ast.RelationAssociation
		}
	}
	if rel.Body == 0 {
		return NewParserError("Missing body in the relationship", tok.Pos)
	}
	return nil
}

func (p *Parser) HasArrowOnLine() bool {
	for i := 0; ; i++ {
		tok := p.stream.PeekTokenAt(i)
		if tok.Type == tokenizer.NEWLINE || tok.Type == tokenizer.EOF ||
			tok.Type == tokenizer.SEMICOLON || tok.Type == tokenizer.LBRACE ||
			tok.Type == tokenizer.RBRACE {
			break
		} else if tok.Type == tokenizer.STRING {
			continue
		}
		if _, ok := p.scanArrowTokensFrom(i); ok {
			return true
		}
	}
	return false
}

func (p *Parser) scanArrowTokensFrom(startIdx int) (int, bool) {
	idx := startIdx
	var sawDirection, sawAttrs, isLolipop bool

	tok := p.stream.PeekTokenAt(idx)
	idx++

	switch tok.Type {
	case tokenizer.LANGLE:
		if p.stream.PeekTokenAt(idx).Type == tokenizer.PIPE {
			idx++
		}
	case tokenizer.RBRACE:
		// valid left arrow head
	case tokenizer.LPAREN:
		if p.stream.PeekTokenAt(idx).Type != tokenizer.RPAREN {
			return 0, false
		}
		idx++
		isLolipop = true
	case tokenizer.IDENTIFIER:
		if tok.Literal != "x" && tok.Literal != "o" {
			return 0, false
		}
	case tokenizer.HASH, tokenizer.ASTERISK, tokenizer.PLUS, tokenizer.CARET:
		// valid left arrow head
	case tokenizer.DOT, tokenizer.DASH:
		idx--
	default:
		return 0, false
	}

	tok = p.stream.PeekTokenAt(idx)
	idx++

	bodyTokType := tok.Type
	var oppositeBodyTokType tokenizer.TokenType
	switch bodyTokType {
	case tokenizer.DOT:
		oppositeBodyTokType = tokenizer.DASH
	case tokenizer.DASH:
		oppositeBodyTokType = tokenizer.DOT
	default:
		return 0, false
	}

	hasBody := false
	for {
		nextTok := p.stream.PeekTokenAt(idx)
		if nextTok.Type == tokenizer.EOF || nextTok.Type == tokenizer.NEWLINE {
			break
		}
		switch nextTok.Type {
		case bodyTokType:
			hasBody = true
			idx++
			continue
		case oppositeBodyTokType:
			return 0, false
		}

		if nextTok.Type != tokenizer.LBRACKET && keyword.Classify(nextTok.Literal) != keyword.Direction {
			break
		} else if sawAttrs || sawDirection {
			return 0, false
		}

		if isLolipop {
			return 0, false
		}

		for range 2 {
			curr := p.stream.PeekTokenAt(idx)
			if !sawDirection && keyword.Classify(curr.Literal) == keyword.Direction {
				sawDirection = true
				idx++
			}
			curr = p.stream.PeekTokenAt(idx)
			if !sawAttrs && curr.Type == tokenizer.LBRACKET {
				sawAttrs = true
				idx++ // consume [
				for {
					attrTok := p.stream.PeekTokenAt(idx)
					idx++
					if attrTok.Type == tokenizer.RBRACKET {
						break
					}
					if attrTok.Type == tokenizer.EOF || attrTok.Type == tokenizer.NEWLINE {
						return 0, false
					}
				}
			}
		}

		if !sawAttrs && !sawDirection {
			continue
		}

		if p.stream.PeekTokenAt(idx).Type != bodyTokType {
			return 0, false
		}
		idx++
	}

	rTok := p.stream.PeekTokenAt(idx)
	hasRightArrowhead := false
	switch rTok.Type {
	case tokenizer.PIPE:
		idx++
		if p.stream.PeekTokenAt(idx).Type != tokenizer.RANGLE {
			return 0, false
		}
		idx++
		hasRightArrowhead = true
	case tokenizer.RANGLE, tokenizer.LBRACE:
		idx++
		hasRightArrowhead = true
	case tokenizer.LPAREN:
		if sawDirection || sawAttrs || isLolipop {
			return 0, false
		}
		idx++
		if p.stream.PeekTokenAt(idx).Type != tokenizer.RPAREN {
			return 0, false
		}
		idx++
		hasRightArrowhead = true
	case tokenizer.IDENTIFIER:
		if rTok.Literal != "x" && rTok.Literal != "o" {
			return 0, false
		}
		idx++
		hasRightArrowhead = true
	case tokenizer.HASH, tokenizer.ASTERISK, tokenizer.PLUS, tokenizer.CARET:
		idx++
		hasRightArrowhead = true
	}

	if !hasBody && !hasRightArrowhead && !isLolipop {
		return 0, false
	}

	return idx - startIdx, true
}
