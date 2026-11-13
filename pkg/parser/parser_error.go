package parser

import (
	"errors"
	"fmt"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

type ParserError struct {
	Err  error
	Span tokenizer.SourceSpan
}

func (e ParserError) Error() string {
	return fmt.Sprintf("parser: %v at %d:%d", e.Err, e.Span.Start.Line, e.Span.Start.Col)
}

func (e ParserError) Unwrap() error {
	return e.Err
}

type parserError struct {
	Err  error
	Tok  tokenizer.Token
	File string
}

var _ error = parserError{}

func (e parserError) Error() string {
	return fmt.Sprintf("parser: %v at %d:%d token: %s(%s)", e.Err, e.Tok.Span.Start.Line, e.Tok.Span.Start.Col, e.Tok.Type.String(), e.Tok.Literal)
}

func (e parserError) Unwrap() error {
	return e.Err
}

func NewParserError(tok tokenizer.Token, message string) error {
	return ParserError{Err: errors.New(message), Span: tok.Span}
}

func NewParserErrorAtIdent(id ast.Ident, message string) error {
	return ParserError{Err: errors.New(message), Span: id.Span}
}

func NewParserErrorAt(span tokenizer.SourceSpan, message string) error {
	return ParserError{Err: errors.New(message), Span: span}
}

func NewParserErrorf(tok tokenizer.Token, format string, args ...any) error {
	return ParserError{Err: fmt.Errorf(format, args...), Span: tok.Span}
}

func WrapParserError(tok tokenizer.Token, err error) error {
	return ParserError{Err: err, Span: tok.Span}
}
