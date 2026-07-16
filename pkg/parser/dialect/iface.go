// Package dialect provides implementations and definitions for Dialect
// interface.
//
// Dialect describes a way to parse a method or field declaration by adhering
// to a specific language syntax.
//
// E.g. Go dialect expects a method declaration to be of the form: Method(a int, b bool) error
//
// And a field declaration to be of the form: Field string
//
// Dialects are responsible for parsing the tokens and producing a usable
// representation of the method or field declaration.
package dialect

import (
	"errors"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

// ErrParsingDialect indicates the tokens do not match the dialect's expected syntax
var ErrParsingDialect = errors.New("tokens do not match dialect syntax")

// Dialect defines how to parse fields and methods for a specific language dialect
type Dialect interface {
	Name() string
	ParseField(toks []tokenizer.Token, visibility ast.VisibilityKind, modifiers []string) (ast.Field, error)
	ParseMethod(toks []tokenizer.Token, visibility ast.VisibilityKind, modifiers []string) (ast.Method, error)
}
