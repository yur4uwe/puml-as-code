package dialect

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func NewGoDialect() Dialect {
	return GoDialect{}
}

func (g GoDialect) Name() string {
	return "go"
}

func (g GoDialect) ParseField(toks []tokenizer.Token, visibility ast.VisibilityKind, modifiers []string) (ast.Field, error) {
	return g.parseField(toks, visibility, modifiers)
}

func (g GoDialect) ParseMethod(toks []tokenizer.Token, visibility ast.VisibilityKind, modifiers []string) (ast.Method, error) {
	return g.parseMethod(toks, visibility, modifiers)
}
