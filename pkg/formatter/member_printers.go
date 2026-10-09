package formatter

import (
	"fmt"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

func printClassSeparator(s *formatterState, sep ast.ClassSeparator) {
	s.writeIndent()
	if sep.Label == "" {
		s.buf.WriteString(strings.Repeat(string(sep.Type), sep.TypeCount))
		return
	}

	leftCount := sep.TypeCount / 2
	rightCount := sep.TypeCount - leftCount
	s.buf.WriteString(strings.Repeat(string(sep.Type), leftCount))
	fmt.Fprintf(&s.buf, " %s ", sep.Label)
	s.buf.WriteString(strings.Repeat(string(sep.Type), rightCount))
}

func printField(s *formatterState, field ast.Field) {
	s.writeIndent()
	for _, modifier := range field.FieldModifiers() {
		fmt.Fprintf(&s.buf, "{%s} ", modifier)
	}
	s.buf.WriteString(visibilitySymbol(field.FieldVisibility()))
	s.buf.WriteString(field.String())
}

func printMethod(s *formatterState, method ast.Method) {
	s.writeIndent()
	for _, modifier := range method.MethodModifiers() {
		fmt.Fprintf(&s.buf, "{%s} ", modifier)
	}
	s.buf.WriteString(visibilitySymbol(method.MethodVisibility()))
	s.buf.WriteString(method.String())
}
