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

	if len(bound.Params) > 0 {
		s.buf.WriteString("(")
		for i, param := range bound.Params {
			if i > 0 {
				s.buf.WriteString(", ")
			}
			s.buf.WriteString(param.Key)
			s.buf.WriteString("=")
			s.buf.WriteString(param.Value)
		}
		s.buf.WriteString(")")
	} else if bound.ID != "" {
		s.buf.WriteString("(id=")
		s.buf.WriteString(bound.ID)
		s.buf.WriteString(")")
	}

	if bound.Tools != nil {
		s.buf.WriteString("{")
		s.buf.WriteString(bound.Tools.File)
		if bound.Tools.Caption != "" {
			s.buf.WriteString(", ")
			s.buf.WriteString(bound.Tools.Caption)
		}
		for _, opt := range bound.Tools.Options {
			s.buf.WriteString(", ")
			s.buf.WriteString(opt.Key)
			s.buf.WriteString("=")
			s.buf.WriteString(opt.Value)
		}
		s.buf.WriteString("}")
	} else if bound.TrailingName != "" {
		s.buf.WriteString(" ")
		s.buf.WriteString(bound.TrailingName)
	}

	s.buf.WriteString("\n")
}

func (s *fState) printScaleCommand(st ast.ScaleCommand) {
	s.buf.WriteString("scale ")
	if st.IsMax {
		s.buf.WriteString("max ")
	}
	s.buf.WriteString(st.Lhs)
	if st.Sep != "" {
		if st.Sep != "." {
			s.buf.WriteString(" ")
		}
		s.buf.WriteString(st.Sep)
		if st.Sep != "." {
			s.buf.WriteString(" ")
		}
		s.buf.WriteString(st.Rhs)
	}
	if st.Unit != "" {
		s.buf.WriteString(" ")
		s.buf.WriteString(st.Unit)
	}
	s.buf.WriteString("\n")
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
