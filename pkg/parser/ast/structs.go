package ast

type Statement interface {
	StatementNode() any
}

type Member interface {
	MemberNode() any
}

type EntityKind int

const (
	UnknownKind EntityKind = iota
	ClassKind
	InterfaceKind
	PackageKind
)

type Entity struct {
	Identifier string
	Alias      string
	Kind       EntityKind
	Members    []Member
}

var _ Statement = Entity{}
var _ Member = Entity{}

func (e Entity) StatementNode() any {
	return e
}

func (e Entity) MemberNode() any {
	return e
}

type Relationship struct {
	From string
	To   string
	Type RelationType

	MultiplicityFrom Cardinality
	MultiplicityTo   Cardinality

	Comment string
}

var _ Statement = Relationship{}
var _ Member = Relationship{}

func (r Relationship) StatementNode() any {
	return r
}

func (r Relationship) MemberNode() any {
	return r
}

type Diagram struct {
	Name       string
	Title      string
	Statements []Statement
}
