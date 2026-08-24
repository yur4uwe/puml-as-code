// Package parser implements a parser for PlantUML files
package parser

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseVisibilityCommand(tok tokenizer.Token) (ast.VisibilityCommand, error) {
	cmd := ast.VisibilityCommand{
		Kind: ast.VisibilityCMDUnknown,
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}
	switch keyword.Classify(tok.Literal) {
	case keyword.Hide:
		cmd.Kind = ast.VisibilityCMDHide
	case keyword.Show:
		cmd.Kind = ast.VisibilityCMDShow
	case keyword.Remove:
		cmd.Kind = ast.VisibilityCMDRemove
	case keyword.Restore:
		cmd.Kind = ast.VisibilityCMDRestore
	}
	cmd.Target = p.stream.ReadUntilNewline()
	cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
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

	cmd := ast.DirectionCommand{
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}

	for _, token := range expectedSeq {
		if _, ok := p.stream.TryConsume(token); !ok {
			return cmd, NewParserError("Unexpected diagram direction modifier", token)
		}
	}
	switch tok.Literal {
	case "left":
		cmd.Direction = ast.LeftToRightDirection
	case "top":
		cmd.Direction = ast.TopToBottomDirection
	}
	if res := p.stream.ConsumeUntilType(tokenizer.NEWLINE); len(res) != 0 {
		return cmd, NewParserError("Unexpected tokens after direction command", p.stream.PeekTokenAt(0))
	}
	cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
	return cmd, nil
}

func (p *Parser) parseDirective(tok1 tokenizer.Token) (ast.Statement, error) {
	directiveNameTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return nil, NewParserError("Expected directive name", p.stream.PeekTokenAt(0))
	}

	if directiveNameTok.Pos.Offset != tok1.Pos.Offset+1 {
		return nil, NewParserError("Expected directive name right after !", directiveNameTok)
	}

	if strings.HasPrefix(directiveNameTok.Literal, "include") {
		return p.parseIncludeDirective(directiveNameTok)
	}
	return nil, NewParserError("Unknown directive", directiveNameTok)
}

func (p *Parser) parseIncludeDirective(tok tokenizer.Token) (ast.IncludeDirective, error) {
	var kind ast.IncludeKind
	switch tok.Literal {
	case "include_many":
		kind = ast.IncludeMany
	case "include_once", "include":
		kind = ast.IncludeOnce
	default:
		return ast.IncludeDirective{}, NewParserError("Unknown include directive", tok)
	}

	dir := ast.IncludeDirective{
		Kind: kind,
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}

	filePathToks := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.EXCLAMATION)
	dir.Path = p.stream.TokensToString(filePathToks)
	if p.stream.AssertType(tokenizer.EXCLAMATION) {
		p.stream.Emit() // consume '!'
		// Id or order must be a single token
		dir.Tag = p.stream.Emit().Literal
	}

	p.stream.EmitCommentToks()

	dir.TrailingTrivia = p.stream.DumpCollectedTrivia()

	return dir, nil
}

