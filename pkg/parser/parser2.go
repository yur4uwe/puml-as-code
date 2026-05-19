package parser

import (
	"errors"
	"yur4uwe/pac/pkg/tokenizer"
)

func Parse(input string) error {
	stream := tokenizer.NewTokenStream(input)
	_ = stream

	if str, found := stream.TryReadDiagramBounds(); found && str == "startuml" {
		return errors.New("Could not find diagram start (@startuml)")
	}

	for {
		tok := stream.Emit()
		if tok.Type == tokenizer.EOF {
			return errors.New("Unexpected EOF")
		}

		if str, found := stream.TryReadDiagramBounds(); found && str == "enduml" {
			break
		}
	}

	return nil
}
