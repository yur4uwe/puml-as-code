package gogenerator

import (
	"fmt"
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

func RenderRelationshipType(rel *ast.Relationship, from bool) string {
	var maxMult int
	if from {
		maxMult = rel.MultiplicityTo.Max
	} else {
		maxMult = rel.MultiplicityFrom.Max
	}

	name := ""
	if from {
		name = rel.To.Name
	} else {
		name = rel.From.Name
	}

	targetType := RenderType(ast.CustomType(name))
	switch rel.Type {
	case ast.Composition:
		if maxMult == -1 {
			return "[]" + targetType
		}
		if maxMult > 1 {
			return fmt.Sprintf("[%d]", maxMult) + targetType
		}
		return targetType
	case ast.Association, ast.Aggregation:
		if maxMult == -1 {
			return "[]*" + trimPointer(targetType)
		}
		if maxMult > 1 {
			return fmt.Sprintf("[%d]*", maxMult) + targetType
		}
		return "*" + trimPointer(targetType)
	case ast.Inheritance:
		return targetType
	default:
		return ""
	}

}

func trimPointer(s string) string {
	if len(s) > 0 && s[0] == '*' {
		return s[1:]
	}
	return s
}
