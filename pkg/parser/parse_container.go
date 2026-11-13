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

func (p *Parser) parseContainerStyle() (*ast.ContainerStyle, tokenizer.SourceSpan, error) {
	cntStyle := &ast.ContainerStyle{}

	tokenAfterHash := p.stream.PeekTokenAt(0)
	if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.NUMBER) {
		return cntStyle, tokenAfterHash.Span, fmt.Errorf("unexpected token in style declaration: %s(%s)", tokenAfterHash.Type, tokenAfterHash.Literal)
	}

	// --- First item: may be a bare color (only valid here) ---
	switch p.stream.PeekTokenAt(0).Literal {
	case "back", "line", "text":
	default:
		tok := p.stream.Emit()
		bg, err := p.parseColorExpr(tok)
		if err != nil {
			return cntStyle, p.stream.LastSemanticToken().Span, err
		}
		cntStyle.Background = bg

		if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); !ok {
			return cntStyle, tokenizer.SourceSpan{}, nil
		}
	}

	// --- Remaining items: strictly target:color pairs ---
	for {
		tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
		if !ok {
			return cntStyle, tok.Span, fmt.Errorf("unexpected token describing entity colorable target: %s(%s)", tok.Type, tok.Literal)
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
				return cntStyle, lineStyleTok.Span, fmt.Errorf("unexpected token describing line style: %s(%s)", lineStyleTok.Type, lineStyleTok.Literal)
			}

			ls, err := parseLineStyle(lineStyleTok)
			if err != nil {
				return cntStyle, lineStyleTok.Span, err
			}
			cntStyle.LineStyle = ls

			if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); ok {
				continue
			}
			// no semicolon — fall through to parse optional :color
		case "back", "text":
			targetKind = tok.Literal
		default:
			return cntStyle, tok.Span, fmt.Errorf("unexpected value for color target specifier: %s", tok.Literal)
		}

		if _, ok := p.stream.TryConsumeType(tokenizer.COLON); !ok {
			peekTok := p.stream.PeekTokenAt(0)
			return cntStyle, peekTok.Span, errors.New("missing colon after specified color target")
		}

		colorTok := p.stream.Emit()
		expr, err := p.parseColorExpr(colorTok)
		if err != nil {
			return cntStyle, p.stream.LastSemanticToken().Span, err
		}

		switch targetKind {
		case "text":
			cntStyle.TextColor = expr
		case "line":
			cntStyle.LineColor = expr
		case "back":
			cntStyle.Background = expr
		}

		if _, ok := p.stream.TryConsumeType(tokenizer.SEMICOLON); !ok {
			break
		}
	}

	return cntStyle, tokenizer.SourceSpan{}, nil
}

func (p *Parser) parseContainer(tok tokenizer.Token) *ast.Container {
	containerClass := keyword.Classify(tok.Literal)
	container := &ast.Container{
		Kind: p.mapKeywordToContainerKind(containerClass),
	}

	// Handle container name, alias, stereotype, color
	// Name rules:
	// - Name and alias must be a string or Identifier
	// - Name and alias CAN be of the same type
	// - if Name and alias are of the same type, alias is on the left
	// and Name is on the right of 'as' keyword

	if containerClass != keyword.Together {
		if err, span := p.parseContainerHeader(container); err != nil {
			p.ast.RecordError(err, span, container)
			p.syncToBlockStart()
		}
	}

	lbraceTok, ok := p.stream.TryConsumeType(tokenizer.LBRACE)
	if !ok {
		return container
	}
	container.LBrace = lbraceTok.Span.Start

	p.stream.EmitCommentToks()
	container.TrailingTrivia = p.dumpTrivia()

	unclosedBlockErr := ast.NewUnclosedBlockError(tok.Literal, "}")

	for {
		for {
			if _, ok := p.stream.TryConsumeType(tokenizer.NEWLINE); !ok {
				break
			}
		}

		if p.stream.AssertType(tokenizer.EOF) || p.isEndMarker() {
			p.ast.RecordError(unclosedBlockErr, lbraceTok.Span, container)
			break
		} else if p.stream.AssertType(tokenizer.RBRACE) {
			break
		}

		innerTok := p.stream.Emit()
		stmt := p.routeContainerStatement(innerTok)
		if stmt != nil {
			container.Statements = append(container.Statements, stmt)
		}
	}

	if rbraceTok, ok := p.stream.TryConsumeType(tokenizer.RBRACE); ok {
		container.RBrace = rbraceTok.Span.Start
	}

	return container
}

