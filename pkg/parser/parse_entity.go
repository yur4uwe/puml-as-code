package parser

import (
	"errors"
	"fmt"
	"log"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"
)

func (p *Parser) parseEntityStyle(e *ast.Entity) *ast.EntityStyle {
	firstHash, ok := p.stream.TryConsumeType(tokenizer.HASH)
	if !ok {
		return nil
	}
	mark := p.Mark(firstHash)

	entStyle := &ast.EntityStyle{
		Span: firstHash.Span,
	}

	decisionTok := p.stream.Emit()
	if decisionTok.Type == tokenizer.IDENTIFIER || decisionTok.Type == tokenizer.NUMBER {
		isKeyVal := false
		switch decisionTok.Literal {
		case "back", "header", "line", "text":
			isKeyVal = true
		default: // actual color
			bg, err := p.parseColorExpr(decisionTok)
			if err != nil {
				entStyle.Span = p.Span(mark)
				p.ast.RecordError(err, p.stream.LastSemanticToken().Span, e)
				p.syncToBlockStart()
				return entStyle
			}
			entStyle.Background = bg
		}

		hasSemi := false
		if !isKeyVal {
			_, hasSemi = p.stream.TryConsumeType(tokenizer.SEMICOLON)
		}

		if isKeyVal || hasSemi {
			tok := decisionTok
			for {
				targetKind := ""
				switch tok.Literal {
				case "line":
					targetKind = tok.Literal
					if !p.stream.AssertType(tokenizer.DOT) {
						break
					}
					p.stream.Emit() // consume '.'

					lineStyleTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
					if !ok {
						entStyle.Span = p.Span(mark)
						err := fmt.Errorf("unexpected token describing line style: %s(%s)", lineStyleTok.Type, lineStyleTok.Literal)
						p.ast.RecordError(err, lineStyleTok.Span, e)
						p.syncToBlockStart()
						return entStyle
					}

					ls, err := parseLineStyle(lineStyleTok)
					if err != nil {
						entStyle.Span = p.Span(mark)
						p.ast.RecordError(err, lineStyleTok.Span, e)
						p.syncToBlockStart()
						return entStyle
					}
					entStyle.LineStyle = ls

					if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); ok {
						var ok bool
						tok, ok = p.stream.TryConsumeType(tokenizer.IDENTIFIER)
						if !ok {
							entStyle.Span = p.Span(mark)
							err := fmt.Errorf("unexpected token describing entity colorable target: %s(%s)", tok.Type, tok.Literal)
							p.ast.RecordError(err, tok.Span, e)
							p.syncToBlockStart()
							return entStyle
						}
						continue
					}
				case "back", "header", "text":
					targetKind = tok.Literal
				default:
					entStyle.Span = p.Span(mark)
					err := fmt.Errorf("unexpected value for color target specifier: %s", tok.Literal)
					p.ast.RecordError(err, tok.Span, e)
					p.syncToBlockStart()
					return entStyle
				}

				if _, ok := p.stream.TryConsumeType(tokenizer.COLON); !ok {
					entStyle.Span = p.Span(mark)
					tok := p.stream.PeekTokenAt(0)
					err := errors.New("missing colon after specified color target")
					p.ast.RecordError(err, tok.Span, e)
					p.syncToBlockStart()
					return entStyle
				}

				colorTok := p.stream.Emit()
				expr, err := p.parseColorExpr(colorTok)
				if err != nil {
					entStyle.Span = p.Span(mark)
					p.ast.RecordError(err, p.stream.LastSemanticToken().Span, e)
					p.syncToBlockStart()
					return entStyle
				}

				switch targetKind {
				case "header":
					entStyle.Header = expr
				case "text":
					entStyle.TextColor = expr
				case "line":
					entStyle.LineColor = expr
				case "back":
					entStyle.Background = expr
				}

				if !p.stream.AssertType(tokenizer.SEMICOLON) {
					break
				}

				p.stream.Emit() // emit the semicolon

				// update the active target tok
				var ok bool
				tok, ok = p.stream.TryConsumeType(tokenizer.IDENTIFIER)
				if !ok {
					entStyle.Span = p.Span(mark)
					err := fmt.Errorf("unexpected token describing entity colorable target: %s(%s)", tok.Type, tok.Literal)
					p.ast.RecordError(err, tok.Span, e)
					p.syncToBlockStart()
					return entStyle
				}
			}
		}

		// Trailing ##[lineStyle]lineColor after #color or composite styles
		if p.stream.AssertTypeSeq(tokenizer.HASH, tokenizer.HASH) {
			p.stream.Emit() // first #
			decisionTok = p.stream.Emit() // second #
		}
	}

	if decisionTok.Type == tokenizer.HASH {
		// ##[lineStyle]lineColor or ##lineColor
		if _, ok := p.stream.TryConsumeType(tokenizer.LBRACKET); ok {
			lineStyleTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
			if !ok {
				entStyle.Span = p.Span(mark)
				err := fmt.Errorf("unexpected token describing line style: %s(%s)", lineStyleTok.Type, lineStyleTok.Literal)
				p.ast.RecordError(err, lineStyleTok.Span, e)
				p.syncToBlockStart()
				return entStyle
			}

			ls, err := parseLineStyle(lineStyleTok)
			if err != nil {
				entStyle.Span = p.Span(mark)
				p.ast.RecordError(err, lineStyleTok.Span, e)
				p.syncToBlockStart()
				return entStyle
			}
			entStyle.LineStyle = ls

			if _, ok := p.stream.TryConsumeType(tokenizer.RBRACKET); !ok {
				entStyle.Span = p.Span(mark)
				tok := p.stream.PeekTokenAt(0)
				err := fmt.Errorf("unexpected token in line style syntax: %s(%s)", tok.Type, tok.Literal)
				p.ast.RecordError(err, tok.Span, e)
				p.syncToBlockStart()
				return entStyle
			}
		}

		lineColorTok := p.stream.Emit()
		if lineColorTok.Type != tokenizer.IDENTIFIER && lineColorTok.Type != tokenizer.NUMBER {
			entStyle.Span = p.Span(mark)
			err := fmt.Errorf("unexpected token describing line color: %s(%s)", lineColorTok.Type, lineColorTok.Literal)
			p.ast.RecordError(err, lineColorTok.Span, e)
			p.syncToBlockStart()
			return entStyle
		}

		// don't validate the colors right now
		entStyle.LineColor = ast.FromToken(lineColorTok)
		entStyle.Span = p.Span(mark)
		return entStyle // there cannot be any color notation after this
	}

	switch decisionTok.Type {
	case tokenizer.IDENTIFIER, tokenizer.NUMBER, tokenizer.HASH:
	default:
		entStyle.Span = p.Span(mark)
		err := fmt.Errorf("unexpected token in style declaration: %s(%s)", decisionTok.Type, decisionTok.Literal)
		p.ast.RecordError(err, decisionTok.Span, e)
		p.syncToBlockStart()
		return entStyle
	}

	entStyle.Span = p.Span(mark)
	return entStyle
}

