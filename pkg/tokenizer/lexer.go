package tokenizer

import "unicode"

type Lexer struct {
	input    []rune
	position int
	readPos  int
	ch       rune
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.position = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() Token {
	l.findNextTokenStart()

	return TokenFactory(l)
}

func (l *Lexer) findNextTokenStart() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || unicode.IsDigit(l.ch) {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func lookupKeyword(ident string) TokenType {
	switch ident {
	case "class":
		return CLASS
	case "interface":
		return INTERFACE
	case "enum":
		return ENUM
	default:
		return IDENTIFIER
	}
}

func (l *Lexer) readNumber() string {
	start := l.position
	for unicode.IsDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readString() string {
	// assumes l.ch == '"'
	start := l.position
	l.readChar() // consume opening "
	for l.ch != '"' && l.ch != 0 {
		// note: no escape processing for now
		l.readChar()
	}
	if l.ch == '"' {
		l.readChar() // consume closing "
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readLineComment() string {
	start := l.position
	// consume both slashes
	l.readChar()
	l.readChar()
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func isRelationChar(ch rune) bool {
	switch ch {
	case '-', '.', '<', '>', '|', 'o', '*':
		return true
	default:
		return false
	}
}

func (l *Lexer) readRelation() string {
	start := l.position
	for isRelationChar(l.ch) {
		l.readChar()
	}
	return string(l.input[start:l.position])
}

func isVisibilityRune(ch rune) bool {
	return ch == '+' || ch == '-' || ch == '#' || ch == '~'
}

func (l *Lexer) readUMLBounds() Token {
	start := l.position
	l.readChar() // consume '@'
	ident := l.readIdentifier()
	switch ident {
	case "startuml":
		return Token{Type: START, Literal: ident, Pos: start}
	case "enduml":
		return Token{Type: END, Literal: ident, Pos: start}
	default:
		return Token{Type: IDENTIFIER, Literal: ident, Pos: start}
	}
}
