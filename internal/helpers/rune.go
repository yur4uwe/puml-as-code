package helpers

import "unicode"

func IsIdentifierRune(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '\\'
}
