package dialect

import (
	"strconv"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

type RefType int

const (
	Pointer RefType = iota
	Slice
	Array
	Named
)

type GoTypeRef struct {
	Typ       RefType
	ArraySize int
	Name      string
	Base      *GoTypeRef
}

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

func (g *GoTypeRef) PointerTo(base *GoTypeRef) *GoTypeRef {
	g.Typ = Pointer
	g.Base = base
	return g
}

func (g *GoTypeRef) SliceOf(base *GoTypeRef) *GoTypeRef {
	g.Typ = Slice
	g.Base = base
	return g
}

func (g *GoTypeRef) ArrayOf(size int, base *GoTypeRef) *GoTypeRef {
	g.Typ = Array
	g.ArraySize = size
	g.Base = base
	return g
}

func (g *GoTypeRef) Named(name string) *GoTypeRef {
	g.Typ = Named
	g.Name = name
	return g
}

type GoParameter struct {
	Name string
	Type GoTypeRef
}

type GoField struct {
	Name       string
	Type       GoTypeRef
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

type GoMethod struct {
	Name       string
	ReturnType []GoTypeRef
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
