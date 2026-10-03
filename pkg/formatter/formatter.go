// Package formatter implements PlantUML formatter.
package formatter

import (
	"strings"

	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/tokenizer"
)

type fState struct {
	disabled bool
	currTab  int
}

func Format(src string) (string, error) {
	tree, err := parser.NewParser(dialect.LaxDialect{}).Parse(src)
	if err != nil {
		return "", err
	}

	state := &fState{}

	for _, stmt := range tree.Statements {
		formatStatement(state, stmt)
	}

	return "Formatter connected", nil
}

func formatStatement(state *fState, stmt ast.Statement) {
	for _, comment := range stmt.GetLeadingTrivia() {
		// Check leading trivia for pragmas
		if strings.Contains(comment.Literal, "pac:fmt:off") {
			state.disabled = true
		} else if strings.Contains(comment.Literal, "pac:fmt:on") {
			state.disabled = false
		}
		emitComment(comment)
	}

	if state.disabled {
		emitRawSpan(stmt.Span())
	} else {
		prettyPrint(stmt)
	}

	for _, comment := range stmt.GetTrailingTrivia() {
		emitComment(comment)
	}
}

func prettyPrint(stmt ast.Statement) {
	panic("unimplemented")
}

func emitRawSpan(sourceSpan tokenizer.SourceSpan) {
	panic("unimplemented")
}

func emitComment(comment tokenizer.Token) {
	panic("unimplemented")
}
