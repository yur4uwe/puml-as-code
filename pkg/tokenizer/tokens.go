package tokenizer

import (
	"encoding/json"
	"errors"
	"fmt"
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
	DASH        // -
	TILDE       // ~
	HASH        // #
	PIPE        // |
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
	PACKAGE_SEPARATOR // :: by default, configurable
)

type TokenPos struct {
	Line   uint
	Col    uint
	Offset uint
}

func (p TokenPos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type Token struct {
	Type          TokenType
	Literal       string
	Pos           TokenPos
	LeadingTrivia []Token
}

var singleCharTokens = map[rune]TokenType{
	'\n': NEWLINE,
	'(':  LPAREN,
	')':  RPAREN,
	'{':  LBRACE,
	'}':  RBRACE,
	'[':  LBRACKET,
	']':  RBRACKET,
	'<':  LANGLE,
	'>':  RANGLE,
	',':  COMMA,
	';':  SEMICOLON,
	':':  COLON,
	'.':  DOT,
	'=':  EQUALS,
	'+':  PLUS,
	'-':  DASH,
	'~':  TILDE,
	'#':  HASH,
	'|':  PIPE,
	'*':  ASTERISK,
	'\\': BACKSLASH,
	'^':  CARET,
	'$':  DOLLAR,
	'%':  PERCENT,
	'@':  AT,
	'!':  EXCLAMATION,
	'_':  UNDERSCORE,
}

// ResolveUnambiguousToken handles tokens with obvious, unambiguous identification (no lookahead, no mode).
func ResolveUnambiguousToken(l *Lexer) (Token, bool) {
	if l.isEOF() {
		return Token{Type: EOF, Literal: "", Pos: l.getPos()}, true
	}

	if tt, ok := singleCharTokens[l.ch]; ok {
		return l.consumeChar(tt, string(l.ch)), true
	}

	start := l.getPos()
	switch l.ch {
	case '"':
		return Token{Type: STRING, Literal: l.readString(), Pos: start}, true
	case '\'':
		return Token{Type: COMMENT, Literal: l.readLineComment(), Pos: start}, true
	case '/':
		if l.peekChar() == '\'' {
			return Token{Type: COMMENT, Literal: l.readBlockComment(), Pos: start}, true
		}
		return l.consumeChar(SLASH, string(l.ch)), true
	}

	return Token{}, false
}

// ResolveAmbiguousToken handles identifiers, keywords and numbers.
func ResolveAmbiguousToken(l *Lexer) Token {
	if helpers.IsIdentifierRune(l.ch) || l.ch == '\\' {
		start := l.getPos()
		lit := l.readIdentifier()
		return Token{Type: IDENTIFIER, Literal: lit, Pos: start}
	}

	if unicode.IsDigit(l.ch) {
		start := l.getPos()
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
