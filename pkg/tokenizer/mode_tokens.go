package tokenizer

import (
	"strings"
	"unicode"
	"yur4uwe/pac/internal/helpers"
)

func resolveStyleModeToken(l *Lexer) Token {
	if l.ch == '#' {
		return l.consumeChar(HASH, "#")
	}

	l.PopMode()

	break_chars := make([]rune, 0, 5)
	break_chars = append(break_chars, ' ', '\n', '\r', '\t')

	if l.CurrentMode() == MODE_NOTE {
		break_chars = append(break_chars, ':')
	} else if l.CurrentMode() == MODE_CLASS_DEF {
		break_chars = append(break_chars, '{')
	}

	start := l.position
	lit := l.readUntil(break_chars...)
	return Token{Type: IDENTIFIER, Literal: lit, Pos: start}
}

func resolveDefaultModeToken(l *Lexer) (Token, bool) {
	switch l.ch {
	case '\n':
		return l.consumeChar(NEWLINE, "\n"), true
	case '(':
		if l.peekChar() != ')' {
			return l.consumeChar(LPAREN, "("), true
		}
		start := l.position
		l.readChar() // consume '('
		l.readChar() // consume ')'
		tok := l.readRelation()
		if tok.Type == ILLEGAL {
			l.jumpToPosition(start + 1)
			return Token{Type: LPAREN, Literal: "(", Pos: l.position}, true
		}
		tok.Literal = "()" + tok.Literal
		tok.Pos = start
		return tok, true
	case ':':
		l.PushMode(MODE_LABEL)
		return l.consumeChar(COLON, ":"), true
	}
	return Token{}, false
}

func resolveClassDefModeToken(l *Lexer) (Token, bool) {
	switch {
	case l.ch == '\n':
		l.PopMode()
		return l.consumeChar(NEWLINE, "\n"), true
	case l.ch == '{':
		if l.PopMode() != MODE_PACKAGE_DEF {
			l.PushMode(MODE_CLASS)
		}
		return l.consumeChar(LBRACE, "{"), true
	case l.ch == '#':
		l.PushMode(MODE_STYLE)
		return l.consumeChar(HASH, "#"), true
	case l.isPackageSeparator():
		start := l.position
		for _ = range l.packageSeparator {
			l.readChar()
		}
		return Token{Type: SEPARATOR, Literal: string(l.packageSeparator), Pos: start}, true
	case string(l.lookAhead(len("extends"))) == "extends":
		start := l.position
		for _ = range "extends" {
			l.readChar()
		}
		return Token{Type: RELATIONSHIP, Literal: "extends", Pos: start}, true
	case string(l.lookAhead(len("implements"))) == "implements":
		start := l.position
		for _ = range "implements" {
			l.readChar()
		}
		return Token{Type: RELATIONSHIP, Literal: "implements", Pos: start}, true
	case l.ch == '$' && l.isClassNameSet:
		start := l.position
		lit := l.readIdentifier()
		return Token{Type: TAG, Literal: lit, Pos: start}, true
	case helpers.IsIdentifierRune(l.ch) || l.ch == '\\' || l.ch == '$':
		start := l.position
		lit := l.readIdentifier()
		if l.peekChar() != ' ' && !l.isPackageSeparator() && l.ch != '<' {
			// we have a 'none' package separator read until ' ' or '\n'
			// TODO: Should actually stop if finds  a package separator
			lit += l.readUntil(' ', '\n', '\r')
		}
		tt := lookupKeyword(lit)
		l.isClassNameSet = tt == IDENTIFIER
		// In CLASS_DEF mode, identifiers are class names or keywords
		return Token{Type: tt, Literal: lit, Pos: start}, true
	}
	return Token{}, false
}

func resolveClassModeToken(l *Lexer) (Token, bool) {
	switch {
	case l.ch == '\n':
		return l.consumeChar(NEWLINE, "\n"), true
	case l.ch == '(':
		return l.consumeChar(LPAREN, "("), true
	case l.ch == ')':
		return l.consumeChar(RPAREN, ")"), true
	case l.ch == '}':
		if l.position > 0 && !helpers.IsIdentifierRune(l.input[l.position-1]) {
			l.PopMode()
		}
		return l.consumeChar(RBRACE, "}"), true
	case helpers.IsClassSeparator(l.ch) && helpers.IsClassSeparator(l.peekChar()) && l.ch == l.peekChar():
		start := l.position
		sepRune := string(l.readChar()) + string(l.readChar())
		l.PushMode(MODE_LABEL)
		return Token{Type: SEPARATOR, Literal: sepRune, Pos: start}, true
	case helpers.IsVisibilityRune(l.ch):
		return l.consumeChar(VISIBILITY, string(l.ch)), true
	case helpers.IsIdentifierRune(l.ch) || l.ch == '\\':
		start := l.position
		lit := l.readUntil('\n', '\r', ' ', ':', '(', ')')
		lit = strings.ReplaceAll(lit, "\\", "")
		return Token{Type: IDENTIFIER, Literal: strings.TrimSpace(lit), Pos: start}, true
	}
	return Token{}, false
}

