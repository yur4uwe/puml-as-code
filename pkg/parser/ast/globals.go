// Package ast contains the AST nodes for the parser.
package ast

type TitleDef struct {
	Text string
	Trivia
}

// StatementNode implements [Statement].
func (t TitleDef) StatementNode() Statement {
	return t
}

var _ Statement = TitleDef{}

type UnhandledStatement struct {
	Text string
	Trivia
}

// StatementNode implements [Statement].
func (t UnhandledStatement) StatementNode() Statement {
	return t
}

var _ Statement = UnhandledStatement{}
