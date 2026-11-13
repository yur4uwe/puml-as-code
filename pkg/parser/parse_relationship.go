package parser

import (
	"errors"
	"fmt"
	"strconv"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"
)

func (p *Parser) parseRelationshipStyle() (*ast.RelationshipStyle, tokenizer.SourceSpan, error) {
	relStyle := &ast.RelationshipStyle{}

	tokenAfterHash := p.stream.PeekTokenAt(0)
	if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.NUMBER) {
		return relStyle, tokenAfterHash.Span, fmt.Errorf("unexpected token in style declaration: %s(%s)", tokenAfterHash.Type, tokenAfterHash.Literal)
	}

	// --- First item: may be a bare color ---
	switch p.stream.PeekTokenAt(0).Literal {
	case "back", "line", "text":
	default:
		// lines only allow simple colors
		tok := p.stream.Emit()
		if tok.Type != tokenizer.IDENTIFIER && tok.Type != tokenizer.NUMBER {
			return relStyle, tok.Span, fmt.Errorf("expected identifier or a number for a line color value, got: %s(%s)", tok.Type, tok.Literal)
		}
		relStyle.LineColor = ast.FromToken(tok)

		if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); !ok {
			return relStyle, tokenizer.SourceSpan{}, nil
		}
	}

	for {
		tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
		if !ok {
			return relStyle, tok.Span, fmt.Errorf("unexpected token describing relationship colorable target: %s(%s)", tok.Type, tok.Literal)
		}

		targetKind := ""
		switch tok.Literal {
		case "line":
			targetKind = tok.Literal
			if _, ok = p.stream.TryConsumeType(tokenizer.DOT); !ok {
				break // no dot — fall through to colon:color
			}

			lineStyleTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
			if !ok {
				return relStyle, lineStyleTok.Span, fmt.Errorf("unexpected token describing line style: %s(%s)", lineStyleTok.Type, lineStyleTok.Literal)
			}

			ls, err := parseLineStyle(lineStyleTok)
			if err != nil {
				return relStyle, lineStyleTok.Span, err
			}
			relStyle.LineStyle = ls

			if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); ok {
				continue
			}
			// no semicolon, so fall through to parse optional :color
		case "text":
			targetKind = tok.Literal
		default:
			return relStyle, tok.Span, fmt.Errorf("unexpected value for color target specifier: %s", tok.Literal)
		}

		if _, ok := p.stream.TryConsumeType(tokenizer.COLON); !ok {
			peekTok := p.stream.PeekTokenAt(0)
			return relStyle, peekTok.Span, errors.New("missing colon after specified color target")
		}

		colorTok := p.stream.Emit()
		expr, err := p.parseColorExpr(colorTok)
		if err != nil {
			return relStyle, p.stream.LastSemanticToken().Span, err
		}

		switch targetKind {
		case "text":
			relStyle.TextColor = expr
		case "line":
			relStyle.LineColor = expr
		}

		if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); !ok {
			break
		}
	}

	return relStyle, tokenizer.SourceSpan{}, nil
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

