package parser

import (
	"errors"
	"log"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

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
			p.recordDiagnostic(p.newDiagnosticError(err, span, container))
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
			p.recordDiagnostic(p.newDiagnosticError(unclosedBlockErr, lbraceTok.Span, container))
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
	if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
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

	cntr.Color = p.tryParseColor()

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
			return ast.Ident{}, ast.TargetRef{}, NewParserErrorFromIdent(lhsTarget.Name, "container name contains unexpected member")
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
			return ast.Ident{}, ast.TargetRef{}, NewParserErrorFromIdent(rhsTarget.Name, "container alias contains unexpected member")
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
