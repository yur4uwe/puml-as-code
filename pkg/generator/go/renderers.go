package gogenerator

import (
	"strings"
	"yur4uwe/pac/pkg/parser/ast"
)

func RenderType(t ast.TypeRef) string {
	base := ""
	switch t.Kind {
	case ast.Int:
		base = "int"
	case ast.String:
		base = "string"
	case ast.Float:
		base = "float64"
	case ast.Bool:
		base = "bool"
	case ast.Void:
		base = ""
	case ast.Custom:
		if t.Name == "" {
			base = "any"
		} else {
			base = strings.TrimSpace(t.Name)
		}
	default:
		base = "any"
	}
	return base
}

func trimPointer(s string) string {
	if len(s) > 0 && s[0] == '*' {
		return s[1:]
	}
	return s
}
