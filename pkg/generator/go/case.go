package gogenerator

import (
	"log"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

func ensureUpperFirst(s string) string {
	firstByte := string(s[0])
	firstByte = strings.ToUpper(firstByte)
	return firstByte + s[1:]
}

func ensureLowerFirst(s string) string {
	firstByte := string(s[0])
	firstByte = strings.ToLower(firstByte)
	return firstByte + s[1:]
}

func isLower(s byte) bool {
	return s >= 'a' && s <= 'z'
}

func isUpper(s byte) bool {
	return s >= 'A' && s <= 'Z'
}

func ensureCorrectCase(ownerName string, fieldName string, visibility ast.VisibilityKind) string {
	newName := fieldName

	switch visibility {
	case ast.VisibilityProtected, ast.VisibilityPrivate:
		log.Printf("warning: %s encapsulation for field %s.%s is not supported in Go, using package encapsulation instead", visibility, ownerName, fieldName)
		fallthrough
	case ast.VisibilityPackage:
		if !isUpper(newName[0]) {
			return fieldName
		}
		newName = ensureLowerFirst(newName)
		log.Printf("warning: changing field name capitalization %s.%s to %s", ownerName, fieldName, newName)
	case ast.VisibilityPublic:
		if !isLower(newName[0]) {
			return fieldName
		}
		newName = ensureUpperFirst(newName)
		log.Printf("warning: changing field name capitalization %s.%s to %s", ownerName, fieldName, newName)
	}

	return newName
}
