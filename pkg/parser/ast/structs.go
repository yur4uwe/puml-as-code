package ast

import "yur4uwe/pac/pkg/tokenizer"

type Statement interface {
	StatementNode() Statement
}

type Member interface {
	MemberNode() Member
}

type EntityKind int

const (
	UnknownEntityKind EntityKind = iota
	ClassKind
	AbstractClassKind
	InterfaceKind
	EnumKind
	EntityClassKind
	StructKind
	AnnotationKind
	ProtocolKind
	CircleKind
	DiamondKind
	ExceptionKind
	MetaclassKind
	RecordKind
	DataclassKind
	StereotypeKind // For the standalone "stereotype" keyword

	// Robustness (BCE)
	BceEntityKind
	BoundaryKind
	ControlKind

	// Component / Mixed
	ActorKind
	ComponentKind
	ArtifactKind
)

func (k EntityKind) AllowsBody() bool {
	switch k {
	case CircleKind, DiamondKind:
		return false
	default:
		return true
	}
}

type Entity struct {
	Identifier     string
	Alias          string
	Kind           EntityKind
	Stereotype     string
	Generic        string
	Color          string
	Members        []Member
	LeadingTrivia  []tokenizer.Token
	TrailingTrivia []tokenizer.Token
}

var _ Statement = Entity{}

func (e Entity) StatementNode() Statement {
	return e
}

type ContainerKind int

const (
	UnknownContainerKind ContainerKind = iota
	PackageKind
	TogetherKind
	NamespaceKind
	FolderKind
	FrameKind
	RectangleKind
	DatabaseKind
	CloudKind
	NodeKind
)

type Container struct {
	Identifier     string
	Alias          string
	Kind           ContainerKind
	Stereotype     string
	Color          string
	Statements     []Statement
	LeadingTrivia  []tokenizer.Token
	TrailingTrivia []tokenizer.Token
}

var _ Statement = Container{}

func (c Container) StatementNode() Statement {
	return c
}

type Relationship struct {
	LHS       string
	RHS       string
	Direction DirectionKind

	TypeLHS RelationType
	TypeRHS RelationType
	MultLHS Cardinality
	MultRHS Cardinality

	// Arrow itself
	Body           rune // '-', '.'
	LArrow, RArrow rune // '<'/'>', etc.
	// Special case for left/righ arrow rune of relationship:
	// if the arrow is like '--|>', the '|' is used to distinguish it from '-->'
	// which would have end = '>'

	Label          string
	Attrs          []string
	LeadingTrivia  []tokenizer.Token
	TrailingTrivia []tokenizer.Token
}

var _ Statement = Relationship{}

func (r Relationship) StatementNode() Statement {
	return r
}

type ClassSeparator struct {
	// Optional label text
	Label string
	// Separator type. One of "-", "=", ".", "_"
	Type rune
}

var _ Member = ClassSeparator{}

func (cs ClassSeparator) MemberNode() Member {
	return cs
}

// type Field struct {
// 	// Its impossible to know what is the type of the field and what is its name
// 	// I will do go's way and first fill the name and then the type
// 	// Then during generation, use known types to find where the name actually is
// 	Raw        string
// 	Name       string
// 	Type       TypeRef
// 	Visibility VisibilityKind
// 	// Optional modifiers
// 	Modifiers []string
// 	// This shit is so fucked up...
// 	// What do you mean only way to distinguish between field and method is
// 	// by presence of parenthesis?
// }

type Field interface {
	Member
	FieldName() string
	FieldModifiers() []string
	FieldVisibility() VisibilityKind
}

// type Method struct {
// 	// Name and type will be inferred by assuming that '()' will be right after
// 	// the name
// 	Raw        string
// 	Name       string
// 	ReturnType []TypeRef // due to multiple return types languages, this is a slice
// 	Parameters []Parameter
// 	Modifiers  []string
// 	Visibility VisibilityKind
// }

type Method interface {
	Member
	MethodName() string
	MethodModifiers() []string
	MethodVisibility() VisibilityKind
}

type Diagram struct {
	Name           string
	Title          string
	Statements     []Statement
	LeadingTrivia  []tokenizer.Token
	TrailingTrivia []tokenizer.Token
}

// DirectionKind is a multi-purpose enum for representing literal directions
//
// Possible values:
//   - Left
//   - Right
//   - Top
//   - Bottom
type DirectionKind int

const (
	UnknownDirectionKind DirectionKind = iota
	Left
	Right
	Top
	Bottom
)

func (d DirectionKind) String() string {
	switch d {
	case Left:
		return "left"
	case Right:
		return "right"
	case Top:
		return "top"
	case Bottom:
		return "bottom"
	default:
		return "unknown"
	}
}

type Note struct {
	// note left of Class: note left of Class
	// where:
	// Text = note left of Class
	// Direction = DirectionKind(Left)
	// Target = Class
	// Color = ""
	Text           string
	Direction      DirectionKind
	Target         string
	Color          string
	Alias          string
	LeadingTrivia  []tokenizer.Token
	TrailingTrivia []tokenizer.Token
}

var _ Statement = Note{}

func (n Note) StatementNode() Statement { return n }

type DiagramBound struct {
	IsStart        bool
	Type           string // after '@start' or '@end' e.g. uml for 'startuml', gantt for 'startgantt', etc.
	ID             string // for identifying the diagram in files there there are more than one
	Name           string // in essence file name for the rendered diagram
	Opts           map[string]string
	LeadingTrivia  []tokenizer.Token
	TrailingTrivia []tokenizer.Token
}

var _ Statement = DiagramBound{}

func (d DiagramBound) StatementNode() Statement {
	return d
}
