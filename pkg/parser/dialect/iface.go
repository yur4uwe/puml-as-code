// Package dialect provides implementations and definitions for Dialect
// interface. Dialects define how to parse fields and methods for a specific
// language dialect.
package dialect

import (
	"errors"
	"slices"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

// ErrParsingDialect indicates the tokens do not match the dialect's expected syntax
var ErrParsingDialect = errors.New("tokens do not match dialect syntax")

// Dialect defines how to parse fields and methods for a specific language dialect
type Dialect interface {
	Name() string
	ParseField(toks []tokenizer.Token) (ast.Parameter, error)
	ParseMethod(toks []tokenizer.Token) (string, []ast.TypeRef, []ast.Parameter, error)
}
