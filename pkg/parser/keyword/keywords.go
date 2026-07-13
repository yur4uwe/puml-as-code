// Package keyword provides a list of keywords and their respective keyword kinds.
// To get the keyword kind of a token, use the Classify function.
package keyword

//go:generate enumer -type=KeywordKind -transform=lower
type KeywordKind byte

const (
	None KeywordKind = iota

	// Entity keywords
	Class
	AbstractClass
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
	val, err := KeywordKindString(ident)
	if err != nil {
		return None
	}
	return val
}
