package tokenizer

import (
	"strings"
	"unicode"
)

type lexerMode byte

const (
	MODE_DEFAULT lexerMode = 1 << iota
	MODE_NOTE
	MODE_LABEL
	MODE_CLASS
)

func (l *Lexer) CurrentMode() lexerMode {
	if len(l.modeStack) == 0 {
		return MODE_DEFAULT
	}
	return l.modeStack[len(l.modeStack)-1]
}

func (l *Lexer) PushMode(mode lexerMode) {
	l.modeStack = append(l.modeStack, mode)
}

func (l *Lexer) PopMode() lexerMode {
	if len(l.modeStack) == 0 {
		return MODE_DEFAULT
	}
	mode := l.modeStack[len(l.modeStack)-1]
	l.modeStack = l.modeStack[:len(l.modeStack)-1]
	return mode
}

type Lexer struct {
	input           []rune
	position        int
	readPos         int
	ch              rune
	modeStack       []lexerMode
	isMultilineNote bool // Track if current note is multi-line (has end note)
}

type lexerState struct {
	position int
	readPos  int
	ch       rune
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:     []rune(input),
		modeStack: make([]lexerMode, 0, 5),
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

func (l *Lexer) jumpToPosition(pos int) {
	l.position = pos
	l.readPos = pos + 1
	l.readChar()
}

func (l *Lexer) NextToken() Token {
	l.findNextTokenStart()

	if tok, resolved := ResolveUnambiguousToken(l); resolved {
		return tok
	}

	if tok, resolved := ResolveContextAwareToken(l); resolved {
		return tok
	}

	return ResolveAmbiguousToken(l)
}

func (l *Lexer) findNextTokenStart() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func isIdentifierRune(ch rune) bool {
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
		} else if isIdentifierRune(l.ch) || unicode.IsDigit(l.ch) {
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
	case "package":
		return PACKAGE
	case "as":
		return ALIAS
	case "annotation":
		return ANNOTATION

	// Documentation and layout
	case "note":
		return NOTE
	case "stereotype":
		return STEREOTYPE

	// Visibility and display control
	case "hide":
		return HIDE
	case "show":
		return SHOW
	case "remove":
		return REMOVE
	case "restore":
		return RESTORE

	// Configuration
	case "skinparam":
		return SKINPARAM
	case "set":
		return SET

	// Layout grouping
	case "together":
		return TOGETHER

	default:
		return IDENTIFIER
	}
}

func lookupNoteKeyword(input string) TokenType {
	switch input {
	case "left", "right", "top", "bottom":
		return NOTE_DIRECTION
	case "of", "on":
		return NOTE_POSITION
	case "link":
		return NOTE_LINK
	case "end":
		return END_BLOCK
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

func (l *Lexer) readLabel() string {
	start := l.position
	for isIdentifierRune(l.ch) || unicode.IsDigit(l.ch) || l.ch == ' ' || l.ch == '<' || l.ch == '>' {
		l.readChar()
	}
	if strings.ToLower(string(l.input[l.position-2:l.position])) == "__" {
		l.jumpToPosition(l.position - 3)
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readNoteLine() string {
	// Reads text content within a note until end of line
	start := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return strings.TrimSpace(string(l.input[start:l.position]))
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
	// consume single quote
	l.readChar()
	start := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return strings.TrimSpace(string(l.input[start:l.position]))
}

func isRelationLineChar(ch rune) bool {
	switch ch {
	case '-', '.':
		return true
	default:
		return false
	}
}

func isRelationLineStartChar(ch rune) bool {
	switch ch {
	case '-', '.', '<', 'o', '*', '#', '+', '}', 'x', '^':
		return true
	default:
		return false
	}
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
	switch lit {
	case "left", "right", "up", "down", "l", "r", "u", "d":
		return true
	default:
		return false
	}
}

func isInlineRelationLetter(input []rune, pos int) bool {
	if pos < 0 || pos >= len(input) {
		return false
	}
	if input[pos] != 'o' && input[pos] != 'x' {
		return false
	}
	prevRel := pos > 0 && isRelationChar(input[pos-1])
	nextRel := pos+1 < len(input) && isRelationChar(input[pos+1])
	return prevRel || nextRel
}

func (l *Lexer) readRelation() Token {
	start := l.position

	// Consume relation runes; stop before direction keywords.
	for isRelationChar(l.ch) {
		if isIdentifierRune(l.ch) && !isInlineRelationLetter(l.input, l.position) {
			break
		}
		l.readChar()
	}

	// Handle optional direction keyword (e.g. -left->).
	if isIdentifierRune(l.ch) {
		state := l.dumpState()
		direction := l.readIdentifier()

		if isRelDirection(direction) && isRelationChar(l.ch) {
			for isRelationChar(l.ch) {
				l.readChar()
			}
			return Token{Type: RELATIONSHIP, Literal: string(l.input[start:l.position]), Pos: start}
		}

		// Not a direction: restore so identifier can be read next.
		l.restoreState(state)

		// Only a single '-' can represent visibility here.
		if l.input[start] == '-' && state.position == start+1 {
			return Token{Type: VISIBILITY, Literal: "-", Pos: start}
		}
	}

	// Safety check: ensure we consumed at least one character (prevents infinite loops)
	literal := string(l.input[start:l.position])
	if literal == "" {
		return Token{Type: ILLEGAL, Literal: string(l.ch), Pos: start}
	}

	return Token{Type: RELATIONSHIP, Literal: literal, Pos: start}
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

func isClassSeparator(ch rune) bool {
	return ch == '-' || ch == '=' || ch == '.' || ch == '_'
}

func (l *Lexer) keywordModeSwitcher(tt TokenType, lit string) {
	switch tt {
	case CLASS, ABSTRACT:
		l.PushMode(MODE_CLASS)
	case NOTE:
		l.PushMode(MODE_NOTE)
		l.isMultilineNote = false // Default to single-line until we see NOTE_POSITION
	}
}