// parseSkinparam parses flat skinparam commands or nested skinparam blocks.
// NOTE: This function appends generated ast.StyleRule statements directly to p.ast.Statements
// instead of returning them, because a single skinparam block can expand into multiple StyleRule
// statements (one per selector hierarchy), whereas the main parser branch dispatch loop expects
// single-statement returns.
func (p *Parser) parseSkinparam() ([]ast.Statement, error) {
	leadingTrivia := p.stream.DumpCollectedTrivia()
	// Peek paramTok to see if it's a target or a block
	paramTok := p.stream.Emit()
	if paramTok.Type == tokenizer.NEWLINE || paramTok.Type == tokenizer.EOF {
		return nil, NewParserError("Expected target or parameter after skinparam", paramTok)
	}

	name := paramTok.Literal
	stereo, _ := p.tryReadStereotype()

	var stmts []ast.Statement
	if p.stream.AssertType(tokenizer.LBRACE) {
		selectors := []string{name}
		if stereo != "" {
			selectors = append(selectors, stereo)
		}
		rules, err := p.parseSkinparamBlock(selectors, p.stream.DumpCollectedTrivia())
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
			Trivia: ast.Trivia{
				LeadingTrivia:  leadingTrivia,
				TrailingTrivia: p.stream.DumpCollectedTrivia(),
			},
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
	cmd := ast.ScaleCommand{
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}
	if _, ok := p.stream.TryConsume(tokenizer.Token{Type: tokenizer.IDENTIFIER, Literal: "max"}); ok {
		cmd.IsMax = true
	}

	tok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
	if !ok {
		// Try to handle 200x300 which is tokenized as an IDENTIFIER
		if tok, ok = p.stream.TryConsumeType(tokenizer.IDENTIFIER); !ok {
			return cmd, NewParserError("Expected number after scale", p.stream.PeekTokenAt(0))
		}
		if !strings.Contains(tok.Literal, "x") {
			return cmd, NewParserError("Expected number after scale", tok)
		}
		parts := strings.Split(tok.Literal, "x")
		if len(parts) == 2 {
			w, errW := strconv.Atoi(parts[0])
			h, errH := strconv.Atoi(parts[1])
			if errW == nil && errH == nil {
				cmd.Width = w
				cmd.Height = h
				return cmd, nil
			}
		}
	}

	val, err := strconv.ParseFloat(tok.Literal, 64)
	if err != nil {
		return cmd, NewParserError(fmt.Sprintf("Invalid number: %s", err.Error()), tok)
	}

	tok = p.stream.Emit()
	switch tok.Type {
	case tokenizer.IDENTIFIER:
		switch tok.Literal {
		case "width":
			cmd.Width, ok = toInteger(val)
			if !ok {
				return cmd, NewParserError("Expected width to be an integer", tok)
			}
		case "height":
			cmd.Height, ok = toInteger(val)
			if !ok {
				return cmd, NewParserError("Expected height to be an integer", tok)
			}
		case "x":
			cmd.Width, ok = toInteger(val)
			if !ok {
				return cmd, NewParserError("Expected width to be an integer", tok)
			}
			heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
			if !ok {
				return cmd, NewParserError("Expected height after 'x'", p.stream.PeekTokenAt(0))
			}
			h, err := strconv.Atoi(heightTok.Literal)
			if err != nil {
				return cmd, NewParserError("Invalid height", heightTok)
			}
			cmd.Height = h
		default:
			return cmd, NewParserError(fmt.Sprintf("Unexpected identifier: %s", tok.Literal), tok)
		}
	case tokenizer.ASTERISK:
		cmd.Width, ok = toInteger(val)
		if !ok {
			return cmd, NewParserError("Expected width to be an integer", tok)
		}
		heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return cmd, NewParserError("Expected height after '*'", p.stream.PeekTokenAt(0))
		}
		h, err := strconv.Atoi(heightTok.Literal)
		if err != nil {
			return cmd, NewParserError("Invalid height", heightTok)
		}
		cmd.Height = h
	case tokenizer.SLASH:
		numer, ok := toInteger(val)
		if !ok {
			return cmd, NewParserError("Expected numerator to be an integer", tok)
		}
		denomTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return cmd, NewParserError("Expected denominator after '/'", p.stream.PeekTokenAt(0))
		}
		denom, err := strconv.Atoi(denomTok.Literal)
		if err != nil {
			return cmd, NewParserError("Invalid denominator", denomTok)
		}
		if denom == 0 {
			return cmd, NewParserError("Denominator cannot be zero", denomTok)
		}
		cmd.Scale = float64(numer) / float64(denom)
	case tokenizer.NEWLINE, tokenizer.EOF:
		// EOF handling is a special case for a test case
		cmd.Scale = val
	default:
		return cmd, NewParserError(fmt.Sprintf("Unexpected token: %s(%s)", tok.Type.String(), tok.Literal), tok)
	}

	if res := p.stream.ConsumeUntilType(tokenizer.NEWLINE); len(res) != 0 {
		return cmd, NewParserError("Unexpected tokens after scale command", p.stream.PeekTokenAt(0))
	}
	cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
	return cmd, nil
}

