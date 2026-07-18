// Package keyword provides a list of keywords and their respective keyword kinds.
// To get the keyword kind of a token, use the Classify function.
package keyword

//go:generate enumer -type=KeywordKind -transform=lower
type KeywordKind byte

const (
	None KeywordKind = iota

	// Entity keywords
	Class
	Abstract
	Struct
	Interface
	Enum
	Annotation
	Record
	Dataclass
	Exception
	Protocol
	Action
	Entity
	Circle
	Diamond
	Metaclass
	Stereotype

	// Container keywords
	Together
	Package
	Folder
	Frame
	Rectangle
	Cloud
	Database
	Node
	Namespace

	// Other statements
	Note
	Skinparam
	Title

	// Commands
	Hide
	Show
	Remove
	Restore
	Set
	Scale

	// Misc
	End
	Direction
	Position
	Alias
)

func Classify(ident string) KeywordKind {
	switch ident {
	case "as":
		return Alias
	case "left", "right", "top", "bottom", "up", "down":
		return Direction
	case "of", "on":
		return Position
	}
	val, err := KeywordKindString(ident)
	if err != nil {
		return None
	}
	return val
}
