package formatter

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

func printNote(s *formatterState, note ast.Note) {
	isMultiline := strings.Contains(note.Text, "\n")
	// Anstract limit where note is too long to be inline
	isTooLong := len(note.Text) > 60

	if isMultiline || isTooLong {
		printBlockNote(s, note)
	} else {
		printInlineNote(s, note)
	}
}

func printInlineNote(s *formatterState, note ast.Note) {
	s.buf.WriteString("note ")
	switch note.Kind {
	case ast.NoteAlias:
		s.buf.WriteByte('"')
		s.buf.WriteString(note.Text)
		s.buf.WriteByte('"')
		s.buf.WriteString(" as ")
		s.buf.WriteString(note.Identifier)
	case ast.NoteTargeted:
		s.buf.WriteString(note.Direction.String())
		if note.Target != nil {
			s.buf.WriteString(" of ")
			s.buf.WriteString(s.targetFQN(*note.Target))
		}
	case ast.NoteLink:
		if note.Direction != ast.DirectionUnknown {
			s.buf.WriteString(note.Direction.String())
			s.buf.WriteByte(' ')
		}
		s.buf.WriteString("on link")
	default:
		panic("unreachable")
	}
	if note.Color != "" {
		s.buf.WriteByte(' ')
		s.buf.WriteString(note.Color)
	}
	if note.Kind != ast.NoteAlias {
		s.buf.WriteString(" : ")
		s.buf.WriteString(note.Text)
	}
}

func printBlockNote(s *formatterState, note ast.Note) {
	s.buf.WriteString("note ")
	switch note.Kind {
	case ast.NoteAlias:
		s.buf.WriteString("as ")
		s.buf.WriteString(note.Identifier)
	case ast.NoteTargeted:
		s.buf.WriteString(note.Direction.String())
		if note.Target != nil {
			s.buf.WriteString(" of ")
			s.buf.WriteString(s.targetFQN(*note.Target))
		}
	case ast.NoteLink:
		if note.Direction != ast.DirectionUnknown {
			s.buf.WriteString(note.Direction.String())
			s.buf.WriteByte(' ')
		}
		s.buf.WriteString("on link")
	default:
		panic("unreachable")
	}
	if note.Color != "" {
		s.buf.WriteByte(' ')
		s.buf.WriteString(note.Color)
	}
	s.emitBlockStartTrivia(note)
	s.buf.WriteString("\n")
	s.buf.WriteString(note.Text)
	s.buf.WriteString("\n")
	s.buf.WriteString("end note")
}

func printDiagramBound(s *formatterState, bound ast.DiagramBound) {
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
}

func printScaleCommand(s *formatterState, scaleCmd ast.ScaleCommand) {
	s.buf.WriteString("scale ")
	if scaleCmd.IsMax {
		s.buf.WriteString("max ")
	}
	s.buf.WriteString(scaleCmd.Lhs)
	if scaleCmd.Sep != "" {
		if scaleCmd.Sep != "." {
			s.buf.WriteString(" ")
		}
		s.buf.WriteString(scaleCmd.Sep)
		if scaleCmd.Sep != "." {
			s.buf.WriteString(" ")
		}
		s.buf.WriteString(scaleCmd.Rhs)
	}
	if scaleCmd.Unit != "" {
		s.buf.WriteString(" ")
		s.buf.WriteString(scaleCmd.Unit)
	}
}

func printSkinparamBlock(s *formatterState, block ast.SkinparamBlock) {
	s.emitWithIndent(block.Keyword)
	s.buf.WriteByte(' ')
	if block.Name != "" {
		s.buf.WriteString(block.Name)
		s.buf.WriteByte(' ')
	}
	if block.Stereotype != "" {
		s.buf.WriteString("<<")
		s.buf.WriteString(block.Stereotype)
		s.buf.WriteString(">> ")
	}
	s.buf.WriteString("{")
	s.emitBlockStartTrivia(block)
	s.buf.WriteByte('\n')

	s.formatBlockStatements(block.Statements)

	s.emitWithIndent("}")
}