func (p *Parser) setAliasAndName(ent *ast.Entity, nameOrAlias tokenizer.Token) error {
	switch nameOrAlias.Type {
	case tokenizer.STRING:
		if !ent.Alias.IsZero() {
			return NewParserError(nameOrAlias, "entity alias already set")
		}
		ent.Alias = ast.IdentFromToken(nameOrAlias)
		return nil
	case tokenizer.IDENTIFIER:
		if !ent.Identifier.Name.IsZero() {
			return NewParserError(nameOrAlias, "entity name already set")
		}
		targetRef, err := p.parseTargetRef(nameOrAlias)
		if err != nil || targetRef.Member != nil {
			return NewParserErrorFromIdent(targetRef.Name, "entity name contains unexpected member")
		}
		ent.Identifier = targetRef
		return nil
	default:
		return NewParserError(nameOrAlias, "expected token for entity identifier or alias")
	}
}

func (p *Parser) parseEntityHeader(ent *ast.Entity) {
	// Unexpectedly, class and other entity definitions have very strict syntax:
	// class <entity name> as <entity alias> <generics> <stereotype> <styles> <body>

	identTok := p.stream.Emit()
	err := p.setAliasAndName(ent, identTok)
	if err != nil {
		p.ast.RecordError(err, identTok.Span, ent)
		p.syncToBlockStart()
		return
	}

	if _, hasAlias := p.stream.TryConsumeKW(keyword.Alias); hasAlias {
		secondIdentTok := p.stream.Emit()
		if err := p.setAliasAndName(ent, secondIdentTok); err != nil {
			p.ast.RecordError(err, secondIdentTok.Span, ent)
			p.syncToBlockStart()
			return
		}
	}

	if ent.Identifier.Name.IsZero() {
		// If no identifier is set, use the alias
		ent.Identifier = ast.TargetRef{Name: ent.Alias}
		ent.Alias = ast.Ident{}
	}

	if !p.stream.AssertTypeSeq(tokenizer.LANGLE, tokenizer.LANGLE) {
		gen, span, err := p.tryReadGeneric()
		if err != nil && !errors.Is(err, stream.ErrStartMarkerNotFound) {
			p.ast.RecordError(err, span, ent)
			p.syncToBlockStart()
			return
		}
		ent.Generic = gen
	}

	preTags, span, err := p.tryReadTags()
	if err != nil {
		p.ast.RecordError(err, span, ent)
		p.syncToBlockStart()
		return
	}

	stereo, span, err := p.tryParseStereotype()
	if err != nil && !errors.Is(err, stream.ErrStartMarkerNotFound) {
		p.ast.RecordError(err, span, ent)
		p.syncToBlockStart()
		return
	}
	ent.Stereotype = stereo

	postTags, span, err := p.tryReadTags()
	if err != nil {
		p.ast.RecordError(err, span, ent)
		p.syncToBlockStart()
		return
	}

	if len(preTags) > 0 {
		if len(postTags) > 0 {
			log.Printf("warning: tags %v after stereotype on entity %q are ignored in favor of pre-stereotype tags %v\n", postTags, ent.Identifier.EntityName(), preTags)
		}
		ent.Tags = preTags
	} else if len(postTags) > 0 {
		ent.Tags = postTags
	}

	ent.Style = p.parseEntityStyle(ent)
}