func (p *Parser) setAliasAndName(ent *ast.Entity, nameOrAlias tokenizer.Token) ([]string, error) {
	switch nameOrAlias.Type {
	case tokenizer.STRING:
		if ent.Alias != "" {
			return nil, NewParserError("Entity alias already set", nameOrAlias)
		}
		ent.Alias = nameOrAlias.Literal
		return nil, nil
	case tokenizer.IDENTIFIER:
		if ent.Identifier != "" {
			return nil, NewParserError("Entity name already set", nameOrAlias)
		}
		if _, ok := p.stream.TryConsumePackageSeparator(); !ok {
			ent.Identifier = nameOrAlias.Literal
			return nil, nil
		}
		pkgPath := []string{nameOrAlias.Literal}
		for {
			tok := p.stream.Emit()
			if tok.Type != tokenizer.IDENTIFIER {
				return nil, NewParserError("Expected identifier after package separator", tok)
			}
			ent.Identifier = tok.Literal

			if _, ok := p.stream.TryConsumePackageSeparator(); !ok {
				break
			}
			pkgPath = append(pkgPath, tok.Literal)
		}
		return pkgPath, nil
	default:
		return nil, NewParserError("Expected token for entity identifier or alias", nameOrAlias)
	}
}

func wrapInContainers(ent ast.Entity, pkgPath []string) ast.Statement {
	var current ast.Statement = ent
	for _, pkg := range slices.Backward(pkgPath) {
		// Leave with Kind = ast.ContainerUnknown
		// This will be a good hint to distinguish between
		// inline and block container declarations
		current = ast.Container{
			Identifier: pkg,
			Statements: []ast.Statement{current},
		}
	}
	return current
}

// tok is the kind of an entity (class, interface, struct, enum, etc.)
func (p *Parser) parseEntity(tok tokenizer.Token) (ast.Statement, error) {
	ent := &ast.Entity{
		Kind: p.mapTokenToEntityKind(tok),
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}

	// Unexpectedly, class and other entity definitions have very strict syntax:
	// class <entity name> as <entity alias> <generics> <stereotype> <styles> <body>

	pkgPath, err := p.setAliasAndName(ent, p.stream.Emit())
	if err != nil {
		return nil, err
	}

	if _, hasAlias := p.stream.TryConsumeKW(keyword.Alias); hasAlias {
		aliasPkgPath, err := p.setAliasAndName(ent, p.stream.Emit())
		if err != nil {
			return nil, err
		}
		if len(aliasPkgPath) > 0 {
			pkgPath = aliasPkgPath
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

	preTags := p.tryReadTags()
	stereo, err := p.tryReadStereotype()
	if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
		return nil, err
	}
	ent.Stereotype = stereo
	postTags := p.tryReadTags()
	if len(preTags) > 0 {
		if len(postTags) > 0 {
			log.Printf("warning: tags %v after stereotype on entity %q are ignored in favor of pre-stereotype tags %v\n", postTags, ent.Identifier, preTags)
		}
		ent.Tags = preTags
	} else if len(postTags) > 0 {
		ent.Tags = postTags
	}

	ent.Color = p.tryParseColor()

	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		// No body return entity as is
		ent.TrailingTrivia = p.stream.DumpCollectedTrivia()
		return wrapInContainers(*ent, pkgPath), nil
	}

	p.stream.EmitCommentToks()
	ent.TrailingTrivia = p.stream.DumpCollectedTrivia()

	// We are inside a body
	for {
		for {
			if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
				break
			}
		}

		member, err := p.parseEntityMember()
		if err != nil {
			return nil, err
		}
		if member == nil {
			break
		}

		ent.Members = append(ent.Members, member)
	}

	p.stream.EmitCommentToks()

	if closingTrivia := p.stream.DumpCollectedTrivia(); len(closingTrivia) > 0 {
		ent.TrailingTrivia = append(ent.TrailingTrivia, closingTrivia...)
	}
	return wrapInContainers(*ent, pkgPath), nil
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
		leadingTrivia := p.stream.DumpCollectedTrivia()
		member, err = p.parseFieldOrMethod(&mod, vis, unamb(tokenizer.LBRACE), leadingTrivia)
		if err != nil {
			return nil, err
		}
		return member, nil
	} else if err != tokenizer.ErrStartMarkerNotFound {
		// Ignore the particular error because of the reasons above
		return nil, err
	}

	tok := p.stream.Emit()
	leadingTrivia := p.stream.DumpCollectedTrivia()
	switch tok.Type {
	case tokenizer.RBRACE:
		// It shouldn't arrive here, but just in case
		return nil, nil
	case tokenizer.DASH, tokenizer.TILDE, tokenizer.HASH, tokenizer.PLUS:
		vis = p.mapTokenToVisibility(tok.Type)
	case tokenizer.IDENTIFIER:
		// simply to not error out
	case tokenizer.NEWLINE:
		// We encountered a comment
		// We should retry the whole loop
	default:
		return nil, NewParserError("Unexpected token in entity body", tok)
	}
	return p.parseFieldOrMethod(nil, vis, tok, leadingTrivia)
}

