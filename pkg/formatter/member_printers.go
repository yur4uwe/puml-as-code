package formatter

import (
	"yur4uwe/pac/pkg/parser/ast"
)

func printClassSeparator(s *formatterState, sep ast.ClassSeparator) {
	if sep.Label == "" {
		s.writeIndent()
		for range sep.TypeCount {
			s.buf.WriteRune(sep.Type)
		}
		return
	}

	leftCount := sep.TypeCount / 2
	rightCount := sep.TypeCount - leftCount
	s.writeIndent()
	for range leftCount {
		s.buf.WriteRune(sep.Type)
	}
	s.buf.WriteString(sep.Label)
	for range rightCount {
		s.buf.WriteRune(sep.Type)
	}
}
