package tokenizer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

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
	input []rune
	lexerState
}

type lexerState struct {
	position                int
	readPos                 int
	ch                      rune
	packageSeparator        []rune
	isDefaultSeparator      bool
	expectingSeparatorValue bool
	rawModeDelimiter        []rune
	previousTokenType       TokenType
	line                    uint
	col                     uint
}

func (l lexerState) getPos() TokenPos {
	return TokenPos{
		Line:   l.line,
		Col:    l.col,
		Offset: uint(l.position),
	}
}

func (l lexerState) String() string {
	return fmt.Sprintf("\nposition: %s\nreadPos: %d\nch: %q", l.getPos().String(), l.readPos, l.ch)
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
		lexerState: lexerState{
			packageSeparator:   []rune{'.'},
			isDefaultSeparator: true,
			previousTokenType:  ILLEGAL,
		},
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() (ch rune) {
	ch = l.ch
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}

	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}

	l.position = l.readPos
	l.readPos++
	return
}

func (l *Lexer) isEOF() bool {
	return l.ch == 0
}

func (l *Lexer) consumeChar(tokenType TokenType, literal string) Token {
	tok := Token{Type: tokenType, Literal: literal, Pos: l.getPos()}
	l.readChar()
	return tok
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) Emit() Token {
	if len(l.rawModeDelimiter) > 0 {
		return Token{Type: STRING, Literal: l.readRawUntilDelimiter(), Pos: l.getPos()}
	}

	l.findNextTokenStart()

	if l.expectingSeparatorValue {
		l.expectingSeparatorValue = false
		l.isDefaultSeparator = false // Explicitly mark as non-default
		start := l.getPos()
		val := l.readUntilWhitespaceOrNewline()
		if val == "none" {
			l.packageSeparator = nil
		} else {
			l.packageSeparator = []rune(val)
		}
		return Token{Type: IDENTIFIER, Literal: val, Pos: start}
	}

	if len(l.packageSeparator) > 0 && l.isPackageSeparator() {
		// Prioritize PACKAGE_SEPARATOR if:
		// 1. It's multi-character (like the default "::" in some tests, or custom ones)
		// 2. It's explicitly set (not default)
		// 3. It's NOT the default single-character dot (which should be DOT token)

		isMultiChar := len(l.packageSeparator) > 1
		isExplicit := !l.isDefaultSeparator
		isDefaultDot := len(l.packageSeparator) == 1 && l.packageSeparator[0] == '.'

		if isMultiChar || isExplicit || !isDefaultDot {
			start := l.getPos()
			return Token{Type: PACKAGE_SEPARATOR, Literal: l.consumePackageSeparator(), Pos: start}
		}
	}

	var tok Token
	var resolved bool

	if tok, resolved = ResolveUnambiguousToken(l); !resolved {
		tok = ResolveAmbiguousToken(l)
	}

	// Dynamic state update for "set separator"
	if l.previousTokenType == SET_CMD && tok.Type == IDENTIFIER && strings.ToLower(tok.Literal) == "separator" {
		l.expectingSeparatorValue = true
	}
	l.previousTokenType = tok.Type

	return tok
}

func (l *Lexer) findNextTokenStart() {
	for l.ch != '\n' && unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) isPackageSeparator() bool {
	if len(l.packageSeparator) == 0 {
		return false
	}
	for i, r := range l.packageSeparator {
		if l.lookAheadChar(i) != r {
			return false
		}
	}
	return true
}

