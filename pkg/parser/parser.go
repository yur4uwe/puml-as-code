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
}

func (p *Parser) Parse(input string) (*ast.Diagram, error) {
	p.ast = &ast.Diagram{}
	p.symbol_table = make(map[string]*ast.Entity)

	p.stream = tokenizer.NewTokenStream(input)

	if str, found := p.stream.TryReadDiagramBounds(); !found || str != "startuml" {
		return nil, NewParserError("Could not find diagram start (@startuml)", p.stream.PeekTokenAt(0).Pos)
	}

	if title := p.stream.ReadUntilNewline(); title != "" {
		p.ast.Name = title
	}

	for {
		// End condition check should be before consuming a token to avoid swallowing '@'
		if str, found := p.stream.TryReadDiagramBounds(); found && str == "enduml" {
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

		switch tok.Type {
		case tokenizer.TITLE:
			p.parseTitle()
		case tokenizer.HIDE, tokenizer.SHOW, tokenizer.REMOVE, tokenizer.RESTORE:

		case tokenizer.LANGLE:
		case tokenizer.EXCLAMATION:
		case tokenizer.SLASH:

		}

	}

	return p.ast, nil
}
