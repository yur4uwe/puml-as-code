package formatter

import (
	"yur4uwe/pac/pkg/parser/ast"
)

func (s *fState) printNote(note ast.Note) {
	s.buf.WriteString("note ")
	s.buf.WriteString(note.Text)
	s.buf.WriteString("\n")
}

func (s *fState) printDiagramBound(bound ast.DiagramBound) {
	if bound.IsStart {
		s.buf.WriteString("@start")
	} else {
		s.buf.WriteString("@end")
	}
	s.buf.WriteString(bound.Type)
	if bound.ID != "" {
		s.buf.WriteString("(id=")
		s.buf.WriteString(bound.ID)
		s.buf.WriteString(")")
	}
	if bound.IsStart {
		s.buf.WriteString("\n")
	}
}

func (s *fState) printScaleCommand(st ast.ScaleCommand) {
	s.buf.WriteString("scale ")
	if st.IsMax {
		s.buf.WriteString("max ")
	}
	panic("unimplemented")
}

func (s *fState) printSkinparamBlock(st ast.SkinparamBlock) {
	panic("unimplemented")
}

func (s *fState) printSkinparamSetting(st ast.SkinparamSetting) {
	panic("unimplemented")
}

func (s *fState) printStyleBlock(st ast.StyleBlock) {
	panic("unimplemented")
}

func (s *fState) printStyleRule(st ast.StyleRule) {
	panic("unimplemented")
}

func (s *fState) printStyleDeclaration(st ast.StyleDeclaration) {
	panic("unimplemented")
}

func (s *fState) printEntity(st ast.Entity) {
	panic("unimplemented")
}

func (s *fState) printContainer(st ast.Container) {
	if st.Kind == ast.ContainerUnknown {
		// This is an implicit container either created by:
		// class pack.Foo { ... }
		// Or
		// pack.Foo : Bar()
	}
	panic("unimplemented")
}

func (s *fState) printRelationship(st ast.Relationship) {
	panic("unimplemented")
}
