// Package formatter implements PlantUML formatter.
package formatter

import (
	"fmt"
	"strings"
	"unicode"

	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/tokenizer"
)

type formatterState struct {
	source   []rune
	buf      strings.Builder
	disabled bool
	tabDepth int
	indent   int

	// recollected from the AST
	packageSeparator string
}

func Format(src string, opts ...FormatOptions) (string, error) {
	var opt FormatOptions
	if len(opts) == 0 {
		opt = DefaultFormatOptions()
	}
	tree, err := parser.NewParser(dialect.LaxDialect{}).Parse(src)
	if err != nil {
		return "", err
	}

	state := &formatterState{
		source: []rune(src),
		buf:    strings.Builder{},
		indent: opt.IndentSize,

		packageSeparator: ".",
	}

	for _, node := range tree.Statements {
		formatNode(state, node)
	}

	return state.buf.String(), nil
}

func formatNode(s *formatterState, node ast.Node) {
	for _, comment := range node.GetLeadingTrivia() {
		// Check leading trivia for pragmas
		if strings.Contains(comment.Literal, "pac:fmt:off") {
			s.disabled = true
		} else if strings.Contains(comment.Literal, "pac:fmt:on") {
			s.disabled = false
		}
		s.emitTrivia(comment)
		s.buf.WriteByte('\n')
	}

	if s.disabled {
		s.emitRawSpan(node.Span())
	} else {
		switch node := node.(type) {
		case ast.Statement:
			s.formatStatement(node)
		case ast.Member:
			s.formatMember(node)
		}
	}

	s.emitBlockEndTrivia(node)
	s.buf.WriteString("\n")
}

func (s *formatterState) formatMember(member ast.Member) {
	switch mem := member.(type) {
	case ast.ClassSeparator:
		printClassSeparator(s, mem)
	case ast.Field:
		s.writeIndent()
		for _, modifier := range mem.FieldModifiers() {
			fmt.Fprintf(&s.buf, "{%s} ", modifier)
		}
		s.buf.WriteString(mem.FieldVisibility().String())
		s.buf.WriteString(mem.String())
	case ast.Method:
		s.writeIndent()
		for _, modifier := range mem.MethodModifiers() {
			fmt.Fprintf(&s.buf, "{%s} ", modifier)
		}
		s.buf.WriteString(mem.MethodVisibility().String())
		s.buf.WriteString(mem.String())
	default:
		panic("unreachable")
	}
	s.emitBlockEndTrivia(member)
}

func (s *formatterState) formatStatement(stmt ast.Statement) {
	switch st := stmt.(type) {
	case ast.DiagramBound:
		printDiagramBound(s, st)
	case ast.TextBlock:
		printTextBlock(s, st)
	case ast.UnhandledStatement:
		s.emitRawSpan(st.NodeSpan)
	case ast.Entity:
		printEntity(s, st)
	case ast.Container:
		printContainer(s, st)
	case ast.Relationship:
		printRelationship(s, st)
	case ast.Note:
		printNote(s, st)
	case ast.StyleDeclaration:
		// To emit semicolons or not, that is the question
		s.emitWithIndent(st.Property + ": " + st.Value)
	case ast.StyleRule:
		printStyleRule(s, st)
	case ast.StyleBlock:
		printStyleBlock(s, st)
	case ast.SkinparamSetting:
		printSkinparamSetting(s, st)
	case ast.SkinparamBlock:
		printSkinparamBlock(s, st)
	case ast.GenericCommand:
		s.buf.WriteString(st.Name)
		s.buf.WriteString(" ")
		for i, arg := range st.Args {
			if i > 0 {
				s.buf.WriteString(" ")
			}
			s.buf.WriteString(arg)
		}
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
	case ast.ScaleCommand:
		printScaleCommand(s, st)
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
	case ast.SetCommand:
		if st.Key == "separator" {
			s.packageSeparator = st.Value
		}
		s.emitWithIndent("set ")
		s.buf.WriteString(st.Key)
		s.buf.WriteString(" ")
		s.buf.WriteString(st.Value)
	case ast.DirectionCommand:
		switch st.Direction {
		case ast.LeftToRightDirection:
			s.buf.WriteString("left to right")
		case ast.TopToBottomDirection:
			s.buf.WriteString("top to bottom")
		}
		s.buf.WriteString(" direction")
	default:
		panic("unimplemented")
	}
}

func (s *formatterState) emitRawSpan(span tokenizer.SourceSpan) {
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
			if len(inputLines) > 1 {
				s.buf.WriteRune('\n')
			}
			continue
		}

		sourceIndent := max(s.getIndent()+localLineIndents[i], 0)
		currTab := strings.Repeat(" ", sourceIndent)
		s.buf.WriteString(currTab)
		s.buf.WriteString(string(line))
		if i < len(inputLines)-1 {
			s.buf.WriteRune('\n')
		}
	}
}