// Can be either a field or a method
// but start with a field as either start the same
//
// Returns the parsed field or method, must be asserted to be a field or method
func (p *Parser) parseFieldOrMethod(mod *string, vis ast.VisibilityKind, entryTok tokenizer.Token, leadingTrivia []tokenizer.Token) (ast.Member, error) {
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
	var mods []string = nil

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
				vis = p.mapTokenToVisibility(tok.Type)
				canEncounterVisibility = false
				continue
			}
			entry = append(entry, tok)
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
			entryTok,
		)
	}

	isMethod := mustBeMethod || (!mustBeField && containsLParen)

	trailingTrivia := p.stream.DumpCollectedTrivia()
	opts := dialect.MemberOptions{
		Visibility:     vis,
		Modifiers:      mods,
		LeadingTrivia:  leadingTrivia,
		TrailingTrivia: trailingTrivia,
	}
	if isMethod {
		return p.Dialect.ParseMethod(entry, &opts)
	} else {
		return p.Dialect.ParseField(entry, &opts)
	}
}

func (p *Parser) parseContainer(tok tokenizer.Token) (ast.Container, error) {
	containerClass := keyword.Classify(tok.Literal)
	container := ast.Container{
		Kind: p.mapKeywordToContainerKind(containerClass),
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
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
		preTags := p.tryReadTags()
		container.Stereotype, err = p.tryReadStereotype()
		if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
			return container, err
		}
		postTags := p.tryReadTags()
		if len(preTags) > 0 {
			if len(postTags) > 0 {
				log.Printf("warning: tags %v after stereotype on container %q are ignored in favor of pre-stereotype tags %v\n", postTags, container.Identifier, preTags)
			}
			container.Tags = preTags
		} else if len(postTags) > 0 {
			container.Tags = postTags
		}

		container.Color = p.tryParseColor()
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		if _, ok = p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
			return container, NewParserError("Expected container body to end", p.stream.PeekTokenAt(0))
		}
		container.TrailingTrivia = p.stream.DumpCollectedTrivia()
		return p.wrapImplicitPackageContainers(container), nil
	}

	p.stream.EmitCommentToks()
	container.TrailingTrivia = p.stream.DumpCollectedTrivia()

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
			return container, NewParserError("Expected a statement in a container body", tok)
		}
		container.Statements = append(container.Statements, stmts...)
	}

	p.stream.EmitCommentToks()
	if closingTrivia := p.stream.DumpCollectedTrivia(); len(closingTrivia) > 0 {
		container.TrailingTrivia = append(container.TrailingTrivia, closingTrivia...)
	}
	return p.wrapImplicitPackageContainers(container), nil
}

func (p *Parser) wrapImplicitPackageContainers(c ast.Container) ast.Container {
	if p.stream.PackageSeparator == "" || !strings.Contains(c.Identifier, p.stream.PackageSeparator) {
		return c
	}
	segments := strings.Split(c.Identifier, p.stream.PackageSeparator)
	if len(segments) <= 1 {
		return c
	}

	current := ast.Container{
		Kind:       c.Kind,
		Identifier: segments[len(segments)-1],
		Alias:      c.Alias,
		Stereotype: c.Stereotype,
		Color:      c.Color,
		Statements: c.Statements,
	}
	for _, pkg := range slices.Backward(segments[:len(segments)-1]) {
		current = ast.Container{
			Kind:       c.Kind,
			Identifier: pkg,
			Statements: []ast.Statement{current},
		}
	}
	return current
}

