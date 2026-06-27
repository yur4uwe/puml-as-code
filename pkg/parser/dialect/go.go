package dialect

import (
	"errors"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func NewGoDialect() Dialect {
	return GoDialect{}
}

type GoDialect struct{}

var _ Dialect = GoDialect{}

func (g GoDialect) Name() string {
	return "go"
}

func (g GoDialect) ParseField(toks []tokenizer.Token) (string, ast.TypeRef, error) {
	if len(toks) == 0 {
		return "", ast.TypeRef{}, errors.New("expected at least one token")
	}
	return "", ast.TypeRef{}, nil
}

func (g GoDialect) ParseMethod(toks []tokenizer.Token) (string, ast.TypeRef, []ast.Parameter, error) {
	if len(toks) == 0 {
		return "", ast.TypeRef{}, nil, errors.New("expected at least one token")
	}
	return "", ast.TypeRef{}, nil, nil
}