// tok is the kind of an entity (class, interface, struct, enum, etc.)
func (p *Parser) parseEntity(tok tokenizer.Token) *ast.Entity {
	ent := &ast.Entity{
		Kind: p.mapTokenToEntityKind(tok),
	}

	p.parseEntityHeader(ent)

	lbraceTok, ok := p.stream.TryConsumeType(tokenizer.LBRACE)
	if !ok {
		// No body, return entity as is
		return ent
	}

	ent.LBrace = lbraceTok.Span.Start

	p.stream.EmitCommentToks()
	ent.TrailingTrivia = p.dumpTrivia()

	unclosedBlockErr := ast.NewUnclosedBlockError("class", "}")

	// We are inside a body
	for {
		for {
			if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
				break
			}
		}

		if p.stream.AssertType(tokenizer.EOF) || p.isEndMarker() {
			// place unclosed block diagnostic on the opener of the block
			p.ast.RecordError(unclosedBlockErr, lbraceTok.Span, ent)
			break
		} else if p.stream.AssertType(tokenizer.RBRACE) {
			break
		}

		member := p.routeEntityMember(p.stream.PeekTokenAt(0))
		if member == nil {
			break
		}
		ent.Members = append(ent.Members, member)
	}

	rbraceTok, ok := p.stream.TryConsumeType(tokenizer.RBRACE)
	if !ok {
		return ent
	}
	ent.RBrace = rbraceTok.Span.Start

	return ent
}
