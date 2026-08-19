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

func (g GoDialect) ParseField(toks []tokenizer.Token, opts *MemberOptions) (ast.Field, error) {
	return g.parseField(toks, opts)
}

func (g GoDialect) ParseMethod(toks []tokenizer.Token, opts *MemberOptions) (ast.Method, error) {
	return g.parseMethod(toks, opts)
}
