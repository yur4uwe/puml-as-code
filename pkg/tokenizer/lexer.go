package tokenizer

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"yur4uwe/pac/internal/helpers"
)

//go:generate enumer -type=lexerMode -transform=upper
type lexerMode byte

const (
	MODE_DEFAULT lexerMode = iota
	MODE_NOTE
	MODE_LABEL
	MODE_PACKAGE_DEF
	MODE_CLASS_DEF
	MODE_CLASS
	MODE_CLASS_STYLE
	MODE_STYLE
	MODE_ACTION
	MODE_QUALIFIER
)

func (l *Lexer) CurrentMode() lexerMode {
	if len(l.modeStack) == 0 {
		return MODE_DEFAULT
	}
	return l.modeStack[len(l.modeStack)-1]
}

func (l *Lexer) ModeAt(index int) lexerMode {
	if index < 0 || index >= len(l.modeStack) {
		return MODE_DEFAULT
	}
	return l.modeStack[len(l.modeStack)-1-index]
}

func (l *Lexer) PushMode(mode lexerMode) {
	switch mode {
	case MODE_NOTE:
		l.isMultilineNote = true
		l.noteDeclarationComplete = false
	case MODE_ACTION, MODE_STYLE:
		l.isTargetDetermined = false
	case MODE_CLASS_DEF:
		l.isClassNameSet = false
	}
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
	input []rune
	lexerState
}

type lexerState struct {
	position                int
	readPos                 int
	ch                      rune
	modeStack               []lexerMode
	packageSeparator        []rune
	isMultilineNote         bool
	noteDeclarationComplete bool
	isTargetDetermined      bool
	isClassNameSet          bool
}

