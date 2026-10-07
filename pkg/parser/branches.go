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
	mark := p.Mark(tok)
	cmd := ast.VisibilityCommand{
		Kind: ast.VisibilityCMDUnknown,
		BaseNode: ast.BaseNode{
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
	cmd.NodeSpan = p.Span(mark)
	cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
	return cmd, nil
}

func (p *Parser) parseDiagDirection(tok tokenizer.Token) (ast.DirectionCommand, error) {
	mark := p.Mark(tok)
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
		BaseNode: ast.BaseNode{
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
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		return cmd, NewParserError("Unexpected tokens after direction command", p.stream.PeekTokenAt(0))
	}
	cmd.NodeSpan = p.Span(mark)
	cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
	return cmd, nil
}

func (p *Parser) parseDirective(tok1 tokenizer.Token) (ast.Statement, error) {
	startMark := p.Mark(tok1)
	directiveNameTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return nil, NewParserError("Expected directive name", p.stream.PeekTokenAt(0))
	}

	if directiveNameTok.Span.Start.Offset != tok1.Span.Start.Offset+1 {
		return nil, NewParserError("Expected directive name right after !", directiveNameTok)
	}

	if strings.HasPrefix(directiveNameTok.Literal, "include") {
		dir, err := p.parseIncludeDirective(directiveNameTok)
		dir.NodeSpan = p.Span(startMark)
		return dir, err
	} else {
		return p.parseUnhandledDirective(tok1, directiveNameTok)
	}
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

	mark := p.Mark(tok)
	dir := ast.IncludeDirective{
		Kind: kind,
		BaseNode: ast.BaseNode{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}

	filePathToks := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.EXCLAMATION)
	dir.Path = p.stream.TokensToString(filePathToks)
	if p.stream.AssertType(tokenizer.EXCLAMATION) {
		p.stream.Emit() // consume '!'
		// Id or order must be a single token
		lastDirTok := p.stream.Emit()
		dir.Tag = lastDirTok.Literal
		dir.NodeSpan = p.SpanTo(mark, lastDirTok)
	} else {
		dir.NodeSpan = p.SpanTo(mark, filePathToks[len(filePathToks)-1])
	}

	p.stream.EmitCommentToks()

	dir.TrailingTrivia = p.stream.DumpCollectedTrivia()

	return dir, nil
}

func (p *Parser) extractScaleTokens() (lhs, sep, rhs, unit string, err error) {
	return "", "", "", "", nil
}

func isDecimalInt(s string) bool {
	return strings.ContainsAny(s, ".exob")
}

func isDecimalFloat(s string) bool {
	return strings.ContainsAny(s, "exob")
}

