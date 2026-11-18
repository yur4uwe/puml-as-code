package tokenizer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"yur4uwe/pac/internal/helpers"
)

var (
	ErrInvalidDot          = errors.New("invalid dot in number")
	ErrInvalidTrailingChar = errors.New("invalid trailing character")
	ErrInvalidScientific   = errors.New("invalid scientific notation")
	ErrInvalidEncDigit     = errors.New("invalid digit for base")
	ErrInvalidBase         = errors.New("invalid number base")
	ErrMissingDigits       = errors.New("encoded number missing digits")
)

type Lexer struct {
	input       string
	position    int
	nextBytePos int
	ch          rune
	line        uint
	col         uint
}

type lexerState struct{}

func (l *Lexer) Pos() Pos {
	return Pos{
		Line:   l.line,
		Col:    l.col,
		Offset: uint(l.position),
	}
}

func (l *Lexer) String() string {
	return fmt.Sprintf("\nposition: %s\nreadPos: %d\nch: %q", l.Pos().String(), l.nextBytePos, l.ch)
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
	}
	l.advancePtr()
	return l
}

func (l *Lexer) advancePtr() {
	if l.nextBytePos >= len(l.input) {
		l.ch = 0
		l.position = l.nextBytePos
		return
	}

	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else if l.ch != 0 {
		l.col++
	}

	l.position = l.nextBytePos
	b := l.input[l.nextBytePos]

	var width int
	if b < utf8.RuneSelf {
		l.ch = rune(b)
		width = 1
	} else {
		l.ch, width = utf8.DecodeRuneInString(l.input[l.nextBytePos:])
	}

	l.nextBytePos += width
}

func (l *Lexer) isEOF() bool {
	return l.ch == 0
}

func (l *Lexer) consumeChar(tokenType TokenType) Token {
	start := l.Pos()
	l.advancePtr()
	end := l.Pos()

	lit := l.input[start.Offset:end.Offset]

	tok := Token{
		Type:    tokenType,
		Literal: lit,
		Span:    SourceSpan{Start: start, End: end},
	}
	return tok
}

func (l *Lexer) peekChar() rune {
	if l.nextBytePos >= len(l.input) {
		return 0
	}
	b := l.input[l.nextBytePos]
	if b < utf8.RuneSelf {
		return rune(l.input[l.nextBytePos])
	}

	r, _ := utf8.DecodeLastRuneInString(l.input[l.nextBytePos:])
	return r
}

func (l *Lexer) Emit() Token {
	l.findNextTokenStart()

	if tok, resolved := ResolveUnambiguousToken(l); resolved {
		return tok
	} else {
		return ResolveAmbiguousToken(l)
	}
}

func (l *Lexer) findNextTokenStart() {
	for l.ch != '\n' && l.ch != '\r' && unicode.IsSpace(l.ch) {
		l.advancePtr()
	}
}

func (l *Lexer) lookAheadRune(n int) rune {
	pos := l.position
	for range n {
		var width int
		b := l.input[pos]
		if b < utf8.RuneSelf {
			width = 1
		} else {
			_, width = utf8.DecodeLastRuneInString(l.input[pos:])
		}

		pos += width
	}
	if pos >= len(l.input) {
		return 0
	}
	return rune(l.input[pos])
}

func (l *Lexer) readUntilWhitespaceOrNewline() string {
	start := l.position
	for !l.isEOF() && !unicode.IsSpace(l.ch) {
		l.advancePtr()
	}
	return l.input[start:l.position]
}

