package parser

import (
	"errors"
	"fmt"

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
