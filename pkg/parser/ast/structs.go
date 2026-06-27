package ast

type Statement interface {
	StatementNode() any
}

type Member interface {
	MemberNode() any
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
	Identifier string
	Alias      string
	Kind       EntityKind
	Stereotype string
	Color      string
	Members    []Member
}

var (
	_ Statement = Entity{}
	_ Member    = Entity{}
)

func (e Entity) StatementNode() any {
	return e
}

func (e Entity) MemberNode() any {
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
	Name       string
	Alias      string
	Kind       ContainerKind
	Stereotype string
	Color      string
	Statements []Statement
}

var _ Statement = Container{}

func (c Container) StatementNode() any {
	return c
}

type Relationship struct {
	From string
	To   string
	Type RelationType

	MultiplicityFrom Cardinality
	MultiplicityTo   Cardinality

	Label string
}

var _ Statement = Relationship{}

func (r Relationship) StatementNode() any {
	return r
}

type ClassSeparator struct {
	// Optional label text
	Label string
	// Separator type. One of "-", "=", ".", "_"
	Type string
}

var _ Member = ClassSeparator{}

func (cs ClassSeparator) MemberNode() any {
	return cs
}

type Field struct {
	// Its impossible to know what is the type of the field and what is its name
	// I will do go's way and first fill the name and then the type
	// Then during generation, use known types to find where the name actually is
	Raw        string
	Name       string
	Type       TypeRef
	Visibility VisibilityKind
	// Optional modifiers
	Modifiers []string
	// This shit is so fucked up...
	// What do you mean only way to distinguish between field and method is
	// by presence of parenthesis?
}

var _ Member = Field{}

func (f Field) MemberNode() any { return f }

type Method struct {
	// Name and type will be inferred by assuming that '()' will be right after
	// the name
	Raw        string
	Name       string
	ReturnType TypeRef
	Parameters []Parameter
	Modifiers  []string
	Visibility VisibilityKind
}

var _ Member = Method{}

func (m Method) MemberNode() any { return m }

type Diagram struct {
	Name       string
	Title      string
	Statements []Statement
}

type DirectionKind int

const (
	UnknownDirectionKind DirectionKind = iota
	Left
	Right
	Top
	Bottom
)

type Note struct {
	// note left of Class: note left of Class
	// where:
	// Text = note left of Class
	// Direction = DirectionKind(Left)
	// Target = Class
	// Color = ""
	Text      string
	Direction DirectionKind
	Target    string
	Color     string
}

var (
	_ Statement = Note{}
	_ Member    = Note{}
)

func (n Note) StatementNode() any { return n }
func (n Note) MemberNode() any    { return n }

type DiagramBound struct {
	IsStart bool
	Type    string // after '@start' or '@end' e.g. uml for 'startuml', gantt for 'startgantt', etc.
	ID      string // for identifying the diagram in files there there are more than one
	Name    string // in essence file name for the rendered diagram
	Opts    map[string]string
}

var _ Statement = DiagramBound{}

func (d DiagramBound) StatementNode() any {
	return d
}