func (l *Lexer) isIdentifierRune() bool {
	return helpers.IsIdentifierRune(l.ch) || l.ch == '\\' || l.ch == '_'
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for !l.isEOF() {
		if !helpers.IsIdentifierRune(l.ch) && !unicode.IsDigit(l.ch) {
			break
		}
		l.advancePtr()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readEncodedNumber() (string, error) {
	start := l.position
	l.advancePtr() // consume '0'

	encoding := l.ch
	l.advancePtr() // consume encoding char (x, b, o)

	var base int
	switch encoding {
	case 'x', 'X':
		base = 16
	case 'b', 'B':
		base = 2
	case 'o', 'O':
		base = 8
	default:
		return l.input[start:l.position], fmt.Errorf("%w: %c", ErrInvalidBase, encoding)
	}

	hasDigits := false
	for !l.isEOF() {
		if !unicode.IsDigit(l.ch) && !helpers.IsIdentifierRune(l.ch) {
			break
		}

		val, ok := getDigitValue(l.ch)
		if !ok || val >= base {
			return l.input[start:l.position], fmt.Errorf("%w: base: '%d', digit: %c", ErrInvalidEncDigit, base, l.ch)
		}
		hasDigits = true
		l.advancePtr()
	}

	if !hasDigits {
		return l.input[start:l.position], ErrMissingDigits
	}

	return l.input[start:l.position], nil
}

func getDigitValue(ch rune) (int, bool) {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0'), true
	}
	if ch >= 'a' && ch <= 'f' {
		return int(ch - 'a' + 10), true
	}
	if ch >= 'A' && ch <= 'F' {
		return int(ch - 'A' + 10), true
	}
	return 0, false
}

func (l *Lexer) readNumber() (string, error) {
	start := l.position

	// Check for encoded numbers (0x, 0b, 0o)
	if l.ch == '0' {
		peek := l.peekChar()
		switch peek {
		case 'x', 'X', 'b', 'B', 'o', 'O':
			return l.readEncodedNumber()
		}
	}

state_integer:
	if unicode.IsDigit(l.ch) {
		l.advancePtr()
		goto state_integer
	}
	if l.ch == '.' {
		l.advancePtr()
		goto state_fraction
	}
	if l.ch == 'e' || l.ch == 'E' {
		l.advancePtr()
		goto state_exponent_start
	}
	goto state_validate

state_fraction:
	if unicode.IsDigit(l.ch) {
		l.advancePtr()
		goto state_fraction
	}
	if l.ch == 'e' || l.ch == 'E' {
		l.advancePtr()
		goto state_exponent_start
	}
	if l.ch == '.' {
		return l.input[start:l.position], ErrInvalidDot
	}
	goto state_validate

state_exponent_start:
	if l.ch == '+' || l.ch == '-' {
		l.advancePtr()
	}
	if !unicode.IsDigit(l.ch) {
		return l.input[start:l.position], ErrInvalidScientific
	}
	goto state_exponent

state_exponent:
	if unicode.IsDigit(l.ch) {
		l.advancePtr()
		goto state_exponent
	}
	if l.ch == '.' {
		return l.input[start:l.position], ErrInvalidDot
	}
	if l.ch == 'e' || l.ch == 'E' {
		return l.input[start:l.position], ErrInvalidScientific
	}
	goto state_validate

state_validate:
	if helpers.IsIdentifierRune(l.ch) {
		return l.input[start:l.position], fmt.Errorf("%w: '%c'", ErrInvalidTrailingChar, l.ch)
	}
	return l.input[start:l.position], nil
}

func (l *Lexer) readString() string {
	start := l.position
	l.advancePtr() // consume opening "
	for l.ch != '"' && !l.isEOF() {
		if l.ch == '\\' && l.peekChar() != 0 {
			l.advancePtr() // consume backslash
			l.advancePtr() // consume escaped char
		} else {
			l.advancePtr()
		}
	}
	l.advancePtr() // consume closing "
	return l.input[start:l.position]
}

func (l *Lexer) readLineComment() string {
	// consume single quote
	l.advancePtr()
	start := l.position
	for l.ch != '\n' && !l.isEOF() {
		l.advancePtr()
	}
	return strings.TrimSpace(l.input[start:l.position])
}

func (l *Lexer) readBlockComment() string {
	// consume "/'"
	l.advancePtr()
	l.advancePtr()
	start := l.position
	end := -1
	for !l.isEOF() {
		if l.ch == '\'' && l.peekChar() == '/' {
			end = l.position
			l.advancePtr()
			l.advancePtr()
			break
		}
		l.advancePtr()
	}
	return strings.TrimSpace(l.input[start:end])
}
