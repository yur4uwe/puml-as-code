package tokenizer

import (
	"encoding/json"
	"unicode"
	"yur4uwe/pac/internal/helpers"
)

//go:generate enumer -type=TokenType -transform=upper -json
type TokenType byte

var _ json.Unmarshaler = (*TokenType)(nil)

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
	ACTION
	SKINPARAM
	SET_PROPERTY
	TOGETHER
	END_BLOCK
	GENERIC
	TAG

	// Note tokens
	NOTE_DIRECTION
	NOTE_POSITION

	LBRACE   // {
	RBRACE   // }
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	SEMICOLON
	COLON
	COMMA
	HASH

	VISIBILITY
	MODIFIER

	RELATIONSHIP

	COMMENT
	SEPARATOR
	ALIAS
	START
	END
	PREPROCESSOR
	TARGET
	ASPECT
)

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
	// In NOTE mode, skip generic/stereotype detection to let note handler process all content
	if l.CurrentMode() == MODE_NOTE && (l.ch == '<' || l.ch == '/') {
		return Token{}, false
	}

	switch l.ch {
	case 0:
		return Token{Type: EOF, Literal: "", Pos: l.position}, true

	case '"':
		start := l.position
		lit := l.readString()
		return Token{Type: STRING, Literal: lit, Pos: start}, true
	case ')':
		return l.consumeChar(RPAREN, ")"), true
	case '[':
		return l.consumeChar(LBRACKET, "["), true
	case ']':
		return l.consumeChar(RBRACKET, "]"), true
	case ',':
		return l.consumeChar(COMMA, ","), true
	case ';':
		return l.consumeChar(SEMICOLON, ";"), true
	case '\'':
		start := l.position
		lit := l.readLineComment()
		return Token{Type: COMMENT, Literal: lit, Pos: start}, true
	}

	// Multi-character unambiguous patterns
	switch {
	case l.ch == '<' && (l.peekChar() == '?' || unicode.IsLetter(l.peekChar())):
		start := l.position
		lit := l.readGeneric()
		return Token{Type: GENERIC, Literal: lit, Pos: start}, true
	case l.ch == '<' && l.peekChar() == '<':
		start := l.position
		tok := l.readStereotype()
		tok.Pos = start
		return tok, true
	case l.ch == '@':
		return l.readUMLBounds(), true
	case l.ch == '!':
		start := l.position
		l.readChar() // consume '!'
		directive := l.readIdentifier()
		return Token{Type: PREPROCESSOR, Literal: "!" + directive, Pos: start}, true
	}

	return Token{}, false
}

// ResolveContextAwareToken handles tokens whose meaning depends on lexer mode.
func ResolveContextAwareToken(l *Lexer) (Token, bool) {
	switch l.CurrentMode() {
	case MODE_DEFAULT:
		return resolveDefaultModeToken(l)
	case MODE_CLASS_DEF:
		return resolveClassDefModeToken(l)
	case MODE_CLASS:
		return resolveClassModeToken(l)
	case MODE_LABEL:
		return resolveLabelModeToken(l)
	case MODE_NOTE:
		return resolveNoteModeToken(l)
	case MODE_STYLE:
		return resolveStyleModeToken(l)
	case MODE_ACTION:
		return resolveActionModeToken(l)
	}
	return Token{}, false
}

// ResolveAmbiguousToken handles lookahead-heavy cases.
func ResolveAmbiguousToken(l *Lexer) Token {
	switch {
	case l.ch == '$':
		start := l.position
		return Token{Type: TAG, Literal: l.readIdentifier(), Pos: start}
	case (l.ch == '\\' || helpers.IsIdentifierRune(l.ch)) && !helpers.IsRelationLineChar(l.peekChar()):
		start := l.position
		lit := l.readIdentifier()
		tt := lookupKeyword(lit)
		if tt != IDENTIFIER {
			l.keywordModeSwitcher(tt)
		}
		if tt == SET_PROPERTY {
			return Token{Type: SET_PROPERTY, Literal: l.readProperty(), Pos: start}
		}
		return Token{Type: tt, Literal: lit, Pos: start}
	case l.ch == ':':
		return l.consumeChar(COLON, ":")
	case l.ch == '{':
		if helpers.IsIdentifierRune(l.peekChar()) {
			return l.readModifier()
		}
		return l.consumeChar(LBRACE, "{")
	case l.ch == '}' && !helpers.IsRelationLineChar(l.peekChar()):
		return l.consumeChar(RBRACE, "}")
	// Relations: line chars (-, .) start alone; others need line chars to follow
	case helpers.IsRelationLineChar(l.ch) || (helpers.IsRelationLineStartChar(l.ch) && helpers.IsRelationLineChar(l.peekChar())) || (l.ch == '<' || l.peekChar() == '|'):
		return l.readRelation()
	// Visibility requires lookahead to avoid relation parsing
	case helpers.IsVisibilityRune(l.ch) && l.ch != '-' && !helpers.IsRelationLineChar(l.peekChar()):
		return l.consumeChar(VISIBILITY, string(l.ch))
	case unicode.IsDigit(l.ch):
		start := l.position
		lit := l.readNumber()
		return Token{Type: NUMBER, Literal: lit, Pos: start}
	case l.ch == '*':
		// Wildcard * when not part of a relationship (e.g., in "remove *")
		if !helpers.IsRelationLineChar(l.peekChar()) {
			return l.consumeChar(IDENTIFIER, "*")
		}
		// Otherwise, try to read as relationship
		return l.readRelation()
	default:
		return l.consumeChar(ILLEGAL, string(l.ch))
	}
}
