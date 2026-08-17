package ast

type GenericCommand struct {
	Name string   `json:",omitempty"`
	Args []string `json:",omitempty"`
}

var _ Statement = GenericCommand{}

type IncludeCommand struct {
	Path string `json:",omitempty"`
	Tag  string `json:",omitempty"`
}

var _ Statement = IncludeCommand{}

type ScaleCommand struct {
	Scale  float64 `json:",omitempty"`
	Width  int     `json:",omitempty"` // PlantUML allows "scale 200 width"
	Height int     `json:",omitempty"`
	IsMax  bool    // PlantUML also allows "scale max 200 width"
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
	Kind   VisibilityCommandKind `json:",omitempty"` // Hide, Show, Remove, Restore
	Target string                `json:",omitempty"` // "empty members", "class Name", "circle", etc.
}

var _ Statement = VisibilityCommand{}

type SetCommand struct {
	Key   string `json:",omitempty"` // e.g. separator
	Value string `json:",omitempty"` // e.g. .
}

var _ Statement = SetCommand{}

type DirectionCommandKind int

const (
	UnknownDirection DirectionCommandKind = iota
	LeftToRightDirection
	TopToBottomDirection
)

type DirectionCommand struct {
	Direction DirectionCommandKind `json:",omitempty"`
}

var _ Statement = DirectionCommand{}

func (d GenericCommand) StatementNode() Statement    { return d }
func (d IncludeCommand) StatementNode() Statement    { return d }
func (d ScaleCommand) StatementNode() Statement      { return d }
func (d VisibilityCommand) StatementNode() Statement { return d }
func (d SetCommand) StatementNode() Statement        { return d }
func (d DirectionCommand) StatementNode() Statement  { return d }
