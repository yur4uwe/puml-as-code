// Package formatter implements PlantUML formatter.
package formatter

import (
	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
)

func Format(src string) (string, error) {
	tree, err := parser.NewParser(dialect.LaxDialect{}).Parse(src)
	if err != nil {
		return "", err
	}

	for _, stmt := range tree.Statements {
		switch stmt.(type) {
		case ast.TitleDef:
		case ast.UnhandledStatement:
		}
	}
	panic("unimplemented")
}