func (p *Parser) parseRelationship(firstTargetTok tokenizer.Token) *ast.Relationship {
	var pe ParserError
	// Entry token is supposedly the first identifier
	var err error
	rel := &ast.Relationship{}
	// Must not fail, guaranteed by the disambiguateIdent check
	rel.LHS, err = p.parseTargetRef(firstTargetTok)
	if errors.As(err, &pe) {
		p.ast.RecordError(err, pe.Span, rel)
		p.syncToLineEnd()
		return rel
	}

	if multTok, ok := p.stream.TryConsumeType(tokenizer.STRING); ok {
		// I think its productive to defer the cardinality parsing to the resolver
		// or someone that actually cares about it
		rel.MultLHS, err = ast.ParseCardinality(multTok.Literal)
		if err != nil {
			infoSpan := multTok.Span
			err := fmt.Errorf("incorrect multiplicity syntax: %w", err)
			p.ast.RecordInfo(err, infoSpan, rel)
		}
	}

	arrowMark := p.Mark(p.stream.PeekTokenAt(0))

	err = p.parseArrowTokens(rel)
	if errors.As(err, &pe) {
		p.ast.RecordError(err, pe.Span, rel)
		p.syncToLineEnd()
		rel.ArrowSpan = p.Span(arrowMark)
		return rel
	}

	rel.ArrowSpan = p.Span(arrowMark)

	if multTok, ok := p.stream.TryConsumeType(tokenizer.STRING); ok {
		rel.MultRHS, err = ast.ParseCardinality(multTok.Literal)
		if err != nil {
			infoSpan := multTok.Span
			err := fmt.Errorf("incorrect multiplicity syntax: %w", err)
			err = WrapParserError(multTok, err)
			p.ast.RecordInfo(err, infoSpan, rel)
		}
	}

	if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.STRING) {
		nextTok := p.stream.PeekTokenAt(0)
		errSpan := nextTok.Span
		err := NewParserError(nextTok, "Expected identifier or string after relationship")
		p.ast.RecordError(err, errSpan, rel)
		p.syncToLineEnd()
		return rel
	}

	rel.RHS, err = p.parseTargetRef(p.stream.Emit())
	if errors.As(err, &pe) {
		p.ast.RecordError(err, pe.Span, rel)
		p.syncToLineEnd()
		return rel
	}

	if hashTok, ok := p.stream.TryConsumeType(tokenizer.HASH); ok {
		mark := p.Mark(hashTok)
		style, errSpan, err := p.parseRelationshipStyle()
		style.Span = p.Span(mark)
		rel.Style = style
		if err != nil {
			p.ast.RecordError(err, errSpan, rel)
			p.syncToLineEnd()
			return rel
		}
	}

	endingToken := p.stream.PeekTokenAt(0)
	switch endingToken.Type {
	case tokenizer.COLON:
		p.stream.Emit()
		rel.Label = p.stream.ReadUntilNewline()
	case tokenizer.NEWLINE, tokenizer.EOF, tokenizer.RBRACE:
		// Do not consume newline, EOF, or RBRACE
		// outer loop/caller handles the delimiter
	default:
		// literal junk, relationship is fully parsed and we can't know that is
		p.ast.RecordError(NewParserError(endingToken, "Expected newline or colon after relationship"), endingToken.Span, rel)
		p.syncToLineEnd()
	}

	return rel
}

