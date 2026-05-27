package parser

import (
	"errors"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func Parse(input string) (*ast.ClassDiagram, error) {
	stream := tokenizer.NewTokenStream(input)
	_ = stream

	if str, found := stream.TryReadDiagramBounds(); found && str == "startuml" {
		return nil, errors.New("Could not find diagram start (@startuml)")
	}

	for {
		tok := stream.Emit()
		if tok.Type == tokenizer.EOF {
			return nil, errors.New("Unexpected EOF")
		}

		if str, found := stream.TryReadDiagramBounds(); found && str == "enduml" {
			break
		}
	}

	return nil, nil
}