func (p *Parser) parseSetDirective() (ast.Statement, error) {
	leadingTrivia := p.stream.DumpCollectedTrivia()
	tok := p.stream.PeekTokenAt(0)
	keyTok := p.stream.Emit() // consume "separator"

	var sb strings.Builder
	for tok = p.stream.Emit(); tok.Type != tokenizer.NEWLINE && tok.Type != tokenizer.EOF; tok = p.stream.Emit() {
		sb.WriteString(tok.Literal)
	}
	directiveVal := sb.String()
	if keyTok.Literal == "separator" {
		if directiveVal == "none" {
			p.stream.PackageSeparator = ""
		} else {
			p.stream.PackageSeparator = directiveVal
		}
	}

	return ast.SetCommand{
		Key:   keyTok.Literal,
		Value: directiveVal,
		Trivia: ast.Trivia{
			LeadingTrivia:  leadingTrivia,
			TrailingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}, nil
}

func (p *Parser) parseContinerIdent(tok tokenizer.Token) (tokenizer.Token, error) {
	if tok.Type == tokenizer.STRING {
		return tok, nil
	}
	if tok.Type != tokenizer.IDENTIFIER {
		return tokenizer.Token{}, NewParserError("Expected identifier for container name", tok)
	}

	var sb strings.Builder
	sb.WriteString(tok.Literal)
	lastTok := tok
	for tok = p.stream.PeekTokenAt(0); tok.Type != tokenizer.EOF; tok = p.stream.PeekRawTokenAt(0) {
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
			return tokenizer.Token{}, NewParserError("Expected container name to be a single identifier", tok)
		}
		switch tok.Type {
		case tokenizer.LBRACE, tokenizer.LANGLE, tokenizer.HASH, tokenizer.NEWLINE, tokenizer.DOLLAR:
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
	}
	return "", "", NewParserError("Invalid container alias and identifier combination", lhs)
}

func (p *Parser) parseNote() (ast.Note, error) {
	// tok is a keyword 'note'
	note := ast.Note{
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}
	tok := p.stream.Emit()
	var err error
	if tok.Type == tokenizer.STRING {
		err = p.parseInlineAliasNote(&note, tok)
	} else {
		class := keyword.Classify(tok.Literal)
		switch class {
		case keyword.Direction:
			err = p.parseReltiveNote(&note, tok)
		case keyword.Position:
			err = p.parseLinkNote(&note, tok)
		case keyword.Alias:
			err = p.parseMultilineAliasNote(&note)
		default:
			return note, WrapParserError(fmt.Errorf("expected direction, string, note position or alias after 'note', got %s", class.String()), tok)
		}
	}
	if err != nil {
		return note, err
	}
	return note, nil
}

func (p *Parser) parseReltiveNote(note *ast.Note, dirTok tokenizer.Token) error {
	note.Direction = p.mapTokenToDirection(dirTok)
	if relativeTok, ok := p.stream.TryConsumeKW(keyword.Position); ok {
		target, err := p.parseTargetRef(p.stream.Emit()) // consume target
		if err != nil {
			return err
		}
		if strings.ToLower(target.Entity) != "link" && relativeTok.Literal == "on" {
			return NewParserError("Unexpected identifier for a note link target", relativeTok)
		}
		note.Target = target
	} else if tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER); ok {
		return NewParserError("Unexpected identifier after direction", tok)
	}
	p.tryParseColor()
	return p.parseNoteBody(note)
}

