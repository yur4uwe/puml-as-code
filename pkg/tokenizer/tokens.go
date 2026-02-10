package tokenizer

import (
	"unicode"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	NEWLINE
	IDENTIFIER
	STRING
	NUMBER

	CLASS
	STRUCT
	INTERFACE
	ENUM
	PACKAGE
	ANNOTATION
	NOTE
	STEREOTYPE

	LBRACE
	RBRACE
	LPAREN
	RPAREN
	SEMICOLON
	COLON
	COMMA

	VISIBILITY
	MODIFIER

	RELATIONSHIP

	COMMENT
	SEPARATOR
	ALIAS
	START
	END
)

func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case NEWLINE:
		return "NEWLINE"
	case IDENTIFIER:
		return "IDENTIFIER"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case CLASS:
		return "CLASS"
	case STRUCT:
		return "STRUCT"
	case INTERFACE:
		return "INTERFACE"
	case ENUM:
		return "ENUM"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case SEMICOLON:
		return ";"
	case COLON:
		return ":"
	case COMMA:
		return ","
	case VISIBILITY:
		return "VISIBILITY"
	case RELATIONSHIP:
		return "RELATIONSHIP"
	case COMMENT:
		return "COMMENT"
	case START:
		return "@startuml"
	case END:
		return "@enduml"
	case PACKAGE:
		return "PACKAGE"
	case ANNOTATION:
		return "ANNOTATION"
	case NOTE:
		return "NOTE"
	case MODIFIER:
		return "MODIFIER"
	case ALIAS:
		return "ALIAS"
	case STEREOTYPE:
		return "STEREOTYPE"
	case SEPARATOR:
		return "SEPARATOR"
	default:
		return "UNKNOWN"
	}
}

type TokenPos struct {
	line uint
	col  uint
}

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

func TokenFactory(l *Lexer) Token {
	var tok Token
	switch {
	case l.ch == 0:
		tok = Token{Type: EOF, Literal: "", Pos: l.position}
	case l.ch == '{':
		if isLetter(l.peekChar()) {
			tok = l.readModifier()
			break
		}
		tok = Token{Type: LBRACE, Literal: "{", Pos: l.position}
		l.readChar()
	case l.ch == '}':
		l.readChar()
		if l.ch == '-' {
			tok = l.readRelation()
			tok.Literal = "}" + tok.Literal
		} else {
			tok = Token{Type: RBRACE, Literal: "}", Pos: l.position}
		}
	case l.ch == '(':
		tok = Token{Type: LPAREN, Literal: "(", Pos: l.position}
		l.readChar()
	case l.ch == ')':
		tok = Token{Type: RPAREN, Literal: ")", Pos: l.position}
		l.readChar()
	case l.ch == ':':
		tok = Token{Type: COLON, Literal: ":", Pos: l.position}
		l.readChar()
	case l.ch == ',':
		tok = Token{Type: COMMA, Literal: ",", Pos: l.position}
		l.readChar()
	case l.ch == '@':
		tok = l.readUMLBounds()
		l.readChar()
	case l.ch == '\n':
		tok = Token{Type: NEWLINE, Literal: "\n", Pos: l.position}
		l.readChar()
	case l.ch == '"':
		start := l.position
		lit := l.readString()
		tok = Token{Type: STRING, Literal: lit, Pos: start}
	case l.ch == '/' && l.peekChar() == '/':
		start := l.position
		lit := l.readLineComment()
		tok = Token{Type: COMMENT, Literal: lit, Pos: start}
	case l.ch == '<' && l.peekChar() == '<':
		tok = l.readStereotype()
	case isRelationChar(l.ch) || l.ch == '~':
		// Check if this is a relation (includes decorated relations like #--, x--, etc.)
		// or a standalone visibility marker
		if isVisibilityRune(l.ch) && !isRelationChar(l.peekChar()) {
			// It's a visibility marker: single +, -, #, or ~ not followed by a relation char
			tok = Token{Type: VISIBILITY, Literal: string(l.ch), Pos: l.position}
			l.readChar()
		} else {
			// It's a relationship (possibly decorated)
			tok = l.readRelation()
		}
	case l.ch == '\\' || l.ch == '$' || isLetter(l.ch):
		start := l.position
		lit := l.readIdentifier()
		tt := lookupKeyword(lit)
		tok = Token{Type: tt, Literal: lit, Pos: start}
	case unicode.IsDigit(l.ch):
		start := l.position
		lit := l.readNumber()
		tok = Token{Type: NUMBER, Literal: lit, Pos: start}
	default:
		// illegal or unhandled single rune; treat as ILLEGAL token but advance
		tok = Token{Type: ILLEGAL, Literal: string(l.ch), Pos: l.position}
		l.readChar()
	}
	return tok
}
