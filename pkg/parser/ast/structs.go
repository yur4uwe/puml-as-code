package ast

import (
	"strings"

	"yur4uwe/pac/pkg/tokenizer"
)

type Node interface {
	Span() tokenizer.SourceSpan
	GetLeadingTrivia() []tokenizer.Token
	GetTrailingTrivia() []tokenizer.Token
}

type Statement interface {
	Node
	StatementNode() Statement
}

type Member interface {
	Node
	MemberNode() Member
}

type BaseNode struct {
	NodeSpan       tokenizer.SourceSpan
	LeadingTrivia  []tokenizer.Token `json:",omitempty"`
	TrailingTrivia []tokenizer.Token `json:",omitempty"`
}

func (bn BaseNode) GetLeadingTrivia() []tokenizer.Token  { return bn.LeadingTrivia }
func (bn BaseNode) GetTrailingTrivia() []tokenizer.Token { return bn.TrailingTrivia }
func (bn BaseNode) Span() tokenizer.SourceSpan           { return bn.NodeSpan }

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
	BaseNode
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
	BaseNode
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

func (t TargetRef) FQN() string {
	var sb strings.Builder
	for _, pkg := range t.PackagePath {
		sb.WriteString(pkg)
		sb.WriteByte('.')
	}
	sb.WriteString(t.Entity)
	if t.Member != "" {
		sb.WriteString("::")
		sb.WriteString(t.Member)
	}
	return sb.String()
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
	BaseNode
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
	BaseNode
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

//go:generate enumer -type=NoteKind -transform=lower -trimprefix=Note -json
type NoteKind int

const (
	NoteUnknown NoteKind = iota

	// NoteInlineAlias represents a note defined with a text string and alias.
	//
	// Syntax (single-line):
	//   note "Text" as <alias> [#color]
	//
	// Example:
	//   note "Active connection" as N1
	//
	// Syntax (multiline):
	//   note as <alias> [#color]
	//     <text>
	//   end note
	//
	// Example:
	//   note as N2
	//     This is a floating note
	//   end note
	//
	NoteAlias

	// NoteTargeted represents a note positioned relative to an entity (or previous statement).
	//
	// Syntax (single-line):
	//   note <left|right|top|bottom> [of <target>] [#color] : <text>
	//
	// Syntax (multiline):
	//   note <left|right|top|bottom> [of <target>] [#color]
	//     <text>
	//   end note
	//
	// Example:
	//   note left of User : Authenticated via OAuth
	NoteTargeted

	// NoteLink represents a note attached to the preceding or active relationship link.
	//
	// Syntax (single-line):
	//   note on link [#color] : <text>
	//
	// Syntax (multiline):
	//   note on link [#color]
	//     <text>
	//   end note
	//
	// Example:
	//   note on link : TLS Encrypted
	NoteLink
)

type Note struct {
	Kind       NoteKind      `json:",omitempty"`
	Text       string        `json:",omitempty"`
	Direction  DirectionKind `json:",omitempty"`
	Target     *TargetRef    `json:",omitempty"`
	Color      string        `json:",omitempty"`
	Identifier string        `json:",omitempty"`
	BaseNode
}

var _ Statement = Note{}

func (n Note) StatementNode() Statement { return n }

type BoundOption struct {
	Key   string `json:",omitempty"`
	Value string `json:",omitempty"`
}

type BoundToolOptions struct {
	File    string        `json:",omitempty"`
	Caption string        `json:",omitempty"`
	Options []BoundOption `json:",omitempty"`
}

func (t BoundToolOptions) Get(key string) (string, bool) {
	for _, opt := range t.Options {
		if opt.Key == key {
			return opt.Value, true
		}
	}
	return "", false
}

type DiagramBound struct {
	IsStart      bool
	Type         string            `json:",omitempty"`
	ID           string            `json:",omitempty"`
	Params       []BoundOption     `json:",omitempty"`
	Tools        *BoundToolOptions `json:",omitempty"`
	TrailingName string            `json:",omitempty"`
	BaseNode
}

var _ Statement = DiagramBound{}

func (d DiagramBound) StatementNode() Statement {
	return d
}

func (d DiagramBound) DiagramName() string {
	if d.Tools != nil && d.Tools.File != "" {
		return d.Tools.File
	}
	return d.TrailingName
}

func (d DiagramBound) Name() string {
	return d.DiagramName()
}

func (d DiagramBound) GetParam(key string) (string, bool) {
	for _, p := range d.Params {
		if p.Key == key {
			return p.Value, true
		}
	}
	return "", false
}
