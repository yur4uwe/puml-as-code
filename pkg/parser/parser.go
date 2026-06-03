package parser

import (
	"fmt"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

type parserError struct {
	Message string
	Pos     tokenizer.TokenPos
}

var _ error = parserError{}

func (e parserError) Error() string {
	return fmt.Sprintf("parser: %s at %d:%d", e.Message, e.Pos.Line, e.Pos.Col)
}

func NewParserError(message string, pos tokenizer.TokenPos) error {
	return parserError{Message: message, Pos: pos}
}

type Parser struct {
	symbol_table map[string]*ast.Entity
	ast          *ast.Diagram
	stream       *tokenizer.TokenStream
	skinparam    ast.Skinparam
	styles       []string // TODO: should be an actual struct and not collection of strings
}

func (p *Parser) Parse(input string) (*ast.Diagram, error) {
	p.ast = &ast.Diagram{}
	p.symbol_table = make(map[string]*ast.Entity)

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
		switch tok.Type {
		case tokenizer.TITLE:
			err = p.parseTitle()
		case tokenizer.HIDE, tokenizer.SHOW, tokenizer.REMOVE, tokenizer.RESTORE:
			err = p.parseVisibilityCommand(tok)
		case tokenizer.SCALE:
			err = p.parseScale()
		case tokenizer.LANGLE, tokenizer.SKINPARAM:
			err = p.parseStyles(tok)
		case tokenizer.EXCLAMATION:
		case tokenizer.DIRECTION:
			err = p.parseDiagDirection(tok)
		}
		if err != nil {
			return nil, err
		}
	}

	return p.ast, nil
}
