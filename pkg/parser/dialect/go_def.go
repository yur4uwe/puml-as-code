package dialect

import (
	"strconv"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

// RefType describes the kind of a GoTypeRef node in the recursive type chain.
type RefType int

const (
	Pointer RefType = iota
	Slice
	Array
	Named // Terminal node — the base type name (e.g. "int", "MyStruct")
)

// GoTypeRef is a recursive linked-list representation of a Go type.
// Each node is either a modifier (Pointer, Slice, Array) wrapping a Base type,
// or a terminal Named node. For example, *[]User is represented as:
//
//	Pointer → Slice → Named("User")
//
// Qualified names (e.g. pkg.Type) are chained Named nodes:
//
//	Named("pkg") → Named("Type")
type GoTypeRef struct {
	Typ       RefType
	ArraySize int    // Only meaningful when Typ == Array
	Name      string // Only meaningful when Typ == Named
	Base      *GoTypeRef
}

// String reconstructs the Go type syntax from the recursive chain.
// The hasName flag tracks whether a Named node was already emitted,
// so that chained Named nodes (qualified names) are joined with '.'.
func (g *GoTypeRef) String() string {
	var sb strings.Builder
	var hasName bool
	for curr := g; curr != nil; curr = curr.Base {
		switch curr.Typ {
		case Slice:
			sb.WriteString("[]")
		case Pointer:
			sb.WriteString("*")
		case Array:
			sb.WriteRune('[')
			if curr.ArraySize <= 0 {
				sb.WriteRune('?')
			} else {
				sb.WriteString(strconv.Itoa(curr.ArraySize))
			}
			sb.WriteRune(']')
		case Named:
			if hasName {
				sb.WriteRune('.')
			}
			hasName = true
			sb.WriteString(curr.Name)
		}
	}
	return sb.String()
}

func PointerTo(base *GoTypeRef) *GoTypeRef {
	return &GoTypeRef{
		Typ:  Pointer,
		Base: base,
	}
}

func SliceOf(base *GoTypeRef) *GoTypeRef {
	return &GoTypeRef{
		Typ:  Slice,
		Base: base,
	}
}

func ArrayOf(size int, base *GoTypeRef) *GoTypeRef {
	return &GoTypeRef{
		Typ:       Array,
		ArraySize: size,
		Base:      base,
	}
}

func NamedRef(name string) *GoTypeRef {
	return &GoTypeRef{
		Typ:  Named,
		Name: name,
	}
}

// GoParameter represents a name+type pair used for both method parameters
// and return values. Name is empty for unnamed parameters/returns.
// Type is nil for untyped parameters (e.g. in Go's "a, b int" shorthand,
// "a" is initially untyped until backfilled).
type GoParameter struct {
	Name string
	Type *GoTypeRef
}

// GoField is the Go-specific implementation of [ast.Field].
// Consumers should type-assert from ast.Field to access Type.
type GoField struct {
	Name       string
	Type       *GoTypeRef
	Visibility ast.VisibilityKind
	Modifiers  []string
}

// FieldModifiers implements [ast.Field].
func (g *GoField) FieldModifiers() []string {
	return g.Modifiers
}

// FieldName implements [ast.Field].
func (g *GoField) FieldName() string {
	return g.Name
}

// FieldVisibility implements [ast.Field].
func (g *GoField) FieldVisibility() ast.VisibilityKind {
	return g.Visibility
}

// MemberNode implements [ast.Field].
func (g *GoField) MemberNode() ast.Member {
	return g
}

var (
	_ ast.Field  = (*GoField)(nil)
	_ ast.Member = (*GoField)(nil)
)

// GoMethod is the Go-specific implementation of [ast.Method].
// ReturnType uses GoParameter to support both named (e error) and
// unnamed (error) return values. Consumers should type-assert from
// ast.Method to access Parameters and ReturnType.
type GoMethod struct {
	Name       string
	ReturnType []GoParameter // Named returns have Name set, unnamed have Name empty
	Parameters []GoParameter
	Modifiers  []string
	Visibility ast.VisibilityKind
}

// MemberNode implements [ast.Method].
func (g *GoMethod) MemberNode() ast.Member {
	return g
}

// MethodModifiers implements [ast.Method].
func (g *GoMethod) MethodModifiers() []string {
	return g.Modifiers
}

// MethodName implements [ast.Method].
func (g *GoMethod) MethodName() string {
	return g.Name
}

// MethodVisibility implements [ast.Method].
func (g *GoMethod) MethodVisibility() ast.VisibilityKind {
	return g.Visibility
}

var (
	_ ast.Method = (*GoMethod)(nil)
	_ ast.Member = (*GoMethod)(nil)
)

type GoDialect struct{}

var _ Dialect = GoDialect{}
