package ast

type GenericCommand struct {
	Name string   `json:",omitempty"`
	Args []string `json:",omitempty"`
	Trivia
}

var _ Statement = GenericCommand{}

//go:generate enumer -type=IncludeKind -transform=lower -trimprefix=Include -json
type IncludeKind int

const (
	IncludeOnce IncludeKind = iota
	IncludeMany
)

type IncludeDirective struct {
	Path string      `json:",omitempty"`
	Tag  string      `json:",omitempty"`
	Kind IncludeKind `json:",omitempty"`
	Trivia
}

var _ Statement = IncludeDirective{}

type ScaleCommand struct {
	Scale  float64 `json:",omitempty"`
	Width  int     `json:",omitempty"` // PlantUML allows "scale 200 width"
	Height int     `json:",omitempty"`
	IsMax  bool    // PlantUML also allows "scale max 200 width"
	Trivia
}

var _ Statement = ScaleCommand{}

//go:generate enumer -type=VisibilityCommandKind -transform=lower -trimprefix=VisibilityCMD -json
type VisibilityCommandKind int

const (
	VisibilityCMDUnknown VisibilityCommandKind = iota
	VisibilityCMDHide
	VisibilityCMDShow
	VisibilityCMDRemove
	VisibilityCMDRestore
)

type VisibilityCommand struct {
	Kind   VisibilityCommandKind `json:",omitempty"` // Hide, Show, Remove, Restore
	Target string                `json:",omitempty"` // "empty members", "class Name", "circle", etc.
	Trivia
}

var _ Statement = VisibilityCommand{}

type SetCommand struct {
	Key   string `json:",omitempty"` // e.g. separator
	Value string `json:",omitempty"` // e.g. .
	Trivia
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
	Trivia
}

var _ Statement = DirectionCommand{}

func (d GenericCommand) StatementNode() Statement    { return d }
func (d IncludeDirective) StatementNode() Statement  { return d }
func (d ScaleCommand) StatementNode() Statement      { return d }
func (d VisibilityCommand) StatementNode() Statement { return d }
func (d SetCommand) StatementNode() Statement        { return d }
func (d DirectionCommand) StatementNode() Statement  { return d }