func (p *Parser) parseInlineAliasNote(note *ast.Note, stringTok tokenizer.Token) error {
	note.Text = stringTok.Literal
	if aliasTok, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		return NewParserError("Expected alias keyword after note text", aliasTok)
	}
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return NewParserError("Expected identifier after alias keyword", tok)
	}
	note.Alias = tok.Literal
	p.tryParseColor()
	if res := p.stream.ConsumeUntilType(tokenizer.NEWLINE); len(res) != 0 {
		return NewParserError("Unexpected tokens after inline alias note", p.stream.PeekTokenAt(0))
	}
	note.TrailingTrivia = p.stream.DumpCollectedTrivia()
	return nil
}

func (p *Parser) parseMultilineAliasNote(note *ast.Note) error {
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return NewParserError("Expected identifier after alias keyword", tok)
	}
	p.tryParseColor()
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		return NewParserError("Expected newline after alias keyword", tok)
	}
	return p.parseNoteBody(note)
}

func (p *Parser) parseLinkNote(note *ast.Note, onTok tokenizer.Token) error {
	if onTok.Literal != "on" {
		return NewParserError("Unexpected identifier after 'note'", onTok)
	}
	if _, ok := p.stream.TryConsume(amb(tokenizer.IDENTIFIER, "link")); !ok {
		return NewParserError("Expected 'link' after 'note on'", onTok)
	}
	return p.parseNoteBody(note)
}

func (p *Parser) tryParseColor() string {
	if _, ok := p.stream.TryConsumeType(tokenizer.HASH); !ok {
		return ""
	}
	tokens := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.COLON, tokenizer.LBRACE)
	return p.stream.TokensToString(tokens)
}

func (p *Parser) parseNoteBody(note *ast.Note) error {
	tok := p.stream.Emit()

	noteEndSequence := []tokenizer.Token{unamb(tokenizer.NEWLINE)}
	switch tok.Type {
	case tokenizer.COLON:
		// to not error out
	case tokenizer.NEWLINE:
		noteEndSequence = []tokenizer.Token{amb(tokenizer.IDENTIFIER, "end"), amb(tokenizer.IDENTIFIER, "note")}
		note.TrailingTrivia = p.stream.DumpCollectedTrivia()
	default:
		return NewParserError("Expected ':' or newline after note definition", tok)
	}

	body, err := p.stream.ConsumeTextBlock(noteEndSequence)
	if err != nil {
		return err
	}

	if len(note.TrailingTrivia) == 0 {
		note.TrailingTrivia = p.stream.DumpCollectedTrivia()
	} else {
		if closingTrivia := p.stream.DumpCollectedTrivia(); len(closingTrivia) > 0 {
			note.TrailingTrivia = append(note.TrailingTrivia, closingTrivia...)
		}
	}
	note.Text = strings.TrimSuffix(body, "\n")
	return nil
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

func (p *Parser) parseSkinparamBlock(selectors []string, leadingTrivia []tokenizer.Token) ([]*ast.StyleRule, error) {
	if _, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		return nil, NewParserError("Expected opening brace after skinparam target", p.stream.PeekTokenAt(0))
	}

	currentRule := &ast.StyleRule{
		Selectors:   slices.Clone(selectors),
		Properties:  make(map[string]string),
		IsSkinparam: true,
		Trivia: ast.Trivia{
			LeadingTrivia: leadingTrivia,
		},
	}
	var rules []*ast.StyleRule

	for tok := p.stream.Emit(); tok.Type != tokenizer.RBRACE; tok = p.stream.Emit() {
		if tok.Type == tokenizer.NEWLINE {
			continue
		}

		if tok.Type == tokenizer.EOF {
			return nil, NewParserError("Unexpected EOF in skinparam block", tok)
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

			subRules, err := p.parseSkinparamBlock(subSelectors, p.stream.DumpCollectedTrivia())
			if err != nil {
				return nil, err
			}
			rules = append(rules, subRules...)
		} else {
			// Inline value
			value := p.stream.ReadUntilNewline()
			currentRule.TrailingTrivia = p.stream.DumpCollectedTrivia()
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
		return nil, NewParserError("Expected </style> closing tag", p.stream.PeekTokenAt(0))
	}

	// Consume '</style>'
	for range 4 {
		p.stream.Emit()
	}

	if res := p.stream.ConsumeUntilType(tokenizer.NEWLINE); len(res) != 0 {
		return rules, NewParserError("Unexpected tokens after style block", p.stream.PeekTokenAt(0))
	}
	rules[len(rules)-1].(*ast.StyleRule).TrailingTrivia = p.stream.DumpCollectedTrivia()
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
	leadingTrivia := p.stream.DumpCollectedTrivia()
	if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); ok {
		// Multi-line title block ending with 'end title'
		titleEndSequence := []tokenizer.Token{
			unamb(tokenizer.NEWLINE),
			amb(tokenizer.IDENTIFIER, "end"),
			amb(tokenizer.IDENTIFIER, "title"),
		}
		var err error
		p.ast.Title, err = p.stream.ConsumeTextBlock(titleEndSequence)
		return ast.TitleDef{
			Text: p.ast.Title,
			Trivia: ast.Trivia{
				LeadingTrivia:  leadingTrivia,
				TrailingTrivia: p.stream.DumpCollectedTrivia(),
			},
		}, err
	}

	// Single-line title
	p.ast.Title = p.stream.ReadRawUntilNewline()
	return ast.TitleDef{Text: p.ast.Title}, nil
}

