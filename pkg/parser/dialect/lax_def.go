package dialect

import "yur4uwe/pac/pkg/parser/ast"

type LaxField struct {
	Text string `json:",omitempty"`
}

func (l *LaxField) FieldName() string {
	return l.Text
}

func (l *LaxField) String() string {
	return l.Text
}

var _ ast.Field = (*LaxField)(nil)

type LaxMethod struct {
	Text string `json:",omitempty"`
}

func (l *LaxMethod) MethodName() string {
	return l.Text
}

func (l *LaxMethod) String() string {
	return l.Text
}

var _ ast.Method = (*LaxMethod)(nil)
