package ast

import "yur4uwe/pac/pkg/tokenizer"

type TriviaHolder interface {
	GetLeadingTrivia() []tokenizer.Token
	GetTrailingTrivia() []tokenizer.Token
}

type Statement interface {
	TriviaHolder
	StatementNode() Statement
}

type Member interface {
	TriviaHolder
	MemberNode() Member
}

type Trivia struct {
	LeadingTrivia  []tokenizer.Token `json:",omitempty"`
	TrailingTrivia []tokenizer.Token `json:",omitempty"`
}

func (t Trivia) GetLeadingTrivia() []tokenizer.Token  { return t.LeadingTrivia }
func (t Trivia) GetTrailingTrivia() []tokenizer.Token { return t.TrailingTrivia }

//go:generate enumer -type=EntityKind -transform=lower -trimprefix=Entity -json
type EntityKind int

const (
	EntityUnknown EntityKind = iota
	EntityClass
	EntityAbstractClass
	EntityInterface
	EntityEnum
	EntityEntityClass
	EntityStruct
	EntityAnnotation
	EntityProtocol
	EntityCircle
	EntityDiamond
	EntityException
	EntityMetaclass
	EntityRecord
	EntityDataclass
	EntityStereotype // For the standalone "stereotype" keyword

	// Robustness (BCE)
	EntityBceEntity
	EntityBoundary
	EntityControl

	// Component / Mixed
	EntityActor
	EntityComponent
	EntityArtifact
)

func (k EntityKind) AllowsBody() bool {
	switch k {
	case EntityCircle, EntityDiamond:
		return false
	default:
		return true
	}
}

type Entity struct {
	Trivia
	Identifier string     `json:",omitempty"`
	Alias      string     `json:",omitempty"`
	Kind       EntityKind `json:",omitempty"`
	Stereotype string     `json:",omitempty"`
	Tags       []string   `json:",omitempty"`
	Generic    string     `json:",omitempty"`
	Color      string     `json:",omitempty"`
	Members    []Member   `json:",omitempty"`
}

var _ Statement = Entity{}

func (e Entity) StatementNode() Statement {
	return e
}

//go:generate enumer -type=ContainerKind -transform=lower -trimprefix=Container -json
type ContainerKind int

const (
	ContainerUnknown ContainerKind = iota
	ContainerPackage
	ContainerTogether
	ContainerNamespace
	ContainerFolder
	ContainerFrame
	ContainerRectangle
	ContainerDatabase
	ContainerCloud
	ContainerNode
)

type Container struct {
	Identifier string        `json:",omitempty"`
	Alias      string        `json:",omitempty"`
	Kind       ContainerKind `json:",omitempty"`
	Stereotype string        `json:",omitempty"`
	Tags       []string      `json:",omitempty"`
	Color      string        `json:",omitempty"`
	Statements []Statement   `json:",omitempty"`
	Trivia
}

var _ Statement = Container{}

func (c Container) StatementNode() Statement {
	return c
}

type TargetRef struct {
	PackagePath []string `json:",omitempty"`
	Entity      string   `json:",omitempty"`
	Member      string   `json:",omitempty"`
}

type Relationship struct {
	LHS       TargetRef
	RHS       TargetRef
	Direction DirectionKind `json:",omitempty"`

	TypeLHS RelationType `json:",omitempty"`
	TypeRHS RelationType `json:",omitempty"`
	MultLHS Cardinality
	MultRHS Cardinality

	// Arrow itself
	Body           rune // '-', '.'
	LArrow, RArrow rune `json:",omitempty"`
	// Special case for left/righ arrow rune of relationship:
	// if the arrow is like '--|>', the '|' is used to distinguish it from '-->'
	// which would have end = '>'

	Label string   `json:",omitempty"`
	Attrs []string `json:",omitempty"`
	Trivia
}

var _ Statement = Relationship{}

func (r Relationship) StatementNode() Statement {
	return r
}

type ClassSeparator struct {
	// Optional label text
	Label string `json:",omitempty"`
	// Separator type. One of "-", "=", ".", "_"
	Type rune
	Trivia
}

var _ Member = ClassSeparator{}

func (cs ClassSeparator) MemberNode() Member {
	return cs
}

type Field interface {
	Member
	FieldName() string
	FieldModifiers() []string
	FieldVisibility() VisibilityKind
}

type Method interface {
	Member
	MethodName() string
	MethodModifiers() []string
	MethodVisibility() VisibilityKind
}

type Diagram struct {
	Name       string      `json:",omitempty"`
	Title      string      `json:",omitempty"`
	Statements []Statement `json:",omitempty"`
}

//go:generate enumer -type=DirectionKind -transform=lower -trimprefix=Direction -json
type DirectionKind int

const (
	DirectionUnknown DirectionKind = iota
	DirectionLeft
	DirectionRight
	DirectionTop
	DirectionBottom
)

type Note struct {
	Text      string        `json:",omitempty"`
	Direction DirectionKind `json:",omitempty"`
	Target    TargetRef
	Color     string `json:",omitempty"`
	Alias     string `json:",omitempty"`
	Trivia
}

var _ Statement = Note{}

func (n Note) StatementNode() Statement { return n }

type DiagramBound struct {
	IsStart bool
	Type    string            `json:",omitempty"`
	ID      string            `json:",omitempty"`
	Name    string            `json:",omitempty"`
	Opts    map[string]string `json:",omitempty"`
	Trivia
}

var _ Statement = DiagramBound{}

func (d DiagramBound) StatementNode() Statement {
	return d
}
