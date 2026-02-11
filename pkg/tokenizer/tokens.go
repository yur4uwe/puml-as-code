package tokenizer

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	NEWLINE
	IDENTIFIER
	STRING
	NUMBER

	CLASS
	ABSTRACT
	STRUCT
	INTERFACE
	ENUM
	PACKAGE
	ANNOTATION
	NOTE
	STEREOTYPE
	RECORD
	DATACLASS
	EXCEPTION
	PROTOCOL
	HIDE
	SHOW
	REMOVE
	RESTORE
	SKINPARAM
	SET
	TOGETHER
	END_BLOCK

	// Note tokens
	NOTE_DIRECTION
	NOTE_POSITION
	NOTE_LINK

	LBRACE   // {
	RBRACE   // }
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	SEMICOLON
	COLON
	COMMA

	VISIBILITY
	MODIFIER

	RELATIONSHIP

	COMMENT
	SEPARATOR
	ALIAS
	START
	END
)

func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case NEWLINE:
		return "NEWLINE"
	case IDENTIFIER:
		return "IDENTIFIER"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case CLASS:
		return "CLASS"
	case ABSTRACT:
		return "ABSTRACT"
	case STRUCT:
		return "STRUCT"
	case INTERFACE:
		return "INTERFACE"
	case ENUM:
		return "ENUM"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case LBRACKET:
		return "["
	case RBRACKET:
		return "]"
	case SEMICOLON:
		return ";"
	case COLON:
		return ":"
	case COMMA:
		return ","
	case VISIBILITY:
		return "VISIBILITY"
	case RELATIONSHIP:
		return "RELATIONSHIP"
	case COMMENT:
		return "COMMENT"
	case START:
		return "@startuml"
	case END:
		return "@enduml"
	case PACKAGE:
		return "PACKAGE"
	case ANNOTATION:
		return "ANNOTATION"
	case NOTE:
		return "NOTE"
	case RECORD:
		return "RECORD"
	case DATACLASS:
		return "DATACLASS"
	case EXCEPTION:
		return "EXCEPTION"
	case PROTOCOL:
		return "PROTOCOL"
	case HIDE:
		return "HIDE"
	case SHOW:
		return "SHOW"
	case REMOVE:
		return "REMOVE"
	case RESTORE:
		return "RESTORE"
	case SKINPARAM:
		return "SKINPARAM"
	case SET:
		return "SET"
	case TOGETHER:
		return "TOGETHER"
	case END_BLOCK:
		return "END_BLOCK"
	case MODIFIER:
		return "MODIFIER"
	case ALIAS:
		return "ALIAS"
	case STEREOTYPE:
		return "STEREOTYPE"
	case SEPARATOR:
		return "SEPARATOR"
	case NOTE_DIRECTION:
		return "NOTE_DIRECTION"
	case NOTE_POSITION:
		return "NOTE_POSITION"
	default:
		return "UNKNOWN"
	}
}

type TokenPos struct {
	line uint
	col  uint
}

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

// Token Resolution Architecture: tokens by identification vs mode.
// Stage 1: Unambiguous tokens (no lookahead, no mode)
// Stage 2: Mode-dependent tokens (mode changes or mode-specific meanings)
// Stage 3: Ambiguous tokens (lookahead/complex disambiguation)

// ResolveUnambiguousToken handles tokens with obvious, unambiguous identification (no lookahead, no mode).
func ResolveUnambiguousToken(l *Lexer) (Token, bool) {
	var tok Token

	switch l.ch {
	case 0:
		return Token{Type: EOF, Literal: "", Pos: l.position}, true

	case '\n':
		tok = Token{Type: NEWLINE, Literal: "\n", Pos: l.position}
		if l.CurrentMode() == MODE_LABEL {
			l.PopMode()
		}
		// Pop MODE_NOTE if we're in a single-line note (not multi-line with end note)
		if l.CurrentMode() == MODE_NOTE && !l.isMultilineNote {
			l.PopMode()
		}
		l.readChar()
		return tok, true

	case '"':
		start := l.position
		lit := l.readString()
		return Token{Type: STRING, Literal: lit, Pos: start}, true

	case '(':
		tok = Token{Type: LPAREN, Literal: "(", Pos: l.position}
		l.readChar()
		return tok, true

	case ')':
		tok = Token{Type: RPAREN, Literal: ")", Pos: l.position}
		l.readChar()
		return tok, true

	case '[':
		tok = Token{Type: LBRACKET, Literal: "[", Pos: l.position}
		l.readChar()
		return tok, true

	case ']':
		tok = Token{Type: RBRACKET, Literal: "]", Pos: l.position}
		l.readChar()
		return tok, true

	case ',':
		tok = Token{Type: COMMA, Literal: ",", Pos: l.position}
		l.readChar()
		return tok, true

	case ';':
		tok = Token{Type: SEMICOLON, Literal: ";", Pos: l.position}
		l.readChar()
		return tok, true
	case '\'':
		start := l.position
		lit := l.readLineComment()
		return Token{Type: COMMENT, Literal: lit, Pos: start}, true
	}

	// Multi-character unambiguous patterns
	switch {
	case l.ch == '<' && l.peekChar() == '<':
		return l.readStereotype(), true

	case l.ch == '@':
		tok = l.readUMLBounds()
		l.readChar()
		return tok, true
	}

	return Token{}, false
}