func (l lexerState) String() string {
	return fmt.Sprintf("\nposition: %d\nreadPos: %d\nch: %q\nmodeStack: %v\nisMultilineNote: %t\nisClassNameSet: %t\nisTargetSet: %t\nisTargetDetermined: %t", l.position, l.readPos, l.ch, l.modeStack, l.isMultilineNote, l.isClassNameSet, l.isTargetDetermined, l.isTargetDetermined)
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
		lexerState: lexerState{
			modeStack:        make([]lexerMode, 0, 5),
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

// consumeChar creates a token for the current character, advances the lexer, and returns the token.
// Use this for single-character tokens to reduce boilerplate.
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

func (l *Lexer) dumpState() lexerState {
	return lexerState{
		position:        l.position,
		readPos:         l.readPos,
		ch:              l.ch,
		modeStack:       l.modeStack,
		isMultilineNote: l.isMultilineNote,
	}
}

func (l *Lexer) restoreState(state lexerState) {
	l.position = state.position
	l.readPos = state.readPos
	l.ch = state.ch
}

func (l *Lexer) jumpToPosition(pos int) {
	// to actualize the lexer state
	// .readChar() needs to be called after updating the position
	// but .readChar() also moves the position forward, so we need to adjust it
	l.position = pos - 1
	l.readPos = pos
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

func (l *Lexer) readClassIdentifier() string {
	ident := l.readIdentifier()
	if l.peekChar() == '(' {

	}
	return ident
}

func (l *Lexer) isPackageSeparator() bool {
	sep := l.lookAhead(len(l.packageSeparator))
	return slices.Equal(sep, l.packageSeparator)
}

func (l *Lexer) readGeneric() string {
	start := l.position

	for l.ch != '>' && l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}
	l.readChar() // consume the closing angle bracket
	return string(l.input[start:l.position])
}

func (l *Lexer) readIdentifier() string {
	var result []rune

	if l.ch == '$' || l.ch == '@' {
		result = append(result, l.ch)
		l.readChar()
	}

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
	// I decided to go with PlantUML beta 1.2023.2
	case "package", "namespace":
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
	case "hide", "show", "remove", "restore":
		return ACTION

	// Configuration
	case "skinparam":
		return SKINPARAM
	case "set":
		return SET_PROPERTY

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
	for helpers.IsIdentifierRune(l.ch) || unicode.IsDigit(l.ch) || l.ch == ' ' || l.ch == '<' || l.ch == '>' {
		l.readChar()
	}
	// Check if label ends with __ (which is a separator token, not part of the label)
	if l.position >= 2 && l.position-2 >= start {
		lastTwo := string(l.input[l.position-2 : l.position])
		if lastTwo == "__" {
			// Rewind to exclude the trailing __
			l.jumpToPosition(l.position - 2)
		}
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

// Assumes to be called after encountering a double quote
// reads a string until the closing double quote or end of input
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

func (l *Lexer) readRelation() Token {
	start := l.position

	// Consume relation runes; stop before direction keywords or bracketed styles.
	for helpers.IsRelationChar(l.ch) {
		if helpers.IsIdentifierRune(l.ch) && !helpers.IsInlineRelationLetter(l.input, l.position) {
			break
		}
		l.readChar()
	}

	// Handle bracketed styles (e.g. -[bold]->, -[#red,thickness=2]->)
	if l.ch == '[' {
		l.readChar() // consume '['
		// Read until closing ']', consuming everything inside (colors, styles, thickness, etc.)
		for l.ch != ']' && l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		if l.ch == ']' {
			l.readChar() // consume ']'
		}
		// Continue reading the rest of the relationship arrow
		for helpers.IsRelationChar(l.ch) {
			l.readChar()
		}
		return Token{Type: RELATIONSHIP, Literal: string(l.input[start:l.position]), Pos: start}
	}

	// Handle optional direction keyword (e.g. -left->, -[left]->).
	if helpers.IsIdentifierRune(l.ch) || l.ch == '[' {
		if l.ch == '[' {
			l.readChar()
		}
		state := l.dumpState()
		direction := l.readIdentifier()
		if l.ch == ']' {
			l.readChar()
		}

		if helpers.IsRelDirection(direction) && helpers.IsRelationChar(l.ch) {
			for helpers.IsRelationChar(l.ch) {
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

// lookAhead looks ahead n positions from current readPos
func (l *Lexer) lookAhead(n int) []rune {
	end := min(l.position+n, len(l.input))
	return l.input[l.position:end]
}

func (l *Lexer) readUMLBounds() Token {
	start := l.position
	ch := string(l.readChar()) // consume '@'
	ident := l.readIdentifier()
	switch ident {
	case "startuml":
		return Token{Type: START, Literal: ident, Pos: start}
	case "enduml":
		return Token{Type: END, Literal: ident, Pos: start}
	case "unlinked":
		return Token{Type: TARGET, Literal: ch + ident, Pos: start}
	default:
		return Token{Type: ILLEGAL, Literal: ch + ident, Pos: start}
	}
}

func (l *Lexer) readModifier() Token {
	// starts with '{'
	start := l.position
	l.readChar() // consume '{'
	content := l.readUntil('}', '\n', '\r')
	literal := "{" + content
	if l.ch == '}' {
		literal += "}"
		l.readChar() // consume '}'
	}
	return Token{Type: MODIFIER, Literal: strings.TrimSpace(literal), Pos: start}
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
	return Token{Type: STEREOTYPE, Literal: strings.TrimSpace(string(l.input[start:l.position])), Pos: start}
}

func (l *Lexer) keywordModeSwitcher(tt TokenType) {
	switch tt {
	case CLASS, INTERFACE, ENUM, STRUCT, RECORD, DATACLASS, EXCEPTION, PROTOCOL, ANNOTATION:
		l.PushMode(MODE_CLASS_DEF)
	case PACKAGE:
		l.PushMode(MODE_PACKAGE_DEF)
	case NOTE:
		l.PushMode(MODE_NOTE)
	case ACTION:
		l.PushMode(MODE_ACTION)
	case SKINPARAM:
		l.PushMode(MODE_STYLE)
	}
}

func (l *Lexer) readProperty() string {
	// 'set' is already consumed
	// next consume name of the property
	l.findNextTokenStart()
	name := l.readIdentifier()
	l.findNextTokenStart()
	value := l.readUntil('\n', '\r')
	if name == "separator" {
		l.packageSeparator = []rune(value)
	}
	return fmt.Sprintf("%s=%s", name, value)
}

func (l *Lexer) readUntil(runes ...rune) string {
	var sb strings.Builder
	for !slices.Contains(runes, l.ch) && l.ch != 0 {
		sb.WriteRune(l.readChar())
	}
	return sb.String()
}

func (l *Lexer) isActionAspect(lit string) bool {
	switch lit {
	case "members", "circle", "methods", "fields", "stereotype":
		return true
	default:
		return l.isTargetDetermined
	}
}
