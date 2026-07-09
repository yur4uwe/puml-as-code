package parser

import (
	"errors"
	"fmt"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
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

		var err error
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
	switch tok.Type {
	// Class-like Entities
	case tokenizer.CLASS,
		tokenizer.INTERFACE,
		tokenizer.STRUCT,
		tokenizer.ABSTRACT,
		tokenizer.ENUM,
		tokenizer.ANNOTATION,
		tokenizer.RECORD,
		tokenizer.DATACLASS,
		tokenizer.EXCEPTION,
		tokenizer.PROTOCOL:
		return p.parseEntity(tok)
	// Containers
	case tokenizer.PACKAGE, tokenizer.TOGETHER:
		return p.parseContainer(tok)
	// Special Keywords
	case tokenizer.NOTE:
		return p.parseNote(tok)
	// Handle short-form circle: ()
	case tokenizer.LPAREN:
	case tokenizer.IDENTIFIER:
		// TODO: handle relationships
		return p.parseRelationship(tok)
	default:
		return nil, nil
	}
	panic("unreachable")
}

func (p *Parser) parseDiagramOnlyStatement(tok tokenizer.Token) (ast.Statement, error) {
	switch tok.Type {
	case tokenizer.TITLE:
		return nil, p.parseTitle()
	case tokenizer.HIDE, tokenizer.SHOW, tokenizer.REMOVE, tokenizer.RESTORE:
		return p.parseVisibilityCommand(tok)
	case tokenizer.SCALE:
		return p.parseScale()
	case tokenizer.LANGLE, tokenizer.SKINPARAM:
		if p.stream.AssertType(tokenizer.RANGLE) {
			// Successfully matched "<>", shorthand for diamond
			return ast.Entity{
				Kind: ast.DiamondKind,
			}, nil
		}
		return nil, p.parseStyles(tok)
	case tokenizer.EXCLAMATION:
	case tokenizer.DIRECTION:
		return p.parseDiagDirection(tok)
	default:
		return nil, nil
	}
	panic("unreachable")
}
