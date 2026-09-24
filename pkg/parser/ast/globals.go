// Package ast contains the AST nodes for the parser.
package ast

import (
	"yur4uwe/pac/pkg/tokenizer"
)

//go:generate enumer -type=TextBlockKind -transform=lower -json -trimprefix=Block
type TextBlockKind int

const (
	BlockUnknown TextBlockKind = iota
	BlockLegend
	BlockHeader
	BlockFooter
	BlockTitle
)

// TextBlock implements [Statement].
// if you want to know whether it is multiline, check for \n
type TextBlock struct {
	Kind TextBlockKind
	Text string

	VerticalAlignment   string
	HorizontalAlignment string

	Trivia
}

// StatementNode implements [Statement].
func (t TextBlock) StatementNode() Statement {
	return t
}

var _ Statement = TextBlock{}

type UnhandledStatement struct {
	Text string
	Span tokenizer.SourceSpan
	Trivia
}

// StatementNode implements [Statement].
func (t UnhandledStatement) StatementNode() Statement {
	return t
}

var _ Statement = UnhandledStatement{}
