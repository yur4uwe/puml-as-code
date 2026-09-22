package dialect

import "yur4uwe/pac/pkg/parser/ast"

type LaxField struct {
	Text       string             `json:",omitempty"`
	Visibility ast.VisibilityKind `json:",omitempty"`
	Modifiers  []string           `json:",omitempty"`
	ast.Trivia
}

// FieldModifiers implements [ast.Field].
func (l *LaxField) FieldModifiers() []string {
	return l.Modifiers
}

// FieldName implements [ast.Field].
func (l *LaxField) FieldName() string {
	return l.Text
}

// FieldVisibility implements [ast.Field].
func (l *LaxField) FieldVisibility() ast.VisibilityKind {
	return l.Visibility
}

// MemberNode implements [ast.Field].
func (l *LaxField) MemberNode() ast.Member {
	return l
}

var _ ast.Field = (*LaxField)(nil)

type LaxMethod struct {
	Text       string             `json:",omitempty"`
	Modifiers  []string           `json:",omitempty"`
	Visibility ast.VisibilityKind `json:",omitempty"`
	ast.Trivia
}

// MemberNode implements [ast.Method].
func (l *LaxMethod) MemberNode() ast.Member {
	return l
}

// MethodModifiers implements [ast.Method].
func (l *LaxMethod) MethodModifiers() []string {
	return l.Modifiers
}

// MethodName implements [ast.Method].
func (l *LaxMethod) MethodName() string {
	return l.Text
}

// MethodVisibility implements [ast.Method].
func (l *LaxMethod) MethodVisibility() ast.VisibilityKind {
	return l.Visibility
}

var _ ast.Method = (*LaxMethod)(nil)
