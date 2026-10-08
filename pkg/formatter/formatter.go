// Package formatter implements PlantUML formatter.
package formatter

import (
	"strings"
	"unicode"

	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/tokenizer"
)

type fState struct {
	source   []rune
	buf      strings.Builder
	disabled bool
	tabDepth int
	indent   int
}

func Format(src string) (string, error) {
	tree, err := parser.NewParser(dialect.LaxDialect{}).Parse(src)
	if err != nil {
		return "", err
	}

	state := &fState{
		source: []rune(src),
		buf:    strings.Builder{},
	}

	for _, stmt := range tree.Statements {
		state.formatStatement(stmt)
	}

	return state.buf.String(), nil
}

func (s *fState) formatStatement(stmt ast.Statement) {
	for _, comment := range stmt.GetLeadingTrivia() {
		// Check leading trivia for pragmas
		if strings.Contains(comment.Literal, "pac:fmt:off") {
			s.disabled = true
		} else if strings.Contains(comment.Literal, "pac:fmt:on") {
			s.disabled = false
		}
		s.emitComment(comment)
	}

	if s.disabled {
		s.emitRawSpan(stmt.Span())
	} else {
		s.prettyPrint(stmt)
	}

	for _, comment := range stmt.GetTrailingTrivia() {
		s.emitComment(comment)
	}
}

func (s *fState) prettyPrint(stmt ast.Statement) {
	switch st := stmt.(type) {
	case ast.DiagramBound:
		s.printDiagramBound(st)
	case ast.TextBlock:
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
		newlineIdx := strings.IndexByte(st.Text, '\n')
		if newlineIdx == -1 {
			alignments := mapBlockAlignments(st.VerticalAlignment, st.HorizontalAlignment)
			if alignments != "" {
				s.buf.WriteString(alignments)
				s.buf.WriteByte(' ')
			}
			kind := mapBlockKind(st.Kind)
			s.buf.WriteString(kind)
			s.buf.WriteByte(' ')
		} else {
			kind := mapBlockKind(st.Kind)
			alignments := mapBlockAlignments(st.VerticalAlignment, st.HorizontalAlignment)
			if alignments != "" {
				s.buf.WriteString(kind)
				s.buf.WriteByte(' ')
			}
			s.buf.WriteString(alignments)
			s.buf.WriteByte('\n')
		}

		s.buf.WriteString(st.Text)
		s.buf.WriteString("\n")
	case ast.UnhandledStatement:
		s.emitRawSpan(st.NodeSpan)
	case ast.Entity:
		s.printEntity(st)
	case ast.Container:
		s.printContainer(st)
	case ast.Relationship:
		s.printRelationship(st)
	case ast.Note:
		s.printNote(st)
	case ast.StyleDeclaration:
		s.printStyleDeclaration(st)
	case ast.StyleRule:
		s.printStyleRule(st)
	case ast.StyleBlock:
		s.printStyleBlock(st)
	case ast.SkinparamSetting:
		s.printSkinparamSetting(st)
	case ast.SkinparamBlock:
		s.printSkinparamBlock(st)
	case ast.GenericCommand:
		s.buf.WriteString(st.Name)
		s.buf.WriteString(" ")
		for i, arg := range st.Args {
			if i > 0 {
				s.buf.WriteString(" ")
			}
			s.buf.WriteString(arg)
		}
		s.buf.WriteString("\n")
	case ast.IncludeDirective:
		switch st.Kind {
		case ast.IncludeOnce:
			s.buf.WriteString("!include ")
		case ast.IncludeMany:
			s.buf.WriteString("!include_many ")
		}
		s.buf.WriteString(st.Path)
		if st.Tag != "" {
			s.buf.WriteString("!")
			s.buf.WriteString(st.Tag)
		}
		s.buf.WriteString("\n")
	case ast.ScaleCommand:
		s.printScaleCommand(st)
	case ast.VisibilityCommand:
		switch st.Kind {
		case ast.VisibilityCMDHide:
			s.buf.WriteString("hide ")
		case ast.VisibilityCMDShow:
			s.buf.WriteString("show ")
		case ast.VisibilityCMDRemove:
			s.buf.WriteString("remove ")
		case ast.VisibilityCMDRestore:
			s.buf.WriteString("restore ")
		}
		s.buf.WriteString(st.Target)
		s.buf.WriteString("\n")
	case ast.SetCommand:
		s.buf.WriteString(st.Key)
		s.buf.WriteString(" ")
		s.buf.WriteString(st.Value)
		s.buf.WriteString("\n")
	case ast.DirectionCommand:
		switch st.Direction {
		case ast.LeftToRightDirection:
			s.buf.WriteString("left to right\n")
		case ast.TopToBottomDirection:
			s.buf.WriteString("top to bottom\n")
		}
	default:
		panic("unimplemented")
	}
}

func (s *fState) emitRawSpan(span tokenizer.SourceSpan) {
	// find the previous newline to calculate original indent
	// it will help us find relative indents
	var origIdent uint = 0
	for i := span.Start.Offset; i > 0; i-- {
		if s.source[i] == '\n' {
			break
		} else {
			// Can be buggy if the source contains tabs
			// or other non-whitespace characters
			origIdent++
		}
	}
	if origIdent == span.Start.Offset {
		// Simply impossible for a successfully parsed file
		panic("impossible")
	}
	inputSpan := s.source[span.Start.Offset : span.End.Offset+1]
	inputLines := [][]rune{}
	localLineIndents := []int{}
	var currLine []rune
	currLineIndent := 0
	for i, r := range inputSpan {
		if len(currLine) == 0 && unicode.IsSpace(r) {
			currLineIndent++
			continue
		}

		if inputSpan[i] == '\n' || i == len(inputSpan)-1 {
			inputLines = append(inputLines, currLine)
			localLineIndents = append(localLineIndents, currLineIndent-int(origIdent)+1)
			currLineIndent = 0
			currLine = []rune{}
		} else {
			currLine = append(currLine, inputSpan[i])
		}
	}

	for i, line := range inputLines {
		if i == 0 {
			s.buf.WriteString(string(line))
			s.buf.WriteRune('\n')
			continue
		}

		sourceIndent := max(s.getIndent()+localLineIndents[i], 0)
		currTab := strings.Repeat(" ", sourceIndent)
		s.buf.WriteString(currTab)
		s.buf.WriteString(string(line))
		s.buf.WriteRune('\n')
	}
}

func (s *fState) getIndent() int {
	return s.indent * s.tabDepth
}

func (s *fState) emitComment(comment tokenizer.Token) {
	s.emitRawSpan(comment.Span)
}
