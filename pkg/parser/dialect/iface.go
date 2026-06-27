package dialect

import (
	"errors"
	"slices"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

// ErrMismatch indicates the tokens do not match the dialect's expected syntax
var ErrMismatch = errors.New("tokens do not match dialect syntax")

// Dialect defines how to parse fields and methods for a specific language dialect
type Dialect interface {
	Name() string
	ParseField(toks []tokenizer.Token) (string, ast.TypeRef, error)
	ParseMethod(toks []tokenizer.Token) (string, ast.TypeRef, []ast.Parameter, error)
}

type GenericDialect struct{}

var _ Dialect = GenericDialect{}

func (g GenericDialect) Name() string {
	return "generic"
}

func (g GenericDialect) ParseField(toks []tokenizer.Token) (string, ast.TypeRef, error) {
	if len(toks) == 0 {
		return "", ast.TypeRef{}, nil
	}

	// Case 1: Look for a COLON separator (e.g., "flightNumber : Integer")
	colonIdx := -1
	for i, tok := range toks {
		if tok.Type == tokenizer.COLON {
			colonIdx = i
			break
		}
	}

	if colonIdx != -1 {
		name := joinTokens(toks[:colonIdx])
		typeStr := joinTokens(toks[colonIdx+1:])
		return name, ast.ParseTypeRef(typeStr), nil
	}

	// Case 2: No colon, but we have 2 tokens (e.g., "String data" or "data String")
	if len(toks) == 2 {
		return toks[0].Literal, ast.ParseTypeRef(toks[1].Literal), nil
	}

	// Case 3: Single token (e.g., "data") or multiple tokens without clear structure
	return joinTokens(toks), ast.TypeRef{}, nil
}

func (g GenericDialect) ParseMethod(toks []tokenizer.Token) (string, ast.TypeRef, []ast.Parameter, error) {
	if len(toks) == 0 {
		return "", ast.TypeRef{}, nil, nil
	}

	// 1. Locate the first LPAREN
	lparenIdx := slices.IndexFunc(toks, func(tok tokenizer.Token) bool {
		return tok.Type == tokenizer.LPAREN
	})

	// If no LPAREN exists, treat it as a method with no params or return type
	if lparenIdx == -1 {
		return joinTokens(toks), ast.TypeRef{}, nil, nil
	}

	name := joinTokens(toks[:lparenIdx])

	// 2. Locate the matching RPAREN
	rparenIdx := slices.IndexFunc(toks, func(tok tokenizer.Token) bool {
		return tok.Type == tokenizer.RPAREN
	})

	// If RPAREN is missing
	if rparenIdx == -1 {
		return name, ast.TypeRef{}, nil, nil
	}

	// 3. Extract parameter tokens inside the parens
	paramTokens := toks[lparenIdx+1 : rparenIdx]
	params := parseGenericParameters(paramTokens)

	// 4. Extract return type after RPAREN (e.g., ") : ReturnType" or ") ReturnType")
	var returnType ast.TypeRef
	if rparenIdx+1 < len(toks) {
		remaining := toks[rparenIdx+1:]
		if remaining[0].Type == tokenizer.COLON {
			remaining = remaining[1:]
		}
		if len(remaining) > 0 {
			returnType = ast.ParseTypeRef(joinTokens(remaining))
		}
	}

	return name, returnType, params, nil
}

// Helper to split generic parameters by commas and parse each
func parseGenericParameters(toks []tokenizer.Token) []ast.Parameter {
	var params []ast.Parameter
	if len(toks) == 0 {
		return params
	}

	var current []tokenizer.Token
	for _, tok := range toks {
		if tok.Type == tokenizer.COMMA {
			if len(current) > 0 {
				params = append(params, parseSingleGenericParameter(current))
				current = nil
			}
		} else {
			current = append(current, tok)
		}
	}
	if len(current) > 0 {
		params = append(params, parseSingleGenericParameter(current))
	}

	return params
}

func parseSingleGenericParameter(toks []tokenizer.Token) ast.Parameter {
	colonIdx := -1
	for i, t := range toks {
		if t.Type == tokenizer.COLON {
			colonIdx = i
			break
		}
	}

	if colonIdx != -1 {
		return ast.Parameter{
			Name: joinTokens(toks[:colonIdx]),
			Type: ast.ParseTypeRef(joinTokens(toks[colonIdx+1:])),
		}
	}

	if len(toks) == 2 {
		return ast.Parameter{
			Name: toks[0].Literal,
			Type: ast.ParseTypeRef(toks[1].Literal),
		}
	}

	return ast.Parameter{
		Name: joinTokens(toks),
	}
}

// Helper function to reconstruct strings cleanly
func joinTokens(toks []tokenizer.Token) string {
	var parts []string
	for _, t := range toks {
		parts = append(parts, t.Literal)
	}
	return strings.Join(parts, " ")
}
