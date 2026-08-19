package parser

import (
	"errors"
	"fmt"

	"yur4uwe/pac/pkg/tokenizer"
)

type parserError struct {
	Err error
	Tok tokenizer.Token
}

var _ error = parserError{}

func (e parserError) Error() string {
	return fmt.Sprintf("parser: %v at %d:%d token: %s(%s)", e.Err, e.Tok.Pos.Line, e.Tok.Pos.Col, e.Tok.Type.String(), e.Tok.Literal)
}

func (e parserError) Unwrap() error {
	return e.Err
}

func NewParserError(message string, tok tokenizer.Token) error {
	return parserError{Err: errors.New(message), Tok: tok}
}

func WrapParserError(err error, tok tokenizer.Token) error {
	return parserError{Err: err, Tok: tok}
}
