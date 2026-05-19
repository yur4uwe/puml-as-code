package tokenizer

import (
	"fmt"
	"strings"
	"unicode"
	"yur4uwe/pac/internal/helpers"
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
	expectingSeparatorValue bool
}

func (l lexerState) String() string {
	return fmt.Sprintf("\nposition: %d\nreadPos: %d\nch: %q", l.position, l.readPos, l.ch)
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
		lexerState: lexerState{
			packageSeparator: []rune{'.'},
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
	l.position = l.readPos
	l.readPos++
	return
}

func (l *Lexer) consumeChar(tokenType TokenType, literal string) Token {
	tok := Token{Type: tokenType, Literal: literal, Pos: l.position}
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
	l.findNextTokenStart()

	if l.expectingSeparatorValue {
		l.expectingSeparatorValue = false
		start := l.position
		val := l.readUntilWhitespaceOrNewline()
		if val == "none" {
			l.packageSeparator = nil
		} else {
			l.packageSeparator = []rune(val)
		}
		return Token{Type: SEPARATOR, Literal: val, Pos: start}
	}

	if len(l.packageSeparator) > 0 && l.isPackageSeparator() {
		start := l.position
		return Token{Type: SEPARATOR, Literal: l.consumePackageSeparator(), Pos: start}
	}

	var tok Token
	var resolved bool

	if tok, resolved = ResolveUnambiguousToken(l); !resolved {
		tok = ResolveAmbiguousToken(l)
	}

	// Dynamic state update for "set separator"
	if tok.Type == IDENTIFIER &&
		tok.Literal == "separator" &&
		string(l.input[l.position-len("separator")-4:l.position-len("separator")]) == "set " {
		l.expectingSeparatorValue = true
	}

	return tok
}

func (l *Lexer) findNextTokenStart() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
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
	for l.ch != 0 && !unicode.IsSpace(l.ch) {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readIdentifier() string {
	var result []rune

	for {
		if l.ch == '\\' && l.peekChar() != 0 {
			// Skip the backslash and take the next character literally
			l.readChar()
			result = append(result, l.ch)
			l.readChar()
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
	switch ident {
	// Class declarations
	case "class":
		return CLASS
	case "interface":
		return INTERFACE
	case "enum":
		return ENUM
	case "struct":
		return STRUCT
	case "record":
		return RECORD
	case "dataclass":
		return DATACLASS
	case "exception":
		return EXCEPTION
	case "protocol":
		return PROTOCOL

	// Modifiers and keywords
	case "abstract":
		return ABSTRACT
	case "package", "namespace":
		return PACKAGE
	case "as":
		return ALIAS
	case "annotation":
		return ANNOTATION

	// Documentation and layout
	case "note":
		return NOTE

	// Configuration
	case "skinparam":
		return SKINPARAM

	// Layout grouping
	case "together":
		return TOGETHER

	// Note direction/position (can be treated as identifiers but keeping as keywords for now)
	case "left", "right", "top", "bottom", "up", "down":
		return NOTE_DIRECTION
	case "of", "on":
		return NOTE_POSITION
	case "end":
		return END_BLOCK

	// Actions
	case "hide", "show", "remove", "restore":
		return ACTION

	case "set":
		return SET

	default:
		return IDENTIFIER
	}
}

func (l *Lexer) readNumber() string {
	start := l.position
	// Read leading
	if l.ch == '0' {
		l.readChar()
		if l.ch == 'x' || l.ch == 'X' || l.ch == 'o' || l.ch == 'O' || l.ch == 'b' || l.ch == 'B' {
			l.readChar()
		}
	}

	for unicode.IsDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readString() string {
	start := l.position
	l.readChar() // consume opening "
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' && l.peekChar() != 0 {
			l.readChar() // consume backslash
			l.readChar() // consume escaped char
		} else {
			l.readChar()
		}
	}
	if l.ch == '"' {
		l.readChar() // consume closing "
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readLineComment() string {
	// consume single quote
	l.readChar()
	start := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return strings.TrimSpace(string(l.input[start:l.position]))
}
