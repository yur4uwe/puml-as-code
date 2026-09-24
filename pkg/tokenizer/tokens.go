package tokenizer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
)

type Pos struct {
	Line   uint
	Col    uint
	Offset uint
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type SourceSpan struct {
	Start, End Pos
}

type Token struct {
	Type    TokenType
	Literal string
	Pos     Pos
}

func (t Token) EndOffset() uint {
	return t.Pos.Offset + uint(len([]rune(t.Literal)))
}

func (t Token) EndPos() Pos {
	runeLen := uint(len([]rune(t.Literal)))
	return Pos{
		Line:   t.Pos.Line,
		Col:    t.Pos.Col + runeLen,
		Offset: t.Pos.Offset + runeLen,
	}
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

	// Special case for Windows CRLF line endings
	if l.ch == '\r' {
		l.readChar()
		if l.ch == '\n' {
			return l.consumeChar(NEWLINE, "\n"), true
		}
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

func SpanBetween(first, last Token) SourceSpan {
	if first.Pos.Offset > last.Pos.Offset {
		first, last = last, first
	}
	return SourceSpan{
		Start: first.Pos,
		End:   last.EndPos(),
	}
}

func TokenSliceSpan(toks []Token) (span SourceSpan, ok bool) {
	if len(toks) == 0 {
		return SourceSpan{}, false
	}
	return SpanBetween(toks[0], toks[len(toks)-1]), true
}

// MatchTokenPrefix checks whether the beginning of lineToks matches target.
// It joins the literals of leading tokens until they match target (ignoring spaces),
// allowing both single-token and multi-token representations (e.g. ["!", "endif"],
// ["end", "note"], ["end", "legend"], ["endlegend"], ["alt"]).
// It returns the number of line tokens matched, and whether a match was found.
func MatchTokenPrefix(lineToks []Token, target string) (int, bool) {
	if len(lineToks) == 0 || target == "" {
		return 0, false
	}

	cleanTarget := strings.ToLower(strings.ReplaceAll(target, " ", ""))
	var accumulated strings.Builder

	for i, tok := range lineToks {
		accumulated.WriteString(strings.ToLower(tok.Literal))
		if accumulated.String() == cleanTarget {
			return i + 1, true
		}
		if accumulated.Len() >= len(cleanTarget) {
			return 0, false
		}
	}
	return 0, false
}

// MatchTokenLine checks whether lineToks matches target exactly with no trailing tokens on the line.
// It returns true only if the entire line of tokens forms the target (ignoring spaces).
func MatchTokenLine(lineToks []Token, target string) bool {
	consumed, ok := MatchTokenPrefix(lineToks, target)
	return ok && consumed == len(lineToks)
}
