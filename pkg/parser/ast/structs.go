package ast

type Statement interface {
	Node() any
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

	MultiplicityFrom Multiplicity
	MultiplicityTo   Multiplicity

	Comment string
}

type Diagram struct {
	Name       string
	Title      string
	Statements []Statement
}