func printSkinparamSetting(s *formatterState, setting ast.SkinparamSetting) {
	s.writeIndent()
	if setting.Keyword != "" {
		s.buf.WriteString(setting.Keyword)
		s.buf.WriteByte(' ')
	}
	s.buf.WriteString(setting.Name)
	s.buf.WriteByte(' ')
	if setting.Stereotype != "" {
		s.buf.WriteString("<<")
		s.buf.WriteString(setting.Stereotype)
		s.buf.WriteString(">> ")
	}
	s.buf.WriteString(setting.Value)
}

func printStyleBlock(s *formatterState, block ast.StyleBlock) {
	s.buf.WriteString("<style>")
	s.emitBlockStartTrivia(block)
	s.buf.WriteByte('\n')

	s.formatBlockStatements(block.Statements)

	s.buf.WriteString("\n</style>")
}

func printStyleRule(s *formatterState, block ast.StyleRule) {
	s.writeIndent()
	for i, selector := range block.Selectors {
		if i < len(block.Selectors)-1 {
			s.buf.WriteString(", ")
		}
		s.buf.WriteString(selector)
	}
	s.buf.WriteString(" {")
	s.emitBlockStartTrivia(block)
	s.buf.WriteByte('\n')

	s.formatBlockStatements(block.Statements)

	s.emitWithIndent("}")
}

func printEntity(s *formatterState, ent ast.Entity) {
	if ent.Kind == ast.EntityUnknown {
		// This is an implicit entity either created by:
		// Foo : Bar()
		s.emitWithIndent(ent.Identifier)
		s.buf.WriteString(" : ")
		s.formatBlockMembers(ent.Members)
		return
	}

	// Otherwise, it is a named entity
	actualKindStr := ent.Kind.String()
	switch ent.Kind {
	case ast.EntityAbstractClass:
		actualKindStr = "abstract class"
	case ast.EntityEntityClass:
		actualKindStr = "entity"
	}
	s.emitWithIndent(actualKindStr)
	s.buf.WriteString(" ")
	if ent.Alias != "" {
		s.buf.WriteString(ent.Alias)
		s.buf.WriteString(" as ")
	}
	s.buf.WriteString(ent.Identifier)
	if ent.Generic != "" {
		s.buf.WriteString(" <")
		s.buf.WriteString(ent.Generic)
		s.buf.WriteString(">")
	}
	if ent.Stereotype != "" {
		s.buf.WriteString(" <<")
		s.buf.WriteString(ent.Stereotype)
		s.buf.WriteString(">>")
	}
	for _, tag := range ent.Tags {
		s.buf.WriteString(" $")
		s.buf.WriteString(tag)
	}
	if ent.Color != "" {
		s.buf.WriteString(" ")
		s.buf.WriteString(ent.Color)
	}
	s.buf.WriteString(" {")
	s.emitBlockStartTrivia(ent)
	s.buf.WriteByte('\n')

	s.formatBlockMembers(ent.Members)

	if ent.Kind != ast.EntityUnknown {
		s.emitWithIndent("}")
	}
}

func printContainer(s *formatterState, cont ast.Container) {
	if cont.Kind == ast.ContainerUnknown {
		// This is an implicit container either created by:
		// class pack.Foo { ... }
		// Or
		// pack.Foo : Bar()
		s.emitWithIndent(cont.Identifier)
		s.buf.WriteString(".")
		s.formatBlockStatements(cont.Statements)
		return
	}

	s.emitWithIndent(cont.Kind.String())
	s.buf.WriteString(" ")
	if cont.Alias != "" {
		s.buf.WriteString(cont.Alias)
		s.buf.WriteString(" as ")
	}
	s.buf.WriteString(cont.Identifier)
	if cont.Stereotype != "" {
		s.buf.WriteString(" <<")
		s.buf.WriteString(cont.Stereotype)
		s.buf.WriteString(">>")
	}
	// For 0 tags either way nothing is printed
	for _, tag := range cont.Tags {
		s.buf.WriteString(" $")
		s.buf.WriteString(tag)
	}
	if cont.Color != "" {
		s.buf.WriteString(" ")
		s.buf.WriteString(cont.Color)
	}
	if len(cont.Statements) == 0 {
		s.emitBlockStartTrivia(cont)
		return
	}
	s.buf.WriteString(" {")
	s.emitBlockStartTrivia(cont)
	s.buf.WriteByte('\n')

	s.formatBlockStatements(cont.Statements)

	if cont.Kind != ast.ContainerUnknown {
		s.emitWithIndent("}")
	}
}

