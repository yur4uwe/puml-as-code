package ast

type GenericCommand struct {
	Name string
	Args []string
}

var _ Statement = GenericCommand{}

type IncludeCommand struct {
	Path string
	Tag  string
}

var _ Statement = IncludeCommand{}

type ScaleCommand struct {
	Scale  float64
	Width  int // PlantUML allows "scale 200 width"
	Height int
	IsMax  bool // PlantUML also allows "scale max 200 width"
}

var _ Statement = ScaleCommand{}

type VisibilityCommandKind int

const (
	Unknown VisibilityCommandKind = iota
	Hide
	Show
	Remove
	Restore
)

type VisibilityCommand struct {
	Kind   VisibilityCommandKind // Hide, Show, Remove, Restore
	Target string                // "empty members", "class Name", "circle", etc.
}

var _ Statement = VisibilityCommand{}

type SetCommand struct {
	Key   string // e.g. separator
	Value string // e.g. .
}

var _ Statement = SetCommand{}

type DirectionCommandKind int

const (
	UnknownDirection DirectionCommandKind = iota
	LeftToRightDirection
	TopToBottomDirection
)

type DirectionCommand struct {
	Direction DirectionCommandKind
}

var _ Statement = DirectionCommand{}

func (d GenericCommand) StatementNode() Statement    { return d }
func (d IncludeCommand) StatementNode() Statement    { return d }
func (d ScaleCommand) StatementNode() Statement      { return d }
func (d VisibilityCommand) StatementNode() Statement { return d }
func (d SetCommand) StatementNode() Statement        { return d }
func (d DirectionCommand) StatementNode() Statement  { return d }