func (p *Parser) parseTargetRef(firstTok tokenizer.Token) (ast.TargetRef, error) {
	var ref ast.TargetRef
	segments := []string{firstTok.Literal}

	for {
		if _, ok := p.stream.TryConsumePackageSeparator(); ok {
			tok := p.stream.Emit()
			if tok.Type != tokenizer.IDENTIFIER && tok.Type != tokenizer.STRING {
				return ref, NewParserError("Expected identifier or string after package separator in target ref", tok)
			}
			segments = append(segments, tok.Literal)
			continue
		}
		break
	}

	if len(segments) > 0 {
		ref.Entity = segments[len(segments)-1]
		if len(segments) > 1 {
			ref.PackagePath = segments[:len(segments)-1]
		}
	}

	if p.stream.AssertTypeSeq([]tokenizer.TokenType{tokenizer.COLON, tokenizer.COLON}) {
		p.stream.Emit() // consume first :
		p.stream.Emit() // consume second :

		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.STRING:
			ref.Member = tok.Literal
		case tokenizer.IDENTIFIER:
			var sb strings.Builder
			sb.WriteString(tok.Literal)
			if p.stream.AssertType(tokenizer.LPAREN) {
				p.stream.Emit() // consume (
				sb.WriteString("(")
				paramToks := p.stream.ConsumeUntilType(tokenizer.RPAREN, tokenizer.NEWLINE, tokenizer.EOF)
				sb.WriteString(p.stream.TokensToString(paramToks))
				if _, ok := p.stream.TryConsumeType(tokenizer.RPAREN); !ok {
					return ref, NewParserError("Expected ')' closing method signature in target ref", p.stream.PeekTokenAt(0))
				}
				sb.WriteString(")")
			}
			ref.Member = sb.String()
		default:
			return ref, NewParserError("Expected member identifier or string after '::'", tok)
		}
	}

	return ref, nil
}

