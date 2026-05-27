package helpers

import "unicode"

func IsIdentifierRune(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '\\'
}

func IsRelationLineChar(ch rune) bool {
	switch ch {
	case '-', '.':
		return true
	default:
		return false
	}
}

func IsRelationLineStartChar(ch rune) bool {
	switch ch {
	case '-', '.', '<', 'o', '*', '#', '+', '}', 'x', '^':
		return true
	default:
		return false
	}
}

func IsRelationChar(ch rune) bool {
	switch ch {
	case '-', '.', '<', '>', '|', 'o', '*', '#', '+', '}', '{', 'x', '^':
		return true
	default:
		return false
	}
}

func IsRelDirection(lit string) bool {
	switch lit {
	case "left", "right", "up", "down", "l", "r", "u", "d", "le", "ri", "do":
		return true
	default:
		return false
	}
}

func IsInlineRelationLetter(input []rune, pos int) bool {
	if pos < 0 || pos >= len(input) {
		return false
	}
	if input[pos] != 'o' && input[pos] != 'x' {
		return false
	}
	prevRel := pos > 0 && IsRelationChar(input[pos-1])
	nextRel := pos+1 < len(input) && IsRelationChar(input[pos+1])
	return prevRel || nextRel
}

func IsVisibilityRune(ch rune) bool {
	return ch == '+' || ch == '-' || ch == '#' || ch == '~'
}

func IsClassSeparator(ch rune) bool {
	return ch == '-' || ch == '=' || ch == '.' || ch == '_'
}
