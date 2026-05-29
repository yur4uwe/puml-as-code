package parser

import (
	"errors"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

type Parser struct {
	symbol_table map[string]*ast.Entity
	ast          *ast.Diagram
}

func (p *Parser) Parse(input string) (*ast.Diagram, error) {
	p.ast = &ast.Diagram{}
	p.symbol_table = make(map[string]*ast.Entity)

	stream := tokenizer.NewTokenStream(input)

	if str, found := stream.TryReadDiagramBounds(); !found || str != "startuml" {
		return nil, errors.New("Could not find diagram start (@startuml)")
	}

	if title, found := stream.ReadUntilNewline(); found {
		p.ast.Name = title
	}

	for {
		// End condition check should be before consuming a token to avoid swallowing '@'
		if str, found := stream.TryReadDiagramBounds(); found && str == "enduml" {
			break
		}

		tok := stream.Emit()
		if tok.Type == tokenizer.EOF {
			return nil, errors.New("Unexpected EOF")
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
			if title, found := stream.ReadUntilNewline(); found {
				p.ast.Title = title
			}
		case tokenizer.HIDE, tokenizer.SHOW, tokenizer.REMOVE, tokenizer.RESTORE:
		case tokenizer.LANGLE:
		case tokenizer.EXCLAMATION:
		case tokenizer.SLASH:

		}

	}

	return p.ast, nil
}
