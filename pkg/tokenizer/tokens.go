package tokenizer

import (
	"encoding/json"
	"errors"
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
	RECORD
	DATACLASS
	EXCEPTION
	PROTOCOL
	ACTION
	SKINPARAM
	TOGETHER
	SET       // set
	ALIAS     // as
	END_BLOCK // end

	// Note tokens
	NOTE_DIRECTION
	NOTE_POSITION

	LBRACE      // {
	RBRACE      // }
	LPAREN      // (
	RPAREN      // )
	LBRACKET    // [
	RBRACKET    // ]
	LANGLE      // <
	RANGLE      // >
	SEMICOLON   // ;
	COLON       // :
	COMMA       // ,
	DOT         // .
	EQUALS      // =
	PLUS        // +
	HYPHEN      // -
	TILDE       // ~
	HASH        // #
	VBAR        // |
	ASTERISK    // *
	SLASH       // /
	BACKSLASH   // \
	CARET       // ^
	DOLLAR      // $
	PERCENT     // %
	AT          // @
	EXCLAMATION // !
	UNDERSCORE  // _

	COMMENT
	SEPARATOR // ::
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

// ResolveUnambiguousToken handles tokens with obvious, unambiguous identification (no lookahead, no mode).
func ResolveUnambiguousToken(l *Lexer) (Token, bool) {
	switch l.ch {
	case 0:
		return Token{Type: EOF, Literal: "", Pos: l.position}, true
	case '\n':
		return l.consumeChar(NEWLINE, "\n"), true
	case '(':
		return l.consumeChar(LPAREN, "("), true
	case ')':
		return l.consumeChar(RPAREN, ")"), true
	case '{':
		return l.consumeChar(LBRACE, "{"), true
	case '}':
		return l.consumeChar(RBRACE, "}"), true
	case '[':
		return l.consumeChar(LBRACKET, "["), true
	case ']':
		return l.consumeChar(RBRACKET, "]"), true
	case '<':
		return l.consumeChar(LANGLE, "<"), true
	case '>':
		return l.consumeChar(RANGLE, ">"), true
	case ',':
		return l.consumeChar(COMMA, ","), true
	case ';':
		return l.consumeChar(SEMICOLON, ";"), true
	case ':':
		return l.consumeChar(COLON, ":"), true
	case '.':
		return l.consumeChar(DOT, "."), true
	case '=':
		return l.consumeChar(EQUALS, "="), true
	case '+':
		return l.consumeChar(PLUS, "+"), true
	case '-':
		return l.consumeChar(HYPHEN, "-"), true
	case '_':
		return l.consumeChar(UNDERSCORE, "_"), true
	case '~':
		return l.consumeChar(TILDE, "~"), true
	case '#':
		return l.consumeChar(HASH, "#"), true
	case '|':
		return l.consumeChar(VBAR, "|"), true
	case '*':
		return l.consumeChar(ASTERISK, "*"), true
	case '/':
		return l.consumeChar(SLASH, "/"), true
	case '\\':
		return l.consumeChar(BACKSLASH, "\\"), true
	case '^':
		return l.consumeChar(CARET, "^"), true
	case '$':
		return l.consumeChar(DOLLAR, "$"), true
	case '%':
		return l.consumeChar(PERCENT, "%"), true
	case '@':
		return l.consumeChar(AT, "@"), true
	case '!':
		return l.consumeChar(EXCLAMATION, "!"), true
	case '"':
		start := l.position
		return Token{Type: STRING, Literal: l.readString(), Pos: start}, true
	case '\'':
		start := l.position
		lit := l.readLineComment()
		return Token{Type: COMMENT, Literal: lit, Pos: start}, true
	}

	return Token{}, false
}

// ResolveAmbiguousToken handles identifiers, keywords and numbers.
func ResolveAmbiguousToken(l *Lexer) Token {
	if helpers.IsIdentifierRune(l.ch) || l.ch == '\\' {
		start := l.position
		lit := l.readIdentifier()
		tt := lookupKeyword(lit)
		return Token{Type: tt, Literal: lit, Pos: start}
	}

	if unicode.IsDigit(l.ch) {
		start := l.position
		lit, err := l.readNumber()
		if err != nil {
			// If the error is about a trailing identifier character, it means this is
			// an identifier that just happens to start with digits (like a hex color 00FFFF).
			// We continue reading it as an identifier and combine the literals.
			if errors.Is(err, ErrInvalidTrailingChar) {
				rest := l.readIdentifier()
				fullLit := lit + rest
				return Token{Type: IDENTIFIER, Literal: fullLit, Pos: start}
			}
			return Token{Type: ILLEGAL, Literal: err.Error(), Pos: start}
		}
		return Token{Type: NUMBER, Literal: lit, Pos: start}
	}

	return l.consumeChar(ILLEGAL, string(l.ch))
}
