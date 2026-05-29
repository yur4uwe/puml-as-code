package ast

type Statement interface {
	StatementNode() any
}

type Member interface {
	MemberNode() any
}

type DirectiveKind int

const (
	Unknown DirectiveKind = iota
	Include
	Hide
	Show
	Remove
	Restore
	Set
	Scale
)

type Directive struct {
}

type EntityKind int

const (
	UnknownEntityKind EntityKind = iota
	ClassEntityKind
	InterfaceEntityKind
	PackageEntityKind
)

type Entity struct {
	Identifier string
	Alias      string
	Kind       EntityKind
	Properties map[string]string
}

type Relationship struct {
	From string
	To   string
	Type RelationType

	MultiplicityFrom Cardinality
	MultiplicityTo   Cardinality

	Comment string
}

type Diagram struct {
	Name       string
	Title      string
	Statements []Statement
}
