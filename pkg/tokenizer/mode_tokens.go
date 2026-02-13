package tokenizer

import (
	"fmt"
	"strings"
	"unicode"
	"yur4uwe/pac/internal/helpers"
)

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
		l.PopMode()
		l.PushMode(MODE_CLASS)
		return l.consumeChar(LBRACE, "{"), true
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
	case l.isPackageSeparator():
		return Token{Type: SEPARATOR, Literal: string(l.packageSeparator), Pos: l.position - len(l.packageSeparator)}, true
	case helpers.IsIdentifierRune(l.ch) || l.ch == '\\' || l.ch == '$':
		start := l.position
		lit := l.readIdentifier()
		tt := lookupKeyword(lit)
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
		sepRune := l.ch
		l.readChar()
		l.readChar()
		l.PushMode(MODE_LABEL)
		return Token{Type: SEPARATOR, Literal: string(sepRune) + string(sepRune), Pos: start}, true
	case helpers.IsVisibilityRune(l.ch):
		return l.consumeChar(VISIBILITY, string(l.ch)), true
	case helpers.IsIdentifierRune(l.ch) || l.ch == '\\':
		start := l.position
		lit := l.readIdentifier()
		tt := lookupKeyword(lit)
		if lit == "implements" {
			return Token{Type: RELATIONSHIP, Literal: lit, Pos: start}, true
		}
		if lit == "extends" {
			return Token{Type: RELATIONSHIP, Literal: lit, Pos: start}, true
		}
		if tt == NOTE {
			l.PushMode(MODE_NOTE)
			l.isMultilineNote = false
		}
		if tt == IDENTIFIER {
			return Token{Type: IDENTIFIER, Literal: strings.TrimSpace(lit), Pos: start}, true
		}
		return Token{Type: tt, Literal: strings.TrimSpace(lit), Pos: start}, true
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
	case l.ch == ':':
		l.PopMode()
		l.PushMode(MODE_LABEL)
		l.isMultilineNote = false
		l.noteDeclarationComplete = false
		return l.consumeChar(COLON, ":"), true
	case helpers.IsIdentifierRune(l.ch) || unicode.IsNumber(l.ch) || l.ch == '<':
		start := l.position
		lit := l.readIdentifier()
		ntt := lookupNoteKeyword(lit)
		if ntt == END_BLOCK {
			if l.ch == ' ' {
				l.readChar()
			}
			noteStr := l.readIdentifier()
			l.PopMode()
			l.isMultilineNote = false
			l.noteDeclarationComplete = false
			return Token{Type: END_BLOCK, Literal: fmt.Sprintf("%s %s", lit, noteStr), Pos: start}, true
		}
		if ntt == NOTE_POSITION {
			if lit == "on" {
				l.findNextTokenStart()
				if l.ch != 0 && unicode.IsLetter(l.ch) {
					linkStart := l.position
					nextWord := l.readIdentifier()
					if nextWord == "link" {
						l.isMultilineNote = true
						l.noteDeclarationComplete = false
						return Token{Type: NOTE_POSITION, Literal: "on link", Pos: start}, true
					}
					l.jumpToPosition(linkStart)
				}
			}
			l.isMultilineNote = true
			l.noteDeclarationComplete = false
			return Token{Type: ntt, Literal: strings.TrimSpace(lit), Pos: start}, true
		}
		if ntt != IDENTIFIER {
			return Token{Type: ntt, Literal: strings.TrimSpace(lit), Pos: start}, true
		}
		// IDENTIFIER in multiline note: check if we should consume all content
		if l.isMultilineNote && l.noteDeclarationComplete {
			// Consume all note content until "end note"
			contentStart := start
			var contentLines []string

			// Read rest of current line
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			firstLine := strings.TrimSpace(string(l.input[start:l.position]))
			if firstLine != "" {
				contentLines = append(contentLines, firstLine)
			}

			// Consume all following lines until "end note"
			for {
				if l.ch == '\n' {
					l.readChar()
				}
				if l.ch == 0 {
					break
				}

				l.findNextTokenStart()

				// Check for "end note"
				if unicode.IsLetter(l.ch) {
					checkPos := l.position
					checkWord := l.readIdentifier()
					if checkWord == "end" {
						if l.ch == ' ' {
							l.readChar()
							nextWord := l.readIdentifier()
							if nextWord == "note" {
								l.jumpToPosition(checkPos)
								break
							}
						}
					}
					l.jumpToPosition(checkPos)
				}

				// Read this line as content
				lineStart := l.position
				for l.ch != '\n' && l.ch != 0 {
					l.readChar()
				}
				line := strings.TrimSpace(string(l.input[lineStart:l.position]))
				if line != "" {
					contentLines = append(contentLines, line)
				}
			}

			content := strings.Join(contentLines, "\n")
			return Token{Type: IDENTIFIER, Literal: content, Pos: contentStart}, true
		}
		// Just return the identifier (target element name on declaration line)
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		lit = strings.TrimSpace(string(l.input[start:l.position]))
		return Token{Type: IDENTIFIER, Literal: lit, Pos: start}, true
	default:
		// Non-letter character (like <, /)
		if l.isMultilineNote && l.noteDeclarationComplete {
			// Consume all content until "end note"
			start := l.position
			var contentLines []string

			for {
				lineStart := l.position
				for l.ch != '\n' && l.ch != 0 {
					l.readChar()
				}
				line := strings.TrimSpace(string(l.input[lineStart:l.position]))
				if line != "" {
					contentLines = append(contentLines, line)
				}

				if l.ch == '\n' {
					l.readChar()
				}
				if l.ch == 0 {
					break
				}

				l.findNextTokenStart()

				// Check for "end note"
				if unicode.IsLetter(l.ch) {
					checkPos := l.position
					checkWord := l.readIdentifier()
					if checkWord == "end" {
						if l.ch == ' ' {
							l.readChar()
							nextWord := l.readIdentifier()
							if nextWord == "note" {
								l.jumpToPosition(checkPos)
								break
							}
						}
					}
					l.jumpToPosition(checkPos)
				}
			}

			content := strings.Join(contentLines, "\n")
			return Token{Type: IDENTIFIER, Literal: content, Pos: start}, true
		}
		// Single-line or still on declaration line
		start := l.position
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		lit := strings.TrimSpace(string(l.input[start:l.position]))
		return Token{Type: IDENTIFIER, Literal: lit, Pos: start}, true
	}
}
