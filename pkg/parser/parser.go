package parser

import (
	"errors"
	"fmt"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

type parserError struct {
	Err error
	Pos tokenizer.TokenPos
}

var _ error = parserError{}

func (e parserError) Error() string {
	return fmt.Sprintf("parser: %v at %d:%d", e.Err, e.Pos.Line, e.Pos.Col)
}

func (e parserError) Unwrap() error {
	return e.Err
}

func NewParserError(message string, pos tokenizer.TokenPos) error {
	return parserError{Err: errors.New(message), Pos: pos}
}

func WrapParserError(err error, pos tokenizer.TokenPos) error {
	return parserError{Err: err, Pos: pos}
}

type Parser struct {
	ast     *ast.Diagram
	stream  *tokenizer.TokenStream
	dialect dialect.Dialect
}

func (p *Parser) Parse(input string) (*ast.Diagram, error) {
	p.ast = &ast.Diagram{}

	p.stream = tokenizer.NewTokenStream(input)

	startBound, err := p.stream.ReadDiagramBounds()
	if err != nil {
		return nil, err
	} else if !startBound.IsStart {
		return nil, NewParserError("Expected diagram start marker", p.stream.PeekTokenAt(0).Pos)
	}

	p.ast.Statements = append(p.ast.Statements, startBound)
	p.ast.Name = startBound.Name

	for {
		// End condition check should be before consuming a token to avoid swallowing '@'
		if p.stream.IsDiagramBound() {
			endBound, err := p.stream.ReadDiagramBounds()
			if err != nil {
				return nil, err
			}
			if endBound.IsStart {
				return nil, NewParserError("Unexpected diagram end marker", p.stream.PeekTokenAt(0).Pos)
			}
			if endBound.Type != startBound.Type {
				return nil, NewParserError("Types of starting and ending markers don't match", p.stream.PeekTokenAt(0).Pos)
			}
			p.ast.Statements = append(p.ast.Statements, endBound)
			break
		}

		tok := p.stream.Emit()
		if tok.Type == tokenizer.EOF {
			return nil, NewParserError("Unexpected EOF", tok.Pos)
		} else if tok.Type == tokenizer.NEWLINE {
			// We can leave it like this for now
			// If the newline is relevant it will be consumed
			// by the statement parser in some function
			continue
		}

		// Here tokens that are first in line are handled
		// Parsing shuld be constructed to result in a single statement per iteration

		// Imports via !include
		// Styles via <style>
		// Keyword handling switch
		// Handle comments
		// Handle Identifiers

		stmt, err := p.parseDiagramOnlyStatement(tok)
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			p.ast.Statements = append(p.ast.Statements, stmt)
			continue
		}

		stmt, err = p.parseContainerStatement(tok)
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			p.ast.Statements = append(p.ast.Statements, stmt)
			continue
		}

	}

	return p.ast, nil
}

func (p *Parser) parseContainerStatement(tok tokenizer.Token) (ast.Statement, error) {
	if p.HasArrowOnLine() {
		return p.parseRelationship(tok)
	}

	switch keyword.Classify(tok.Literal) {
	case keyword.Class,
		keyword.Interface,
		keyword.Struct,
		keyword.Abstract,
		keyword.Enum,
		keyword.Annotation,
		keyword.Record,
		keyword.Dataclass,
		keyword.Exception,
		keyword.Protocol:
		// Class-like Entities
		return p.parseEntity(tok)
	// Containers
	case keyword.Package,
		keyword.Together,
		keyword.Folder,
		keyword.Frame,
		keyword.Rectangle,
		keyword.Cloud,
		keyword.Database,
		keyword.Namespace,
		keyword.Node:
		return p.parseContainer(tok)
	// Special Keywords
	case keyword.Note:
		return p.parseNote(tok)
	}

	switch tok.Type {
	case tokenizer.IDENTIFIER:
		return p.parseRelationship(tok)
	default:
		return nil, nil
	}
}

func (p *Parser) parseDiagramOnlyStatement(tok tokenizer.Token) (ast.Statement, error) {
	if p.HasArrowOnLine() {
		return p.parseRelationship(tok)
	}

	switch keyword.Classify(tok.Literal) {
	case keyword.Title:
		return nil, p.parseTitle()
	case keyword.Hide, keyword.Show, keyword.Remove, keyword.Restore:
		return p.parseVisibilityCommand(tok)
	case keyword.Scale:
		return p.parseScale()
	// case tokenizer.LANGLE:
	// 	if p.stream.AssertType(tokenizer.RANGLE) {
	// 		// Successfully matched "<>", shorthand for diamond
	// 		return ast.Entity{
	// 			Kind: ast.DiamondKind,
	// 		}, nil
	// 	}
	// 	// <style> token sequence for now simply read it as a string
	// 	// and ignore it
	// 	styles, err := p.stream.ReadBlock(unamb(tokenizer.LANGLE), unamb(tokenizer.SLASH), amb(tokenizer.IDENTIFIER, "style"), unamb(tokenizer.RANGLE))
	// 	if err != nil {
	// 		return nil, NewParserError("Expected style block to end", p.stream.PeekTokenAt(0).Pos)
	// 	}
	// 	p.ast.Styles = append(p.ast.Styles, styles)
	// 	return nil, nil
	case keyword.Skinparam:
		return nil, p.parseSkinparam()
	case keyword.Direction:
		return p.parseDiagDirection(tok)
	case keyword.Set:
		return nil, p.parseSetDirective()
	default:
		return nil, nil
	}
}