func (p *Parser) parseArrowAttributes() (*ast.RelationshipStyle, error) {
	relStyle := &ast.RelationshipStyle{}
	for {
		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.EOF, tokenizer.NEWLINE:
			return relStyle, NewParserError(tok, "unexpected line terminator while parsing arrow attributes")
		case tokenizer.RBRACKET:
			return relStyle, nil
		case tokenizer.COMMA:
			return relStyle, NewParserError(tok, "unexpected comma in arrow attributes")
		case tokenizer.HASH:
			tok := p.stream.Emit()
			if tok.Type != tokenizer.IDENTIFIER && tok.Type != tokenizer.NUMBER {
				return relStyle, NewParserErrorf(tok, "expected color token in arrow attribute, got: %s(%s)", tok.Type, tok.Literal)
			}

			relStyle.LineColor = ast.FromToken(tok)
		case tokenizer.IDENTIFIER:
			switch tok.Literal {
			case "thickness":
				eqTok, ok := p.stream.TryConsumeType(tokenizer.EQUALS)
				if !ok {
					return relStyle, NewParserErrorf(eqTok, "expected EQAULS(=) after 'thickness', got: %s(%s)", tok.Type, tok.Literal)
				}
				// parse the thickness value
				numTok, ok := p.stream.TryConsumeType(tokenizer.NUMBER)
				if !ok {
					return relStyle, NewParserErrorf(numTok, "expected a number to describe thickness, got: %s(%s)", numTok.Type, numTok.Literal)
				}

				// err is impossible here, parsing a number guarantees that numTok IS actually a number
				thickness, _ := strconv.Atoi(numTok.Literal)
				relStyle.Thickness = thickness
			case "plain":
				relStyle.LineStyle = ast.LineNormal
			case "bold":
				relStyle.LineStyle = ast.LineBold
			case "dashed":
				relStyle.LineStyle = ast.LineDashed
			case "dotted":
				relStyle.LineStyle = ast.LineDotted
			default:
				return relStyle, NewParserErrorf(tok, "recieved unexpected identifier in arrow attributes: %s", tok.Literal)
			}
		}
		if !p.stream.AssertAnyType(tokenizer.COMMA, tokenizer.RBRACKET) {
			return relStyle, NewParserError(p.stream.PeekTokenAt(0), "expected an attribute terminator")
		}
		if _, ok := p.stream.TryConsumeType(tokenizer.COMMA); ok {
			continue
		}
	}
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
		if _, ok := p.stream.TryConsumeType(tokenizer.RPAREN); !ok {
			return NewParserError(p.stream.PeekTokenAt(0), "Expected ')' for lollipop relationship")
		}
		isLolipop = true
		rel.LArrow = rune(tok.Literal[0])
	// position INDEPENDENT start and end tokens
	case tokenizer.IDENTIFIER:
		// special case for 'x' and 'o' in relationship
		if tok.Literal != "x" && tok.Literal != "o" {
			return NewParserError(tok, "Unexpected identifier in relationship definition")
		}
		fallthrough
	case tokenizer.HASH, tokenizer.ASTERISK, tokenizer.PLUS, tokenizer.CARET:
		rel.LArrow = rune(tok.Literal[0])
	case tokenizer.DOT, tokenizer.DASH: // so that encountering them doesn't cause an error
	default:
		return NewParserError(tok, "Unexpected token at the start of relationship definition")
	}

	if tok.Type != tokenizer.DOT && tok.Type != tokenizer.DASH {
		tok = p.stream.Emit() // consume the asserted start token, if any
	}

	bodyTokType := tok.Type
	var oppositeBodyTokType tokenizer.TokenType
	switch bodyTokType {
	case tokenizer.DOT:
		rel.BodyCount = 1
		oppositeBodyTokType = tokenizer.DASH
		switch rel.LArrow {
		case '<':
			rel.TypeLHS = ast.RelationDependency
		case '|':
			rel.TypeLHS = ast.RelationRealization
		}
	case tokenizer.DASH:
		rel.BodyCount = 1
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
		return NewParserError(tok, "Unexpected token as the relationship body")
	}
	var ok bool
	rel.Body = rune(tok.Literal[0])
	for tok.Type != tokenizer.EOF && tok.Type != tokenizer.NEWLINE {
		if tok, ok = p.stream.TryConsumeType(bodyTokType); ok {
			rel.BodyCount++
			continue
		} else if tok, ok = p.stream.TryConsumeType(oppositeBodyTokType); ok {
			// Simply convenient error message
			return NewParserError(tok, "Different body type runes in relationship")
		}
		if !p.stream.AssertType(tokenizer.LBRACKET) && !p.stream.AssertKW(keyword.Direction) {
			nextTok := p.stream.PeekTokenAt(0)
			if nextTok.Type != tokenizer.IDENTIFIER {
				break
			}
			afterIdent := p.stream.PeekTokenAt(1)
			if afterIdent.Type == bodyTokType || afterIdent.Type == tokenizer.LBRACKET {
				// Valid case, misspelled identifier
				// But we fail fast on speculative errors
				return NewParserError(nextTok, "Unexpected identifier in relationship definition, did you mean to write a direction?")
			}
			// RHS identifier encountered
			break
		} else if sawAttrs || sawDirection {
			return NewParserError(tok, "Cannot separate direction and attributes with a body token")
		}

		if isLolipop {
			return NewParserError(tok, "Lolipop interface cannot contain attributes or direction")
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
						return NewParserError(tok, "Unexpected direction in relationship")
					}
					rel.Direction = dir
				}
			}
			if !sawAttrs {
				if _, ok := p.stream.TryConsumeType(tokenizer.LBRACKET); ok {
					sawAttrs = true
					// We have matched attribute container start
					// after the body, now - parse the attributes

					style, err := p.parseArrowAttributes()
					var pe ParserError
					if errors.As(err, &pe) {
						return err
					}

					rel.Style = style
				}
			}
		}
		// If none matched, we mustn't consume trailing arrow body rune
		if !sawAttrs && !sawDirection {
			continue
		}

		// consume trailing arrow body rune
		if tok, ok = p.stream.TryConsumeType(bodyTokType); !ok {
			return NewParserError(tok, "Unexpected token in body relationship definition")
		} else {
			rel.BodyCount++
		}
	}

	tok = p.stream.PeekTokenAt(0)
	switch tok.Type {
	case tokenizer.PIPE:
		// should only be encountered on --|> case as the end of the relationship
		p.stream.Emit()
		if _, ok := p.stream.TryConsumeType(tokenizer.RANGLE); !ok {
			return NewParserError(tok, "Expected '|>' after relationship")
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
			return NewParserError(tok, "Lolipop interface cannot contain direction or attributes")
		}
		if isLolipop {
			return NewParserError(tok, "Cannot have double headed lolipop relationship")
		}
		p.stream.Emit()
		if _, ok := p.stream.TryConsumeType(tokenizer.RPAREN); !ok {
			return NewParserError(p.stream.PeekTokenAt(0), "Expected ')' for lollipop relationship")
		}
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
			// Single token arrow on relationship
			return nil
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
		return NewParserError(tok, "Missing body in the relationship")
	}
	return nil
}