func (p *Parser) parseContainerHeader(cntr *ast.Container) (error, tokenizer.SourceSpan) {
	var err error
	var span tokenizer.SourceSpan
	cntr.Alias, cntr.Identifier, err = p.parseContainerIdentAndAlias()
	if err != nil {
		if perr, ok := err.(parserError); ok {
			span = perr.Tok.Span
		}
		return err, span
	}
	var preTags []string
	preTags, span, err = p.tryReadTags()
	if err != nil {
		return err, span
	}
	cntr.Stereotype, span, err = p.tryParseStereotype()
	if err != nil && !errors.Is(err, stream.ErrStartMarkerNotFound) {
		return err, span
	}
	var postTags []string
	postTags, span, err = p.tryReadTags()
	if err != nil {
		return err, span
	}
	if len(preTags) > 0 {
		if len(postTags) > 0 {
			log.Printf("warning: tags %v after stereotype on container %q are ignored in favor of pre-stereotype tags %v\n", postTags, cntr.Identifier.EntityName(), preTags)
		}
		cntr.Tags = preTags
	} else if len(postTags) > 0 {
		cntr.Tags = postTags
	}

	if hashTok, ok := p.stream.TryConsumeType(tokenizer.HASH); ok {
		mark := p.Mark(hashTok)
		style, errSpan, err := p.parseContainerStyle()
		style.Span = p.Span(mark)
		cntr.Style = style
		if err != nil {
			return err, errSpan
		}
	}

	if !p.stream.AssertAnyType(tokenizer.LBRACE, tokenizer.NEWLINE, tokenizer.EOF) && !p.isEndMarker() {
		nextTok := p.stream.PeekTokenAt(0)
		return NewParserError(nextTok, "unexpected token after container declaration"), nextTok.Span
	}

	return nil, span
}

// parseContainerIdentAndAlias parses the container name and alias.
//
// Returns:
// - Alias (Ident)
// - Identifier (TargetRef)
// - error
func (p *Parser) parseContainerIdentAndAlias() (ast.Ident, ast.TargetRef, error) {
	firstTok := p.stream.Emit()
	var (
		lhsIdent    ast.Ident
		lhsTarget   ast.TargetRef
		lhsIsTarget bool
		err         error
	)

	switch firstTok.Type {
	case tokenizer.STRING:
		lhsIdent = ast.IdentFromToken(firstTok)
	case tokenizer.IDENTIFIER:
		lhsTarget, err = p.parseTargetRef(firstTok)
		if err != nil {
			return ast.Ident{}, ast.TargetRef{}, err
		}
		if lhsTarget.Member != nil {
			return ast.Ident{}, ast.TargetRef{}, NewParserErrorAtIdent(lhsTarget.Name, "container name contains unexpected member")
		}
		lhsIsTarget = true
	default:
		return ast.Ident{}, ast.TargetRef{}, NewParserError(firstTok, "Expected identifier or string for container name")
	}

	if _, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		if lhsIsTarget {
			return ast.Ident{}, lhsTarget, nil
		}
		return ast.Ident{}, ast.TargetRef{Name: lhsIdent}, nil
	}

	secondTok := p.stream.Emit()
	var (
		rhsIdent    ast.Ident
		rhsTarget   ast.TargetRef
		rhsIsTarget bool
	)

	switch secondTok.Type {
	case tokenizer.STRING:
		rhsIdent = ast.IdentFromToken(secondTok)
	case tokenizer.IDENTIFIER:
		rhsTarget, err = p.parseTargetRef(secondTok)
		if err != nil {
			return ast.Ident{}, ast.TargetRef{}, err
		}
		if rhsTarget.Member != nil {
			return ast.Ident{}, ast.TargetRef{}, NewParserErrorAtIdent(rhsTarget.Name, "container alias contains unexpected member")
		}
		rhsIsTarget = true
	default:
		return ast.Ident{}, ast.TargetRef{}, NewParserError(secondTok, "Expected identifier or string for container alias")
	}

	if lhsIsTarget && !rhsIsTarget {
		return rhsIdent, lhsTarget, nil
	}
	if !lhsIsTarget && rhsIsTarget {
		return lhsIdent, rhsTarget, nil
	}
	if lhsIsTarget && rhsIsTarget {
		// When both are identifiers, alias is on the left, identifier is on the right
		return lhsTarget.Name, rhsTarget, nil
	}
	// When both are strings
	return lhsIdent, ast.TargetRef{Name: rhsIdent}, nil
}
