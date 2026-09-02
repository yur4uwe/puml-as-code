package gogenerator

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"yur4uwe/pac/pkg/generator/go/stdlib"
	"yur4uwe/pac/pkg/resolver"
)

func targetFieldName(target *resolver.EntitySymbol) string {
	if target == nil {
		return "unknown"
	}
	if target.AST != nil && target.AST.Identifier != "" {
		return target.AST.Identifier
	}
	return stdlib.SimpleName(target.FQN)
}

func targetTypeName(sourcePkg []string, target *resolver.EntitySymbol) string {
	if target == nil {
		return "unknown"
	}
	ident := targetFieldName(target)
	if slices.Equal(sourcePkg, target.PackagePath) {
		return ident
	}
	return extractPackageName(target.PackagePath) + "." + ident
}

func extractPackageName(pkgPath []string) string {
	if len(pkgPath) == 0 {
		return "root"
	}
	return pkgPath[len(pkgPath)-1]
}

func reparseLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" || strings.Contains(label, " ") {
		return ""
	}
	var isPublic *bool
	nameStartIdx := 0
	switch label[0] {
	case '+':
		isPublic = new(bool)
		*isPublic = true
		nameStartIdx++
	case '-', '~', '#':
		isPublic = new(bool)
		*isPublic = false
		nameStartIdx++
	default:
		// Labels without explicit visibility are treated as descriptive verbs/titles
		return ""
	}
	for i := nameStartIdx; i < len(label); i++ {
		if unicode.IsLetter(rune(label[i])) ||
			unicode.IsNumber(rune(label[i])) ||
			label[i] == '_' {
			continue
		}
		return ""
	}
	if len(label[nameStartIdx:]) == 0 {
		return ""
	}
	if *isPublic {
		return ensureUpperFirst(label[nameStartIdx:])
	}
	return ensureLowerFirst(label[nameStartIdx:])
}

func isValidGoIdent(ident string) bool {
	if ident == "" {
		return false
	}
	for i, r := range ident {
		if i == 0 && unicode.IsDigit(r) {
			return false
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func parseGeneric(genericStr string) ([]GenericView, error) {
	var generics []GenericView

	rawParams := strings.SplitSeq(genericStr, ",")
	for raw := range rawParams {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		parts := strings.SplitN(trimmed, " ", 2)
		name := strings.TrimSpace(parts[0])
		if !isValidGoIdent(name) {
			return nil, fmt.Errorf("invalid generic identifier %q in: %s", name, genericStr)
		}

		constraint := "any"
		if len(parts) > 1 {
			constraint = strings.TrimSpace(parts[1])
			if constraint == "" {
				constraint = "any"
			}
		}

		generics = append(generics, GenericView{
			Name:       name,
			Constraint: constraint,
		})
	}

	if len(generics) == 0 {
		return nil, fmt.Errorf("invalid generic declaration: %s", genericStr)
	}

	return generics, nil
}