func (p *Parser) parseScaleTrial(startTok tokenizer.Token) (ast.ScaleCommand, error) {
	mark := p.Mark(startTok)
	cmd := ast.ScaleCommand{
		BaseNode: ast.BaseNode{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}
	if _, ok := p.stream.TryConsume(tokenizer.Token{Type: tokenizer.IDENTIFIER, Literal: "max"}); ok {
		cmd.IsMax = true
	}

	op1, sep, op2, unit, err := p.extractScaleTokens()
	if err != nil {
		return cmd, err
	}

	// Step 2: Validate against State Machine rules
	isOp1Int := isDecimalInt(op1)
	isOp1Dec := isDecimalFloat(op1)

	if !isOp1Int && !isOp1Dec {
		return cmd, NewParserError("Expected number after scale", startTok)
	}

	// Max constraints
	if cmd.IsMax {
		if isOp1Dec {
			return cmd, NewParserError("Cannot use decimals with 'max'", startTok)
		}
		if sep == "/" {
			return cmd, NewParserError("Cannot use fractions with 'max'", startTok)
		}
		if sep == "" && unit == "" {
			return cmd, NewParserError("Cannot use numbers with 'max' without a unit or box", startTok)
		}
	}

	// Decimal constraints
	if isOp1Dec && sep != "" {
		return cmd, NewParserError("Cannot use decimals with separators ('*', 'x', '/')", startTok)
	}

	// Binary separator constraints (*, x, /)
	if sep != "" {
		if unit != "" {
			return cmd, NewParserError("Cannot specify unit with a box or fraction", startTok)
		}
		if !isDecimalInt(op2) {
			return cmd, NewParserError("Expected integer after separator", startTok)
		}
	}

	// Step 3: Populate AST
	if sep == "." { // e.g. "1.5" factor
		parts := strings.Split(op1, ".")
		cmd.Lhs, cmd.Sep, cmd.Rhs = parts[0], ".", parts[1]
	} else {
		cmd.Lhs, cmd.Sep, cmd.Rhs, cmd.Unit = op1, sep, op2, unit
	}

	cmd.NodeSpan = p.Span(mark)
	p.stream.EmitCommentToks()
	cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
	return cmd, nil
}

func (p *Parser) parseScale(startTok tokenizer.Token) (ast.ScaleCommand, error) {
	mark := p.Mark(startTok)
	cmd := ast.ScaleCommand{
		BaseNode: ast.BaseNode{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}
	if _, ok := p.stream.TryConsume(tokenizer.Token{Type: tokenizer.IDENTIFIER, Literal: "max"}); ok {
		cmd.IsMax = true
	}

	divideSingleTok := func(lit, sep string) error {
		if !strings.Contains(lit, sep) {
			return fmt.Errorf("expected number after scale")
		}
		parts := strings.Split(lit, sep)
		if len(parts) != 2 {
			return fmt.Errorf("expected two parts after '%s'", sep)
		}
		_, errW := strconv.Atoi(parts[0])
		_, errH := strconv.Atoi(parts[1])
		if errW != nil || errH != nil {
			return fmt.Errorf("expected width and height to be integers")
		}
		cmd.Lhs = parts[0]
		cmd.Rhs = parts[1]
		cmd.Sep = sep
		return nil
	}
	var isInt, isFloat bool

	tok := p.stream.Emit()
	switch tok.Type {
	case tokenizer.NUMBER:
		if strings.ContainsAny(tok.Literal, "exob") {
			return cmd, NewParserError("Cannot use non-decimal or scientific integer for scale", tok)
		}
	case tokenizer.DOT:
		// .5 case
		numTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return cmd, NewParserError("Expected number after scale leading dot", numTok)
		}
		if strings.ContainsAny(numTok.Literal, ".exob") {
			return cmd, NewParserError("Cannot use non-decimal integer for shorthand float", numTok)
		}
		cmd.Lhs = "." + numTok.Literal
		if !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
			return cmd, NewParserError("Expected scale to end after fractional scale", p.stream.PeekTokenAt(0))
		}
		if cmd.IsMax {
			return cmd, NewParserError("Cannot use fractions with 'max'", tok)
		}
		cmd.NodeSpan = p.Span(mark)
		p.stream.EmitCommentToks()
		cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
		return cmd, nil
	case tokenizer.IDENTIFIER:
		// Try to handle 200x300 which is tokenized as an IDENTIFIER
		if strings.Contains(tok.Literal, "x") && !strings.HasSuffix(tok.Literal, "x") && !strings.HasPrefix(tok.Literal, "x") {
			if err := divideSingleTok(tok.Literal, "x"); err != nil {
				return cmd, WrapParserError(err, tok)
			}
			cmd.NodeSpan = p.Span(mark)
			p.stream.EmitCommentToks()
			cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
			return cmd, nil
		}

		if before, ok := strings.CutSuffix(tok.Literal, "x"); ok {
			widthStr := before
			if widthStr == "" || strings.ContainsAny(widthStr, ".exob") {
				return cmd, NewParserError("Cannot use non-decimal integer for width of a box", tok)
			}
			if _, err := strconv.Atoi(widthStr); err != nil {
				return cmd, NewParserError("Expected width in a box to be an integer", tok)
			}

			heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
			if !ok {
				return cmd, NewParserError("Expected height after 'x' in a box",
					p.stream.PeekTokenAt(0))
			}
			if strings.ContainsAny(heightTok.Literal, ".exob") {
				return cmd, NewParserError("Cannot use non-decimal integer for height of a box",
					heightTok)
			}
			if _, err := strconv.Atoi(heightTok.Literal); err != nil {
				return cmd, NewParserError("Invalid height for the box", heightTok)
			}

			cmd.Lhs = widthStr
			cmd.Sep = "x"
			cmd.Rhs = heightTok.Literal
			cmd.NodeSpan = p.Span(mark)
			p.stream.EmitCommentToks()
			cmd.TrailingTrivia = p.stream.DumpCollectedTrivia()
			return cmd, nil
		}

		return cmd, NewParserError("Expected number after scale", tok)
	default:
		return cmd, NewParserError("Expected number after scale", p.stream.PeekTokenAt(0))
	}

	valueLit := tok.Literal
	valueTok := tok
	if valueFloat, err := strconv.ParseFloat(valueLit, 64); err == nil {
		isFloat = true
		_, isInt = toInteger(valueFloat)
	}

	setBox := func(sep string) error {
		if !isInt {
			return NewParserError("Expected width in a box to be an integer", tok)
		}
		if strings.ContainsAny(valueLit, ".exob") {
			return NewParserError("Cannot use non-decimal integer for width of a box", valueTok)
		}
		cmd.Lhs = valueLit
		cmd.Sep = sep
		heightTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
		if !ok {
			return NewParserError(fmt.Sprintf("Expected height after '%s' in a box", sep), p.stream.PeekTokenAt(0))
		}
		if strings.ContainsAny(heightTok.Literal, ".exob") {
			return NewParserError("Cannot use non-decimal integer for height of a box", heightTok)
		}
		if _, err := strconv.Atoi(heightTok.Literal); err != nil {
			return NewParserError("Invalid height for the box", heightTok)
		}
		cmd.Rhs = heightTok.Literal
		return nil
	}

	tok = p.stream.Emit()
	switch tok.Type {
	case tokenizer.IDENTIFIER:
		switch tok.Literal {
		case "width":
			if strings.ContainsAny(valueLit, "exob") {
				return cmd, NewParserError("Cannot use non-decimal or scientific integer for width of a box", valueTok)
			}
			cmd.Lhs = valueLit
			cmd.Unit = tok.Literal
		case "height":
			if strings.ContainsAny(valueLit, "exob") {
				return cmd, NewParserError("Cannot use non-decimal or scientific integer for height of a box", valueTok)
			}
			cmd.Lhs = valueLit
			cmd.Unit = tok.Literal
		case "x":
			if err := setBox("x"); err != nil {
				return cmd, err
			}
		default:
			after, ok := strings.CutPrefix(tok.Literal, "x")
			if !ok {
				return cmd, NewParserError(fmt.Sprintf("Unexpected identifier: %s", tok.Literal), tok)
			}
			heightStr := after
			if heightStr == "" || strings.ContainsAny(heightStr, ".exob") {
				return cmd, NewParserError("Cannot use non-decimal integer for height of a box", tok)
			}
			if _, err := strconv.Atoi(heightStr); err != nil {
				return cmd, NewParserError("Invalid height for the box", tok)
			}
			if !isInt || strings.ContainsAny(valueLit, ".exob") {
				return cmd, NewParserError("Expected width in a box to be an integer", valueTok)
			}

			cmd.Lhs = valueLit
			cmd.Sep = "x"
			cmd.Rhs = heightStr
		}
	case tokenizer.ASTERISK:
		if err := setBox("*"); err != nil {
			return cmd, err
		}
	case tokenizer.SLASH:
		if err := setBox("/"); err != nil {
			return cmd, err
		}
	case tokenizer.NEWLINE, tokenizer.EOF:
		// EOF handling is a special case for a test case
		// valueLit is either a bare integer or a float
		if !isFloat {
			return cmd, NewParserError("Expected number after scale", tok)
		}

		if cmd.IsMax {
			return cmd, NewParserError("Cannot use numbers with 'max' without a unit (width/height) or a box", tok)
		}

		if isFloat && !isInt {
			if err := divideSingleTok(valueLit, "."); err != nil {
				return cmd, WrapParserError(err, tok)
			}
		} else {
			cmd.Lhs = valueLit
		}
	default:
		return cmd, NewParserError(fmt.Sprintf("Unexpected token: %s(%s)", tok.Type.String(), tok.Literal), tok)
	}

	if !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
		return cmd, NewParserError("Unexpected tokens after scale command", p.stream.PeekTokenAt(0))
	}
	cmd.NodeSpan = p.Span(mark)
	p.stream.EmitCommentToks()
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
	m := p.Mark(tok)
	ent := &ast.Entity{
		Kind: p.mapTokenToEntityKind(tok),
		BaseNode: ast.BaseNode{
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

	if tok, ok := p.stream.TryConsumeType(tokenizer.LBRACE); !ok {
		// No body return entity as is
		ent.NodeSpan = p.SpanTo(m, tok)
		p.stream.EmitCommentToks()
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

	ent.NodeSpan = p.Span(m)

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
	m := p.Mark(entryTok)

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

	span := p.Span(m)

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
		MemberSpan:     span,
	}
	if isMethod {
		return p.Dialect.ParseMethod(entry, &opts)
	} else {
		return p.Dialect.ParseField(entry, &opts)
	}
}

func (p *Parser) parseContainer(tok tokenizer.Token) (ast.Container, error) {
	m := p.Mark(tok)
	containerClass := keyword.Classify(tok.Literal)
	container := ast.Container{
		Kind: p.mapKeywordToContainerKind(containerClass),
		BaseNode: ast.BaseNode{
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
		stmt, err := p.parseContainerStatement(tok)
		if err != nil {
			return container, err
		}
		if stmt == nil {
			return container, NewParserError("Expected a statement in a container body", tok)
		}
		container.Statements = append(container.Statements, stmt)
	}

	container.NodeSpan = p.Span(m)

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

func (p *Parser) parseSetDirective(startTok tokenizer.Token) (ast.Statement, error) {
	mark := p.Mark(startTok)
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
		BaseNode: ast.BaseNode{
			NodeSpan:       p.Span(mark),
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

func (p *Parser) parseNote(startTok tokenizer.Token) (ast.Note, error) {
	mark := p.Mark(startTok)
	// tok is a keyword 'note'
	note := ast.Note{
		BaseNode: ast.BaseNode{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
	}
	tok := p.stream.Emit()
	var err error
	if tok.Type == tokenizer.STRING {
		err = p.parseInlineIdentNote(&note, tok)
	} else {
		class := keyword.Classify(tok.Literal)
		switch class {
		case keyword.Direction:
			err = p.parseRelativeNote(&note, tok)
		case keyword.Position:
			err = p.parseLinkNote(&note, tok)
		case keyword.Alias:
			err = p.parseMultilineAliasNote(&note)
		default:
			return note, WrapParserError(fmt.Errorf("expected direction, string, note position or alias after 'note', got %s", class.String()), tok)
		}
	}
	note.NodeSpan = p.Span(mark)
	p.stream.EmitCommentToks()
	closingTrivia := p.stream.DumpCollectedTrivia()
	note.TrailingTrivia = append(note.TrailingTrivia, closingTrivia...)
	if err != nil {
		return note, err
	}
	return note, nil
}

func (p *Parser) tryParseColor() string {
	if _, ok := p.stream.TryConsumeType(tokenizer.HASH); !ok {
		return ""
	}
	tokens := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.COLON, tokenizer.LBRACE)
	return p.stream.TokensToString(tokens)
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

func (p *Parser) parseTargetRef(firstTok tokenizer.Token) (ast.TargetRef, error) {
	var ref ast.TargetRef
	segments := []string{firstTok.Literal}

	for {
		length, ok := p.stream.AssertPackageSeparator()
		if !ok {
			break
		}

		afterPkgSep := p.stream.PeekTokenAt(length)
		if afterPkgSep.Type != tokenizer.IDENTIFIER && afterPkgSep.Type != tokenizer.STRING {
			// We have hit the case where the separator token is not a part of the entity reference
			break
		}

		for range length {
			p.stream.Emit()
		}

		segments = append(segments, p.stream.Emit().Literal)
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
	mark := p.Mark(firstTargetTok)
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

	rel.NodeSpan = p.Span(mark)

	endingToken := p.stream.PeekTokenAt(0)
	switch endingToken.Type {
	case tokenizer.NEWLINE, tokenizer.EOF:
		p.stream.Emit()
	case tokenizer.RBRACE:
		// Do not consume RBRACE so enclosing container block loop sees it
		break
	case tokenizer.COLON:
		p.stream.Emit()
		rel.Label = p.stream.ReadUntilNewline()
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
	case tokenizer.ASTERISK:
		rel.TypeRHS = ast.RelationComposition
		p.stream.Emit()
		rel.RArrow = rune(tok.Literal[0])
	case tokenizer.IDENTIFIER:
		// special case for 'x' and 'o' in relationship
		switch tok.Literal {
		case "x":
			p.stream.Emit()
			rel.RArrow = rune(tok.Literal[0])
		case "o":
			if rel.Body == '-' {
				rel.TypeRHS = ast.RelationAggregation
			}
			p.stream.Emit()
			rel.RArrow = rune(tok.Literal[0])
		default:
			return NewParserError("Unexpected identifier in relationship definition", tok)
		}
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
		if rTok.Literal == "x" || rTok.Literal == "o" {
			idx++
			hasRightArrowhead = true
		}
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

func parseLayoutAlignmentToken(tok tokenizer.Token, blockKind ast.TextBlockKind) (horiz string, vert string, isAlign bool) {
	switch strings.ToLower(tok.Literal) {
	case "left", "right", "center":
		return strings.ToLower(tok.Literal), "", true
	case "top", "bottom":
		if blockKind != ast.BlockLegend {
			return "", "", false
		}
		return "", strings.ToLower(tok.Literal), true
	default:
		return "", "", false
	}
}

func (p *Parser) parseLayoutStatement(kwTok tokenizer.Token, prefixAlignment *tokenizer.Token) (ast.Statement, error) {
	mark := p.Mark(kwTok)
	blockKind := mapKwTokToTextBlockKind(keyword.Classify(kwTok.Literal))
	leadingTrivia := p.stream.DumpCollectedTrivia()

	block := ast.TextBlock{
		Kind: blockKind,
		BaseNode: ast.BaseNode{
			LeadingTrivia: leadingTrivia,
		},
	}

	// Process prefix alignment if provided (e.g. "left header", "center footer")
	if prefixAlignment != nil {
		if blockKind == ast.BlockTitle {
			return nil, NewParserError("Title alignment not supported", *prefixAlignment)
		}
		h, v, isAlign := parseLayoutAlignmentToken(*prefixAlignment, blockKind)
		if !isAlign {
			if strings.EqualFold(prefixAlignment.Literal, "top") || strings.EqualFold(prefixAlignment.Literal, "bottom") {
				return nil, NewParserError("Vertical alignment only supported for legend", *prefixAlignment)
			}
			return nil, NewParserError("Invalid alignment", *prefixAlignment)
		}
		block.HorizontalAlignment = h
		block.VerticalAlignment = v
	}

	// Consume any trailing alignment modifiers on the opener line (e.g. "legend top left", "header center")
	for !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
		peekTok := p.stream.PeekTokenAt(0)
		h, v, isAlign := parseLayoutAlignmentToken(peekTok, blockKind)
		if !isAlign {
			break
		}
		if blockKind == ast.BlockTitle {
			break // For title, any tokens on the same line are title text
		}
		if h != "" {
			if block.HorizontalAlignment != "" {
				return nil, NewParserError("Horizontal alignment already set", peekTok)
			}
			block.HorizontalAlignment = h
			p.stream.Emit()
		} else if v != "" {
			if block.VerticalAlignment != "" {
				return nil, NewParserError("Vertical alignment already set", peekTok)
			}
			if blockKind != ast.BlockLegend {
				return nil, NewParserError("Vertical alignment only supported for legend", peekTok)
			}
			block.VerticalAlignment = v
			p.stream.Emit()
		}
	}

	lineContentToks := p.stream.ConsumeUntilType(tokenizer.NEWLINE)

	var body string
	if len(lineContentToks) == 0 {
		// --- MULTI-LINE BLOCK FORM ---
		var err error
		body, err = p.stream.ConsumeTextBlock("end", kwTok.Literal)
		if err != nil {
			if errors.Is(err, tokenizer.ErrScopeDelimiterHit) {
				return nil, NewParserError(fmt.Sprintf("unterminated block statement for %s (hit enclosing scope delimiter)", kwTok.Literal), kwTok)
			}
			return nil, NewParserError(fmt.Sprintf("unterminated block statement for %s", kwTok.Literal), kwTok)
		}
	} else {
		// --- SINGLE-LINE INLINE FORM ---
		// We must consume newline as per ConsumeUntilType contract
		p.stream.MustConsumeType(tokenizer.NEWLINE)

		body = p.stream.SliceInputEnclosingTokens(lineContentToks...)
	}

	block.NodeSpan = p.Span(mark)

	block.Text = body
	p.stream.EmitCommentToks()
	block.TrailingTrivia = p.stream.DumpCollectedTrivia()

	if blockKind == ast.BlockTitle {
		p.ast.Title = block.Text
	}
	return block, nil
}