func (p *Parser) parseRelationship(firstTargetTok tokenizer.Token) (ast.Relationship, error) {
	// Entry token is supposedly the first identifier
	var err error
	var rel ast.Relationship
	rel.LeadingTrivia = p.stream.DumpCollectedTrivia()
	rel.LHS, err = p.parseTargetRef(firstTargetTok)
	if err != nil {
		return rel, err
	}

	if multTok, ok := p.stream.TryConsumeType(tokenizer.STRING); ok {
		rel.MultLHS, err = ast.ParseCardinality(multTok.Literal)
		if err != nil {
			return rel, WrapParserError(err, multTok)
		}
	}

	p.parseArrowTokens(&rel)

	if multTok, ok := p.stream.TryConsumeType(tokenizer.STRING); ok {
		rel.MultRHS, err = ast.ParseCardinality(multTok.Literal)
		if err != nil {
			return rel, WrapParserError(err, multTok)
		}
	}

	if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.STRING) {
		return rel, NewParserError("Expected identifier or string after relationship", p.stream.PeekTokenAt(0))
	}
	rel.RHS, err = p.parseTargetRef(p.stream.Emit())
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
		rel.Label = p.stream.ReadRawUntilNewline()
	default:
		return rel, NewParserError("Expected newline or colon after relationship", endingToken)
	}

	rel.TrailingTrivia = p.stream.DumpCollectedTrivia()
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
			return NewParserError("Unexpected identifier in relationship definition", tok)
		}
		fallthrough
	case tokenizer.HASH, tokenizer.ASTERISK, tokenizer.PLUS, tokenizer.CARET:
		rel.LArrow = rune(tok.Literal[0])
	case tokenizer.DOT, tokenizer.DASH: // so that encountering them doesn't cause an error
	default:
		return NewParserError("Unexpected token at the start of relationship definition", tok)
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
		return NewParserError("Unexpected token as the relationship body", tok)
	}
	var ok bool
	rel.Body = rune(tok.Literal[0])
	for tok.Type != tokenizer.EOF && tok.Type != tokenizer.NEWLINE {
		if tok, ok = p.stream.TryConsumeType(bodyTokType); ok {
			continue
		} else if tok, ok = p.stream.TryConsumeType(oppositeBodyTokType); ok {
			// Simply convenient error message
			return NewParserError("Different body type runes in relationship", tok)
		}
		if !p.stream.AssertType(tokenizer.LBRACKET) && !p.stream.AssertKW(keyword.Direction) {
			break
		} else if sawAttrs || sawDirection {
			return NewParserError("Cannot separate direction and attributes with a body token", tok)
		}

		if isLolipop {
			return NewParserError("Lolipop interface cannot contain attributes or direction", tok)
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
						return NewParserError("Unexpected direction in relationship", tok)
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
							return NewParserError("Unexpected break in relationship attribute container", tok)
						case tokenizer.COMMA:
							if attrSB.Len() == 0 {
								return NewParserError("Unexpected comma in relationship attribute container", tok)
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
			return NewParserError("Unexpected token in body relationship definition", tok)
		}
	}

	tok = p.stream.PeekTokenAt(0)
	switch tok.Type {
	case tokenizer.PIPE:
		// should only be encountered on --|> case as the end of the relationship
		p.stream.Emit()
		if _, ok := p.stream.TryConsumeType(tokenizer.RANGLE); !ok {
			return NewParserError("Expected '|>' after relationship", tok)
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
			return NewParserError("Lolipop interface cannot contain direction or attributes", tok)
		}
		if isLolipop {
			return NewParserError("Cannot have double headed lolipop relationship", tok)
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
			return NewParserError("Unexpected identifier in relationship definition", tok)
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
		return NewParserError("Missing body in the relationship", tok)
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

func (p *Parser) parseInlineMember(firstTok tokenizer.Token) (ast.Statement, error) {
	leadingTrivia := p.stream.DumpCollectedTrivia()

	targetRef, err := p.parseTargetRef(firstTok)
	if err != nil {
		return nil, err
	}

	if _, ok := p.stream.TryConsumeType(tokenizer.COLON); !ok {
		return nil, NewParserError("Expected ':' after entity identifier for inline member declaration", p.stream.PeekTokenAt(0))
	}

	entryTok := p.stream.PeekTokenAt(0)
	vis := ast.VisibilityUnknown
	var mod *string = nil
	switch entryTok.Type {
	case tokenizer.DASH, tokenizer.TILDE, tokenizer.HASH, tokenizer.PLUS:
		p.stream.Emit()
		vis = p.mapTokenToVisibility(entryTok.Type)
	case tokenizer.LBRACE:
		m, err := p.tryReadModifier()
		if err != nil {
			return nil, err
		}
		mod = &m
	case tokenizer.IDENTIFIER:
		p.stream.Emit()
	}
	member, err := p.parseFieldOrMethod(mod, vis, entryTok, leadingTrivia)
	if err != nil {
		return nil, err
	}

	ent := ast.Entity{
		Identifier: targetRef.Entity,
		Kind:       ast.EntityUnknown,
		Members:    []ast.Member{member},
	}

	return wrapInContainers(ent, targetRef.PackagePath), nil
}
