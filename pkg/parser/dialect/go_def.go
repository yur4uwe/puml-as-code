package dialect

import (
	"strconv"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

// RefType describes the kind of a GoTypeRef node in the recursive type chain.
//
//go:generate enumer -type=RefType -transform=lower -trimprefix=Kind -json
type RefType int

const (
	KindPointer RefType = iota
	KindSlice
	KindArray
	KindNamed // Terminal node — the base type name (e.g. "int", "MyStruct")
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
	Typ       RefType    `json:",omitempty"`
	ArraySize int        `json:",omitempty"` // Only meaningful when Typ == Array
	Name      string     `json:",omitempty"` // Only meaningful when Typ == Named
	Base      *GoTypeRef `json:",omitempty"`
}

// String reconstructs the Go type syntax from the recursive chain.
// The hasName flag tracks whether a Named node was already emitted,
// so that chained Named nodes (qualified names) are joined with '.'.
func (g *GoTypeRef) String() string {
	var sb strings.Builder
	var hasName bool
	for curr := g; curr != nil; curr = curr.Base {
		switch curr.Typ {
		case KindSlice:
			sb.WriteString("[]")
		case KindPointer:
			sb.WriteString("*")
		case KindArray:
			sb.WriteRune('[')
			if curr.ArraySize <= 0 {
				sb.WriteRune('?')
			} else {
				sb.WriteString(strconv.Itoa(curr.ArraySize))
			}
			sb.WriteRune(']')
		case KindNamed:
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
		Typ:  KindPointer,
		Base: base,
	}
}

func SliceOf(base *GoTypeRef) *GoTypeRef {
	return &GoTypeRef{
		Typ:  KindSlice,
		Base: base,
	}
}

func ArrayOf(size int, base *GoTypeRef) *GoTypeRef {
	return &GoTypeRef{
		Typ:       KindArray,
		ArraySize: size,
		Base:      base,
	}
}

func NamedRef(name string) *GoTypeRef {
	return &GoTypeRef{
		Typ:  KindNamed,
		Name: name,
	}
}

// GoParameter represents a name+type pair used for both method parameters
// and return values. Name is empty for unnamed parameters/returns.
// Type is nil for untyped parameters (e.g. in Go's "a, b int" shorthand,
// "a" is initially untyped until backfilled).
type GoParameter struct {
	Name string     `json:",omitempty"`
	Type *GoTypeRef `json:",omitempty"`
}

// GoField is the Go-specific implementation of [ast.Field].
// Consumers should type-assert from ast.Field to access Type.
type GoField struct {
	Name       string             `json:",omitempty"`
	Type       *GoTypeRef         `json:",omitempty"`
	Visibility ast.VisibilityKind `json:",omitempty"`
	Modifiers  []string           `json:",omitempty"`
	ast.Trivia
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
	Name       string             `json:",omitempty"`
	ReturnType []GoParameter      `json:",omitempty"` // Named returns have Name set, unnamed have Name empty
	Parameters []GoParameter      `json:",omitempty"`
	Modifiers  []string           `json:",omitempty"`
	Visibility ast.VisibilityKind `json:",omitempty"`
	ast.Trivia
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

func (g *GoMethod) Signature() string {
	var sb strings.Builder
	sb.WriteRune('(')
	for i, param := range g.Parameters {
		if i > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(param.Name)
		sb.WriteRune(' ')
		sb.WriteString(param.Type.String())
	}
	sb.WriteRune(')')

	if len(g.ReturnType) == 0 {
		return sb.String()
	}

	sb.WriteRune(' ')

	hasParens := len(g.ReturnType) > 1 || g.ReturnType[0].Name != ""
	if hasParens {
		sb.WriteRune('(')
	}

	for i, param := range g.ReturnType {
		if i > 0 {
			sb.WriteRune(',')
		}
		if param.Name != "" {
			sb.WriteString(param.Name)
			sb.WriteRune(' ')
		}
		sb.WriteString(param.Type.String())
	}

	if hasParens {
		sb.WriteRune(')')
	}

	return sb.String()
}

var (
	_ ast.Method = (*GoMethod)(nil)
	_ ast.Member = (*GoMethod)(nil)
)

type GoDialect struct{}

var _ Dialect = GoDialect{}