func resolveLabelModeToken(l *Lexer) (Token, bool) {
	switch {
	case l.ch == '\n':
		l.PopMode()
		return l.consumeChar(NEWLINE, "\n"), true
	case l.ch == '(':
		return l.consumeChar(LPAREN, "("), true
	case l.ch == ')':
		return l.consumeChar(RPAREN, ")"), true
	case helpers.IsClassSeparator(l.ch) && helpers.IsClassSeparator(l.peekChar()) && l.ch == l.peekChar():
		sepRune := l.ch
		l.readChar()
		l.readChar()
		l.PopMode()
		return Token{Type: SEPARATOR, Literal: string(sepRune) + string(sepRune), Pos: l.position}, true
	case helpers.IsIdentifierRune(l.ch) || unicode.IsDigit(l.ch):
		lit := l.readLabel()
		return Token{Type: IDENTIFIER, Literal: strings.TrimSpace(lit), Pos: l.position}, true
	}
	return Token{}, false
}

func resolveNoteModeToken(l *Lexer) (Token, bool) {
	switch {
	case l.ch == '\n':
		if !l.isMultilineNote {
			l.PopMode()
		} else {
			// Mark that we've passed the declaration line; content starts after this
			l.noteDeclarationComplete = true
		}
		return l.consumeChar(NEWLINE, "\n"), true
	case l.ch == '"':
		start := l.position
		lit := l.readString()
		if string(l.input[start-2:start]) != "::" {
			l.isMultilineNote = false
			l.noteDeclarationComplete = true
			l.PopMode()
		}
		return Token{Type: STRING, Literal: lit, Pos: start}, true
	case l.ch == ':':
		l.isMultilineNote = false
		l.noteDeclarationComplete = true
		return l.consumeChar(COLON, ":"), true
	case l.ch == '#':
		l.PushMode(MODE_STYLE)
		return l.consumeChar(HASH, "#"), true
	case l.noteDeclarationComplete && l.isMultilineNote:
		start := l.position
		sb := strings.Builder{}
		for l.ch != 0 {
			if l.ch == '\n' {
				sb.WriteRune(l.readChar())
				// check if 'end note' is after a newline to close multiline note
				str := string(l.lookAhead(8))
				if str == "end note" {
					l.isMultilineNote = false
					l.noteDeclarationComplete = false
					return Token{Type: STRING, Literal: sb.String(), Pos: start}, true
				}
			}
			sb.WriteRune(l.readChar())
		}
		panic("malformed multiline note")
	case l.noteDeclarationComplete && !l.isMultilineNote:
		start := l.position
		lit := l.readUntil('\r', '\n')
		l.PopMode()
		return Token{Type: STRING, Literal: lit, Pos: start}, true
	case helpers.IsIdentifierRune(l.ch):
		start := l.position
		lit := l.readIdentifier()
		tt := lookupNoteKeyword(lit)
		if tt == END_BLOCK {
			// consume ' ' between 'end' and 'note'
			l.readChar()
			// read 'note' after 'end' at the end of the note on beginning of the line
			lit += " " + l.readIdentifier()
			l.PopMode()
		}
		return Token{Type: tt, Literal: lit, Pos: start}, true
	}
	return Token{}, false
}

func resolveActionModeToken(l *Lexer) (Token, bool) {
	switch l.ch {
	case '\n':
		l.PopMode()
		return l.consumeChar(NEWLINE, "\n"), true
	case '*':
		l.isTargetDetermined = true
		return l.consumeChar(TARGET, "*"), true
	case '@', '$':
		start := l.position
		lit := l.readIdentifier()
		l.isTargetDetermined = true
		return Token{Type: TARGET, Literal: lit, Pos: start}, true
	default:
		start := l.position
		lit := l.readIdentifier()

		currentActionToken := TARGET
		if l.isActionAspect(lit) {
			currentActionToken = ASPECT
		} else {
			l.isTargetDetermined = true
		}
		return Token{Type: currentActionToken, Literal: lit, Pos: start}, true
	}
}

func resolveQualifierToken(l *Lexer) (Token, bool) {
	switch {
	case helpers.IsIdentifierRune(l.ch):
		start := l.position
		lit := l.readIdentifier()
		return Token{Type: IDENTIFIER, Literal: lit, Pos: start}, true
	case l.ch == ':':
		return l.consumeChar(COLON, ":"), true
	default:
		return Token{}, false
	}
}
