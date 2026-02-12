package tokenizer

import (
	"fmt"
	"strings"
	"unicode"
)

func resolveDefaultModeToken(l *Lexer) (Token, bool) {
	switch l.ch {
	case '\n':
		tok := Token{Type: NEWLINE, Literal: "\n", Pos: l.position}
		l.readChar()
		return tok, true
	case '(':
		if l.peekChar() != ')' {
			return Token{Type: LPAREN, Literal: "(", Pos: l.position}, true
		}
		start := l.position
		l.readChar() // consume '('
		l.readChar() // consume ')'
		tok := l.readRelation()
		if tok.Type == ILLEGAL {
			l.jumpToPosition(start)
			return Token{Type: LPAREN, Literal: "(", Pos: l.position}, true
		}
		tok.Literal = "()" + tok.Literal
		tok.Pos = start
		return tok, true
	}
	return Token{}, false
}

func resolveClassDefModeToken(l *Lexer) (Token, bool) {
	switch {
	case l.ch == '\n':
		tok := Token{Type: NEWLINE, Literal: "\n", Pos: l.position}
		l.PopMode()
		l.readChar()
		return tok, true
	case l.ch == '{':
		l.PopMode()
		l.PushMode(MODE_CLASS)
		tok := Token{Type: LBRACE, Literal: "{", Pos: l.position}
		l.readChar()
		return tok, true
	case string(l.peekAhead(len("extends"))) == "extends":
		start := l.position
		for _ = range "extends" {
			l.readChar()
		}
		return Token{Type: RELATIONSHIP, Literal: "extends", Pos: start}, true
	case string(l.peekAhead(len("implements"))) == "implements":
		start := l.position
		for _ = range "implements" {
			l.readChar()
		}
		return Token{Type: RELATIONSHIP, Literal: "implements", Pos: start}, true
	case l.isPackageSeparator():
		return Token{Type: SEPARATOR, Literal: string(l.packageSeparator), Pos: l.position - len(l.packageSeparator)}, true
	}
	return Token{}, false
}

func resolveClassModeToken(l *Lexer) (Token, bool) {
	switch {
	case l.ch == '\n':
		tok := Token{Type: NEWLINE, Literal: "\n", Pos: l.position}
		l.readChar()
		return tok, true
	case l.ch == '(':
		tok := Token{Type: LPAREN, Literal: "(", Pos: l.position}
		l.readChar()
		return tok, true
	case l.ch == ')':
		tok := Token{Type: RPAREN, Literal: ")", Pos: l.position}
		l.readChar()
		return tok, true
	case l.ch == '}':
		tok := Token{Type: RBRACE, Literal: "}", Pos: l.position}
		if l.position > 0 && !isIdentifierRune(l.input[l.position-1]) {
			l.PopMode()
		}
		l.readChar()
		return tok, true
	case isClassSeparator(l.ch) && isClassSeparator(l.peekChar()) && l.ch == l.peekChar():
		sepRune := l.ch
		l.readChar()
		l.readChar()
		l.PushMode(MODE_LABEL)
		return Token{Type: SEPARATOR, Literal: string(sepRune) + string(sepRune), Pos: l.position}, true
	case isVisibilityRune(l.ch):
		visChar := l.ch
		visPos := l.position
		l.readChar()
		return Token{Type: VISIBILITY, Literal: string(visChar), Pos: visPos}, true
	case isIdentifierRune(l.ch) || l.ch == '\\':
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
		tok := Token{Type: NEWLINE, Literal: "\n", Pos: l.position}
		l.PopMode()
		l.readChar()
		return tok, true
	case isClassSeparator(l.ch) && isClassSeparator(l.peekChar()) && l.ch == l.peekChar():
		sepRune := l.ch
		l.readChar()
		l.readChar()
		l.PopMode()
		return Token{Type: SEPARATOR, Literal: string(sepRune) + string(sepRune), Pos: l.position}, true
	case isIdentifierRune(l.ch) || unicode.IsDigit(l.ch):
		lit := l.readLabel()
		return Token{Type: IDENTIFIER, Literal: strings.TrimSpace(lit), Pos: l.position}, true
	}
	return Token{}, false
}

func resolveNoteModeToken(l *Lexer) (Token, bool) {
	switch {
	case l.ch == '\n':
		tok := Token{Type: NEWLINE, Literal: "\n", Pos: l.position}
		if !l.isMultilineNote {
			l.PopMode()
		}
		l.readChar()
		return tok, true
	case l.ch == ':':
		tok := Token{Type: COLON, Literal: ":", Pos: l.position}
		l.readChar()
		l.PopMode()
		l.PushMode(MODE_LABEL)
		l.isMultilineNote = false // Colon indicates single-line note
		return tok, true
	case unicode.IsLetter(l.ch):
		start := l.position
		lit := l.readIdentifier()
		ntt := lookupNoteKeyword(lit)
		if ntt == END_BLOCK {
			// consume space if present
			if l.ch == ' ' {
				l.readChar()
			}
			// consume 'note' token
			noteStr := l.readIdentifier()
			l.PopMode()
			l.isMultilineNote = false // Note ended
			return Token{Type: END_BLOCK, Literal: fmt.Sprintf("%s %s", lit, noteStr), Pos: start}, true
		}
		if ntt == NOTE_POSITION {
			// of/on indicates a multi-line block note
			l.isMultilineNote = true
		}
		// If it's not a recognized keyword, read the rest of the line as content
		if ntt == IDENTIFIER {
			// Continue reading line content (spaces, digits, etc.) until newline
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			lit = strings.TrimSpace(string(l.input[start:l.position]))
			return Token{Type: IDENTIFIER, Literal: lit, Pos: start}, true
		}
		return Token{Type: ntt, Literal: strings.TrimSpace(lit), Pos: start}, true
	}
	return Token{}, false
}
