package ast

type GenericDirective struct {
	Name string
	Args []string
}

var _ Statement = GenericDirective{}

type IncludeDirective struct {
	Path string
	Tag  string
}

var _ Statement = IncludeDirective{}

type ScaleDirective struct {
	Scale  float64
	Width  int // PlantUML allows "scale 200 width"
	Height int
	IsMax  bool // PlantUML also allows "scale max 200 width"
}

var _ Statement = ScaleDirective{}

type VisibilityDirectiveKind int

const (
	Unknown VisibilityDirectiveKind = iota
	Hide
	Show
	Remove
	Restore
)

type VisibilityDirective struct {
	Kind   VisibilityDirectiveKind // Hide, Show, Remove, Restore
	Target string                  // "empty members", "class Name", "circle", etc.
}

var _ Statement = VisibilityDirective{}

type SetDirective struct {
	Key   string // e.g. separator
	Value string // e.g. ::
}

var _ Statement = SetDirective{}

func (d GenericDirective) StatementNode() any    { return d }
func (d IncludeDirective) StatementNode() any    { return d }
func (d ScaleDirective) StatementNode() any      { return d }
func (d VisibilityDirective) StatementNode() any { return d }
func (d SetDirective) StatementNode() any        { return d }
