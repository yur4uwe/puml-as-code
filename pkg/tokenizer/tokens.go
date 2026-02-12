package tokenizer

import (
	"unicode"
	"yur4uwe/pac/internal/helpers"
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
	SET_PROPERTY
	TOGETHER
	END_BLOCK
	GENERIC

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
	case SET_PROPERTY:
		return "SET_PROPERTY"
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
	case GENERIC:
		return "GENERIC"
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
		lit := l.readGeneric()
		return Token{Type: GENERIC, Literal: lit, Pos: l.position}, true
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
		return resolveDefaultModeToken(l)
	case MODE_CLASS_DEF:
		return resolveClassDefModeToken(l)
	case MODE_CLASS:
		return resolveClassModeToken(l)
	case MODE_LABEL:
		return resolveLabelModeToken(l)
	case MODE_NOTE:
		return resolveNoteModeToken(l)
	}
	return Token{}, false
}

// ResolveAmbiguousToken handles lookahead-heavy cases.
func ResolveAmbiguousToken(l *Lexer) Token {
	switch {
	case (l.ch == '\\' || l.ch == '$' || helpers.IsIdentifierRune(l.ch)) && !helpers.IsRelationLineChar(l.peekChar()):
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

	default:
		return l.consumeChar(ILLEGAL, string(l.ch))
	}
}