func printRelationship(s *formatterState, rel ast.Relationship) {
	s.emitWithIndent(s.targetFQN(rel.LHS))
	s.buf.WriteString(" ")
	if rel.MultLHS != ast.UnknownCardinality {
		s.buf.WriteString(rel.MultLHS.String())
	}
	s.buf.WriteString(" ")
	var lbrCount, rbrCount int
	lbrCount = rel.BodyCount / 2
	rbrCount = rel.BodyCount - lbrCount

	switch rel.LArrow {
	case '<', '}', 'o', 'x', '*', '+', '^', '#':
		s.buf.WriteRune(rel.LArrow)
	case '|':
		s.buf.WriteString("<|")
	default:
	}
	for range lbrCount {
		s.buf.WriteRune(rel.Body)
	}
	if rel.Direction != ast.DirectionUnknown {
		s.buf.WriteString(rel.Direction.String())
	}
	if len(rel.Attrs) > 0 {
		s.buf.WriteString("[")
		for i, attr := range rel.Attrs {
			if i > 0 {
				s.buf.WriteString(",")
			}
			s.buf.WriteString(attr)
		}
		s.buf.WriteString("]")
	}

	for range rbrCount {
		s.buf.WriteRune(rel.Body)
	}
	switch rel.RArrow {
	case '>', '{', 'o', 'x', '*', '+', '^', '#':
		s.buf.WriteRune(rel.RArrow)
	case '|':
		s.buf.WriteString("|>")
	default:
	}

	s.buf.WriteString(" ")
	if rel.MultRHS != ast.UnknownCardinality {
		s.buf.WriteString(rel.MultRHS.String())
	}
	s.buf.WriteString(" ")
	s.buf.WriteString(s.targetFQN(rel.RHS))
	if rel.Label != "" {
		s.buf.WriteString(" : ")
		s.buf.WriteString(rel.Label)
	}
}

func printTextBlock(s *formatterState, block ast.TextBlock) {
	mapBlockAlignments := func(va, ha string) string {
		alignment := ""
		if va != "" {
			alignment = va
		}
		if ha != "" {
			if alignment != "" {
				alignment += " "
			}
			alignment += ha
		}
		return alignment
	}
	mapBlockKind := func(k ast.TextBlockKind) string {
		switch k {
		case ast.BlockLegend:
			return "legend"
		case ast.BlockHeader:
			return "header"
		case ast.BlockFooter:
			return "footer"
		case ast.BlockTitle:
			return "title"
		default:
			panic("unreachable")
		}
	}
	newlineIdx := strings.IndexByte(block.Text, '\n')
	if newlineIdx == -1 {
		alignments := mapBlockAlignments(block.VerticalAlignment, block.HorizontalAlignment)
		if alignments != "" {
			s.buf.WriteString(alignments)
			s.buf.WriteByte(' ')
		}
		kind := mapBlockKind(block.Kind)
		s.buf.WriteString(kind)
		s.buf.WriteByte(' ')
	} else {
		kind := mapBlockKind(block.Kind)
		alignments := mapBlockAlignments(block.VerticalAlignment, block.HorizontalAlignment)
		if alignments != "" {
			s.buf.WriteString(kind)
			s.buf.WriteByte(' ')
		}
		s.buf.WriteString(alignments)
		s.buf.WriteByte('\n')
	}

	s.buf.WriteString(block.Text)
}