// ResolveContextAwareToken handles tokens whose meaning depends on lexer mode.
func ResolveContextAwareToken(l *Lexer) (Token, bool) {
	switch l.CurrentMode() {
	case MODE_DEFAULT:
		return Token{}, false
	case MODE_CLASS:
		switch {
		case isClassSeparator(l.ch) && isClassSeparator(l.peekChar()) && l.ch == l.peekChar():
			sepRune := l.ch
			l.readChar()
			l.readChar()
			l.PushMode(MODE_LABEL)
			return Token{Type: SEPARATOR, Literal: string(sepRune) + string(sepRune), Pos: l.position}, true
		case isVisibilityRune(l.ch):
			l.readChar()
			return Token{Type: VISIBILITY, Literal: string(l.ch), Pos: l.position}, true
		case isIdentifierRune(l.ch) || l.ch == '\\':
			lit := l.readIdentifier()
			return Token{Type: IDENTIFIER, Literal: strings.TrimSpace(lit), Pos: l.position}, true
		}
	case MODE_LABEL:
		switch {
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
	case MODE_NOTE:
		switch {
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
	}

	return Token{}, false
}

// ResolveAmbiguousToken handles lookahead-heavy cases.
func ResolveAmbiguousToken(l *Lexer) Token {
	switch {
	case (l.ch == '\\' || l.ch == '$' || isIdentifierRune(l.ch)) && !isRelationLineChar(l.peekChar()):
		start := l.position
		lit := l.readIdentifier()
		tt := lookupKeyword(lit)
		if tt != IDENTIFIER {
			l.keywordModeSwitcher(tt, lit)
		}
		return Token{Type: tt, Literal: lit, Pos: start}
	case l.ch == ':':
		tok := Token{Type: COLON, Literal: ":", Pos: l.position}
		l.readChar()
		return tok
	case l.ch == '{':
		if isIdentifierRune(l.peekChar()) {
			return l.readModifier()
		}
		tok := Token{Type: LBRACE, Literal: "{", Pos: l.position}
		l.readChar()
		return tok
	case l.ch == '}' && !isRelationLineChar(l.peekChar()):
		// should be very buggy due to {abstract} or alike
		if l.CurrentMode() == MODE_CLASS {
			l.PopMode()
		}
		rbacePos := l.position
		l.readChar()
		return Token{Type: RBRACE, Literal: "}", Pos: rbacePos}
	// Relations: line chars (-, .) start alone; others need line chars to follow
	case isRelationLineChar(l.ch) || (isRelationLineStartChar(l.ch) && isRelationLineChar(l.peekChar())) || (l.ch == '<' || l.peekChar() == '|'):
		return l.readRelation()

	// Visibility requires lookahead to avoid relation parsing
	case isVisibilityRune(l.ch) && l.ch != '-' && !isRelationLineChar(l.peekChar()):
		tok := Token{Type: VISIBILITY, Literal: string(l.ch), Pos: l.position}
		l.readChar()
		return tok

	case unicode.IsDigit(l.ch):
		start := l.position
		lit := l.readNumber()
		return Token{Type: NUMBER, Literal: lit, Pos: start}

	default:
		tok := Token{Type: ILLEGAL, Literal: string(l.ch), Pos: l.position}
		l.readChar()
		return tok
	}
}
