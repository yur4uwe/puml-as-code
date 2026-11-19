package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

type Line struct {
	Span     tokenizer.SourceSpan
	Leading  []ast.Comment
	Trailing []ast.Comment

	Tokens []tokenizer.Token
}
