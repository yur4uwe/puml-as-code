// Package formatter implements PlantUML formatter.
package formatter

import (
	"errors"
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
	useTabs  bool

	lastNode ast.Node
	rawNodes map[ast.Node]bool
	partial  bool

	// recollected from the AST
	packageSeparator string
}

func Format(src string, opts ...FormatOptions) (string, error) {
	return format(src, false, opts)
}

// FormatPartial formats recoverable regions of an incomplete diagram. It
// returns the parser diagnostics alongside the text. If the parser could not
// identify a diagram, it returns the input unchanged alongside the error.
func FormatPartial(src string, opts ...FormatOptions) (string, error) {
	return format(src, true, opts)
}

func format(src string, allowPartial bool, opts []FormatOptions) (string, error) {
	var opt FormatOptions
	if len(opts) == 0 {
		opt = DefaultFormatOptions()
	} else {
		opt = opts[0]
	}
	p := parser.NewParser(dialect.LaxDialect{})
	if allowPartial {
		p.WithTargetID("0")
	}
	tree, err := p.Parse(src)
	if !allowPartial && err != nil {
		return "", err
	}
	if tree == nil {
		return src, err
	}

	state := &formatterState{
		source:  []rune(src),
		buf:     strings.Builder{},
		indent:  opt.IndentSize,
		useTabs: opt.UseTabs,

		packageSeparator: ".",
	}
	if allowPartial {
		// Keep text outside the first diagram intact, even when the parser
		// cannot represent it (e.g. a second diagram or a leading comment).
		first := tree.Statements[0].Span().Start.Offset
		last := tree.Statements[len(tree.Statements)-1]
		if strings.TrimSpace(string(state.source[:first])) != "" {
			return src, err
		}
		end := last.Span().End.Offset
		for _, comment := range last.GetTrailingTrivia() {
			end = max(end, comment.Span.End.Offset)
		}
		if strings.TrimSpace(string(state.source[end:])) != "" {
			return src, err
		}
		state.partial = true
		state.rawNodes = make(map[ast.Node]bool)
		for _, diag := range tree.Diagnostics {
			if diag.Node == nil {
				continue
			}
			// Structural blocks have explicit delimiter positions, so their
			// valid children can still be formatted when a closer is missing.
			if block, ok := diag.Node.(ast.Block); ok && !block.IsClosed() && errors.Is(diag, ast.ErrUnclosedBlock) {
				continue
			}
			state.rawNodes[diag.Node] = true
		}
	}

	for _, node := range tree.Statements {
		formatNode(state, node)
	}

	return state.buf.String(), err
}

func formatNode(s *formatterState, node ast.Node) {
	if s.lastNode != nil && needsBlankLine(s.lastNode, node) {
		s.buf.WriteByte('\n')
	}

	leadingTrivia := node.GetLeadingTrivia()
	for _, comment := range leadingTrivia {
		// Check leading trivia for pragmas
		if strings.Contains(comment.Text, "pac:fmt:off") {
			s.disabled = true
		} else if strings.Contains(comment.Text, "pac:fmt:on") {
			s.disabled = false
		}
		s.emitTrivia(comment)
		s.buf.WriteByte('\n')
	}

	if s.disabled || s.rawNodes[node] {
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
	if buf := s.buf.String(); !strings.HasSuffix(buf, "\n") {
		s.buf.WriteByte('\n')
	}

	s.lastNode = node
}

func (s *formatterState) formatMember(member ast.Member) {
	switch mem := member.(type) {
	case *ast.ClassSeparator:
		printClassSeparator(s, mem)
	case *ast.FieldDeclaration:
		printField(s, mem)
	case *ast.MethodDeclaration:
		printMethod(s, mem)
	default:
		if member != nil {
			s.emitRawSpan(member.Span())
		}
	}
}

func (s *formatterState) formatStatement(stmt ast.Statement) {
	switch st := stmt.(type) {
	case *ast.DiagramBound:
		printDiagramBound(s, st)
	case *ast.TextBlock:
		printTextBlock(s, st)
	case *ast.UnhandledStatement:
		s.emitRawSpan(st.NodeSpan)
	case *ast.Entity:
		printEntity(s, st)
	case *ast.InlineMember:
		printInlineMember(s, st)
	case *ast.Container:
		printContainer(s, st)
	case *ast.Relationship:
		printRelationship(s, st)
	case *ast.Note:
		printNote(s, st)
	case *ast.StyleDeclaration:
		s.emitWithIndent(st.Property)
		s.buf.WriteString(": ")
		s.buf.WriteString(st.Value)
	case *ast.StyleRule:
		printStyleRule(s, st)
	case *ast.StyleBlock:
		printStyleBlock(s, st)
	case *ast.SkinparamSetting:
		printSkinparamSetting(s, st)
	case *ast.SkinparamBlock:
		printSkinparamBlock(s, st)
	case *ast.GenericCommand:
		printGenericCommand(s, st)
	case *ast.IncludeDirective:
		printIncludeDirective(s, st)
	case *ast.ScaleCommand:
		printScaleCommand(s, st)
	case *ast.VisibilityCommand:
		printVisibilityCommand(s, st)
	case *ast.SetCommand:
		printSetCommand(s, st)
	case *ast.DirectionCommand:
		printDirectionCommand(s, st)
	default:
		s.emitRawSpan(stmt.Span())
	}
}

func (s *formatterState) emitRawSpan(span tokenizer.SourceSpan) {
	if span.Start.Offset >= span.End.Offset || span.End.Offset > uint(len(s.source)) {
		return
	}
	// Find the previous newline to calculate original indent for relative indentation
	var origIndent uint
	isLeading := true
	if span.Start.Offset > 0 {
		for i := int(span.Start.Offset) - 1; i >= 0; i-- {
			if s.source[i] == '\n' {
				break
			}
			if !unicode.IsSpace(s.source[i]) {
				// Find the actual indent for lines where the statement emitted with
				// span is preceded by other statement
				isLeading = false
				origIndent = 0
			} else {
				origIndent++
			}
		}
	}

	inputSpan := s.source[span.Start.Offset:span.End.Offset]
	var inputLines [][]rune
	var localLineIndents []int
	var currLine []rune
	currLineIndent := 0

	// Collect lines stripped of leading spaces and
	// line-wise indent depth for each line to reproduce correct indents
	// for non-first lines
	for _, r := range inputSpan {
		if r != '\n' && len(currLine) == 0 && unicode.IsSpace(r) {
			currLineIndent++
			continue
		}

		if r == '\n' {
			inputLines = append(inputLines, currLine)
			localLineIndents = append(localLineIndents, currLineIndent-int(origIndent))
			currLineIndent = 0
			currLine = nil
		} else {
			currLine = append(currLine, r)
		}
	}
	inputLines = append(inputLines, currLine)
	localLineIndents = append(localLineIndents, currLineIndent-int(origIndent))

	for i, line := range inputLines {
		if i == 0 {
			if isLeading {
				s.writeIndent()
			}
			s.buf.WriteString(string(line))
			if len(inputLines) > 1 {
				s.buf.WriteRune('\n')
			}
			continue
		}

		var currTab string
		if s.useTabs {
			currTab = strings.Repeat("\t", s.tabDepth)
		} else {
			sourceIndent := max(s.getIndent()+localLineIndents[i], 0)
			currTab = strings.Repeat(" ", sourceIndent)
		}
		s.buf.WriteString(currTab)
		s.buf.WriteString(string(line))
		if i < len(inputLines)-1 {
			s.buf.WriteRune('\n')
		}
	}
}
