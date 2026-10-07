package ast

type GenericCommand struct {
	Name string   `json:",omitempty"`
	Args []string `json:",omitempty"`
	BaseNode
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
	BaseNode
}

var _ Statement = IncludeDirective{}

type ScaleCommand struct {
	IsMax bool
	Lhs   string `json:",omitempty"` // "1", "2", "200"
	Sep   string `json:",omitempty"` // "", ".", "/", "*", "x"
	Rhs   string `json:",omitempty"` // "5", "3", "100"; empty when Sep is empty
	Unit  string `json:",omitempty"` // "", "width", "height"
	BaseNode
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
	BaseNode
}

var _ Statement = VisibilityCommand{}

type SetCommand struct {
	Key   string `json:",omitempty"` // e.g. separator
	Value string `json:",omitempty"` // e.g. .
	BaseNode
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
	BaseNode
}

var _ Statement = DirectionCommand{}

func (d GenericCommand) StatementNode() Statement    { return d }
func (d IncludeDirective) StatementNode() Statement  { return d }
func (d ScaleCommand) StatementNode() Statement      { return d }
func (d VisibilityCommand) StatementNode() Statement { return d }
func (d SetCommand) StatementNode() Statement        { return d }
func (d DirectionCommand) StatementNode() Statement  { return d }