func (l *Lexer) lookAheadChar(n int) rune {
	pos := l.position + n
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

func (l *Lexer) consumePackageSeparator() string {
	start := l.position
	for range l.packageSeparator {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readUntilWhitespaceOrNewline() string {
	start := l.position
	for !l.isEOF() && !unicode.IsSpace(l.ch) {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) isIdentifierRune() bool {
	return helpers.IsIdentifierRune(l.ch) || l.ch == '\\' || l.ch == '_'
}

func (l *Lexer) readIdentifier() string {
	var result []rune

	for !l.isEOF() {
		if l.ch == '\\' && l.peekChar() != 0 {
			// Skip the backslash and take the next character literally
			l.readChar()
			if l.ch != 0 {
				result = append(result, l.ch)
				l.readChar()
			}
		} else if helpers.IsIdentifierRune(l.ch) || unicode.IsDigit(l.ch) {
			result = append(result, l.ch)
			l.readChar()
		} else {
			break
		}
	}
	return string(result)
}

func lookupKeyword(ident string) TokenType {
	upper := strings.ToUpper(ident)
	// Special cases where keyword string isn't exactly the same as the enum name
	switch upper {
	case "SET":
		return SET_CMD
	case "END":
		return END_BLOCK
	case "AS":
		return ALIAS
	case "LEFT", "RIGHT", "UP", "DOWN", "TOP", "BOTTOM":
		return DIRECTION
	case "OF", "AT", "ON":
		return NOTE_POSITION
	}

	tok, err := TokenTypeString(upper)
	if err != nil {
		return IDENTIFIER
	}
	return tok
}

func (l *Lexer) readEncodedNumber() (string, error) {
	start := l.position
	l.readChar() // consume '0'

	encoding := l.ch
	l.readChar() // consume encoding char (x, b, o)

	var base int
	switch encoding {
	case 'x', 'X':
		base = 16
	case 'b', 'B':
		base = 2
	case 'o', 'O':
		base = 8
	default:
		return string(l.input[start:l.position]), fmt.Errorf("%w: %c", ErrInvalidBase, encoding)
	}

	hasDigits := false
	for !l.isEOF() {
		if !unicode.IsDigit(l.ch) && !helpers.IsIdentifierRune(l.ch) {
			break
		}

		val, ok := getDigitValue(l.ch)
		if !ok || val >= base {
			return string(l.input[start:l.position]), fmt.Errorf("%w: base: '%d', digit: %c", ErrInvalidEncDigit, base, l.ch)
		}
		hasDigits = true
		l.readChar()
	}

	if !hasDigits {
		return string(l.input[start:l.position]), ErrMissingDigits
	}

	return string(l.input[start:l.position]), nil
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
		l.readChar()
		goto state_integer
	}
	if l.ch == '.' {
		l.readChar()
		goto state_fraction
	}
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		goto state_exponent_start
	}
	goto state_validate

state_fraction:
	if unicode.IsDigit(l.ch) {
		l.readChar()
		goto state_fraction
	}
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		goto state_exponent_start
	}
	if l.ch == '.' {
		return string(l.input[start:l.position]), ErrInvalidDot
	}
	goto state_validate

state_exponent_start:
	if l.ch == '+' || l.ch == '-' {
		l.readChar()
	}
	if !unicode.IsDigit(l.ch) {
		return string(l.input[start:l.position]), ErrInvalidScientific
	}
	goto state_exponent

state_exponent:
	if unicode.IsDigit(l.ch) {
		l.readChar()
		goto state_exponent
	}
	if l.ch == '.' {
		return string(l.input[start:l.position]), ErrInvalidDot
	}
	if l.ch == 'e' || l.ch == 'E' {
		return string(l.input[start:l.position]), ErrInvalidScientific
	}
	goto state_validate

state_validate:
	if helpers.IsIdentifierRune(l.ch) {
		return string(l.input[start:l.position]), fmt.Errorf("%w: '%c'", ErrInvalidTrailingChar, l.ch)
	}
	return string(l.input[start:l.position]), nil
}

func (l *Lexer) readString() string {
	l.readChar() // consume opening "
	start := l.position
	for l.ch != '"' && !l.isEOF() {
		if l.ch == '\\' && l.peekChar() != 0 {
			l.readChar() // consume backslash
			l.readChar() // consume escaped char
		} else {
			l.readChar()
		}
	}
	end := l.position
	l.readChar() // consume closing "
	return string(l.input[start:end])
}

func (l *Lexer) readLineComment() string {
	// consume single quote
	l.readChar()
	start := l.position
	for l.ch != '\n' && !l.isEOF() {
		l.readChar()
	}
	return strings.TrimSpace(string(l.input[start:l.position]))
}

func (l *Lexer) readBlockComment() string {
	// consume "/'"
	l.readChar()
	l.readChar()
	var sb strings.Builder
	for !l.isEOF() {
		if l.ch == '\'' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			break
		}
		sb.WriteRune(l.ch)
		l.readChar()
	}
	return sb.String()
}

func (l *Lexer) countFutureSpaces(count_from int) int {
	cnt := 0
	for {
		nextCh := l.lookAheadChar(count_from + cnt)
		if nextCh == ' ' || nextCh == '\t' {
			cnt++
		} else {
			return cnt
		}
	}
}

func (l *Lexer) isRawModeDelimiterFollows() bool {
	offset := 1
	// Skip spaces on the new line to be resilient to indentation, e.g. "  end note"
	offset += l.countFutureSpaces(offset)

	// Match the delimiter case-insensitively
	for i := 0; i < len(l.rawModeDelimiter); i++ {
		nextCh := l.lookAheadChar(offset + i)
		expected := rune(l.rawModeDelimiter[i])
		if unicode.ToLower(nextCh) != expected {
			return false
		}
	}

	// Ensure it's followed by EOF or newline (so "end note-something" doesn't match)
	// allow for spaces after the delimiter but no text
	offset += len(l.rawModeDelimiter)
	offset += l.countFutureSpaces(offset)
	afterDelimiter := l.lookAheadChar(offset)
	if afterDelimiter == 0 || afterDelimiter == '\n' {
		return true
	} else {
		return false
	}
}

func (l *Lexer) readRawUntilDelimiter() string {
	start := l.position

	for !l.isEOF() {
		if l.ch == '\n' &&
			(l.isRawModeDelimiterFollows() ||
				(len(l.rawModeDelimiter) == 1 && l.rawModeDelimiter[0] == '\n')) {
			// Do not consume the delimiter, parser must do it by itself
			break
		}
		l.readChar()
	}

	return string(l.input[start:l.position])
}
