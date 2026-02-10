package tokenizer

import (
	"unicode"
)

type Lexer struct {
	input    []rune
	position int
	readPos  int
	ch       rune
}

type lexerState struct {
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

func (l *Lexer) dumpState() lexerState {
	return lexerState{
		position: l.position,
		readPos:  l.readPos,
		ch:       l.ch,
	}
}

func (l *Lexer) restoreState(state lexerState) {
	l.position = state.position
	l.readPos = state.readPos
	l.ch = state.ch
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
	var result []rune

	for {
		if l.ch == '\\' && l.peekChar() != 0 {
			// Skip the backslash and take the next character literally
			l.readChar()
			result = append(result, l.ch)
			l.readChar()
		} else if isLetter(l.ch) || unicode.IsDigit(l.ch) {
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
	case "class":
		return CLASS
	case "interface":
		return INTERFACE
	case "enum":
		return ENUM
	case "abstract":
		return MODIFIER
	case "package":
		return PACKAGE
	case "as":
		return ALIAS
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
	case '-', '.', '<', '>', '|', 'o', '*', '#', '+', '}', '{', 'x', '^':
		return true
	default:
		return false
	}
}

func isRelDirection(lit string) bool {
	return lit == "left" || lit == "right" || lit == "up" || lit == "down"
}

func isLetterLikeRelChar(char rune) bool {
	return char == 'o' || char == 'x'
}

// Possible optimization is to return other tokens as this function
// can effectively figure out the token it confused for relation
// for now leaving state dump is easy enough to reason about
func (l *Lexer) readRelation() Token {
	start := l.position

	for isRelationChar(l.ch) && (!isLetter(l.ch) || isLetterLikeRelChar(l.ch) && (isRelationChar(l.input[l.position-1]) || isRelationChar(l.input[l.position+1]))) {
		l.readChar()
	}

	// Check if we have a direction keyword
	if isLetter(l.ch) {
		// Peek ahead to see if this looks like a direction
		// Save current position to potentially backtrack
		direction := l.readIdentifier()

		if isRelDirection(direction) && isRelationChar(l.ch) {
			// consume remaining relation chars
			for isRelationChar(l.ch) {
				l.readChar()
			}
		} else {
			// only -IDENTIFIER can reach this point so we can return VISIBILITY and restore to IDENTIFIER start
			l.restoreState(lexerState{
				ch:       l.input[start+1],
				position: start + 1,
				readPos:  start + 2,
			})
			return Token{Type: VISIBILITY, Literal: "-", Pos: start}
		}
	}

	return Token{Type: RELATIONSHIP, Literal: string(l.input[start:l.position]), Pos: start}
}

func isVisibilityRune(ch rune) bool {
	return ch == '+' || ch == '-' || ch == '#' || ch == '~'
}

// peekAhead looks ahead n positions from current readPos
func (l *Lexer) peekAhead(n int) rune {
	pos := l.readPos + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
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

func (l *Lexer) readModifier() Token {
	// starts with '{'
	start := l.position
	l.readChar() // consume '{'
	for l.ch != '}' && l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '}' {
		l.readChar() // consume '}'
	}
	return Token{Type: MODIFIER, Literal: string(l.input[start:l.position]), Pos: start}
}

func (l *Lexer) readStereotype() Token {
	// starts with '<<'
	start := l.position
	l.readChar() // consume '<'
	l.readChar() // consume '<'
	for l.ch != '>' && l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '>' {
		l.readChar() // consume '>'
		l.readChar() // consume '>'
	}
	return Token{Type: STEREOTYPE, Literal: string(l.input[start:l.position]), Pos: start}
}
