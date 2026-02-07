package gogenerator

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/types"
)

type GoCodeGenerator struct{}

var _ types.CodeGenerator = GoCodeGenerator{}

func (GoCodeGenerator) GenerateFromClassDiagram(diagram *ast.ClassDiagram) (string, error) {
	if diagram == nil {
		return "", fmt.Errorf("nil diagram")
	}

	var buf bytes.Buffer
	buf.WriteString("package generated\n\n")

	// deterministic ordering
	names := make([]string, 0, len(diagram.Classes))
	for n := range diagram.Classes {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		c := diagram.Classes[name]
		if c == nil {
			continue
		}

		// struct
		buf.WriteString(fmt.Sprintf("type %s struct {\n", exportName(c.Name, ast.Public)))

		rels, ok := diagram.Relationships.GetRelationships(c.Name)
		if !ok {
			goto attributes
		}

		for _, rel := range rels {
			if rel == nil || rel.From == nil || rel.To == nil {
				continue
			}
			// irrelevant relationship
			if rel.From.Name != c.Name && rel.To.Name != c.Name {
				continue
			}

			var targetName string
			if rel.From.Name == c.Name {
				targetName = rel.To.Name
			} else {
				targetName = rel.From.Name
			}

			typedef := RenderRelationshipType(rel, rel.From.Name == c.Name)
			if typedef == "" {
				continue
			}

			if rel.Type == ast.Inheritance {
				// for inheritance, embed the type
				if rel.From.Name == c.Name {
					buf.WriteString(fmt.Sprintf("\t%s\n", typedef))
				}
				continue
			}
			if rel.To.Name == c.Name {
				fieldName := exportName(targetName, ast.Public)
				buf.WriteString(fmt.Sprintf("\t%s %s\n", fieldName, typedef))
			}
		}

	attributes:
		for _, a := range c.Attributes {
			fieldName := exportName(a.Name, a.Visibility)
			goType := RenderType(a.Type)
			buf.WriteString(fmt.Sprintf("\t%s %s\n", fieldName, goType))
		}

		buf.WriteString("}\n\n")

		recv := receiverName(c.Name)
		for _, m := range c.Methods {
			methodName := exportName(m.Name, m.Visibility)
			params := buildParams(m.Parameters)
			ret := RenderType(m.ReturnType)
			buf.WriteString(fmt.Sprintf("func (%s *%s) %s(%s) %s {\n", recv, exportName(c.Name, ast.Public), methodName, params, ret))
			buf.WriteString("\tpanic(\"TODO: implement\")\n")
			buf.WriteString("}\n\n")
		}
	}

	// format
	src, err := format.Source(buf.Bytes())
	if err != nil {
		// return unformatted on error for easier debugging
		return buf.String(), fmt.Errorf("format error: %w", err)
	}
	return string(src), nil
}

func exportName(name string, vis ast.Visibility) string {
	if name == "" {
		return ""
	}
	// for public we capitalize first rune, otherwise keep as-is (lowercase)
	if vis == ast.Public {
		return capitalize(name)
	}
	return lowerFirst(name)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func receiverName(typeName string) string {
	if typeName == "" {
		return "r"
	}
	first := strings.ToLower(typeName[:1])
	// avoid receiver name collision with common keywords
	if first == "t" || first == "r" || first == "i" {
		return first + "v"
	}
	return first
}

func buildParams(params []ast.Attribute) string {
	parts := make([]string, 0, len(params))
	for i, p := range params {
		n := p.Name
		if n == "" {
			n = fmt.Sprintf("p%d", i)
		}
		parts = append(parts, fmt.Sprintf("%s %s", lowerFirst(n), RenderType(p.Type)))
	}
	return strings.Join(parts, ", ")
}
