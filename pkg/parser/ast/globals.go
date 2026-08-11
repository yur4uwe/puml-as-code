package ast

type TitleDef struct {
	Text string
}

// StatementNode implements [Statement].
func (t TitleDef) StatementNode() Statement {
	return t
}

var _ Statement = TitleDef{}
