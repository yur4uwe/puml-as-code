package tokenizer

import (
	"fmt"
	"strings"
)

type TokenStream struct {
	lexer  *Lexer
	buffer []Token
}

func NewTokenStream(input string) *TokenStream {
	return &TokenStream{
		lexer: NewLexer(input),
	}
}

func (ts *TokenStream) PeekTokenAt(idx int) Token {
	if len(ts.buffer) <= idx {
		for i := len(ts.buffer); i <= idx; i++ {
			tok := ts.lexer.Emit()
			ts.buffer = append(ts.buffer, tok)
		}
	}
	return ts.buffer[idx]
}

func (ts *TokenStream) Emit() Token {
	tok := ts.PeekTokenAt(0)
	if len(ts.buffer) > 0 {
		ts.buffer = ts.buffer[1:]
	}
	return tok
}

func (ts *TokenStream) Assert(token TokenType) error {
	if ts.PeekTokenAt(0).Type != token {
		return fmt.Errorf("expected %s, found %s", token, ts.PeekTokenAt(0).Type)
	}
	ts.Emit()
	return nil
}

func (ts *TokenStream) Consume(token TokenType) (Token, bool) {
	if ts.PeekTokenAt(0).Type == token {
		return ts.Emit(), true
	}
	return Token{}, false
}

// readBetween handles the common pattern of [start markers]...[end markers]
func (ts *TokenStream) readBetween(start, end []TokenType) (string, bool) {
	assertSeq := func(seq []TokenType) bool {
		for i, t := range seq {
			if ts.PeekTokenAt(i).Type != t {
				return false
			}
		}
		return true
	}

	// 1. Check if start markers match
	if !assertSeq(start) {
		return "", false
	}

	// 2. Consume start markers
	for range start {
		ts.Emit()
	}

	// 3. Read content until end markers
	var sb strings.Builder
	for {
		if ts.PeekTokenAt(0).Type == EOF || ts.PeekTokenAt(0).Type == NEWLINE {
			return "", false
		}

		// Check if we hit the end markers
		if !assertSeq(end) {
			sb.WriteString(ts.Emit().Literal)
			continue
		}

		// If we are here, we found the end markers
		for range end {
			ts.Emit()
		}
		break
	}

	return sb.String(), true
}

func (ts *TokenStream) TryReadModifier() (string, bool) {
	return ts.readBetween([]TokenType{LBRACE}, []TokenType{RBRACE})
}

func (ts *TokenStream) TryReadStereotype() (string, bool) {
	return ts.readBetween([]TokenType{LANGLE, LANGLE}, []TokenType{RANGLE, RANGLE})
}

func (ts *TokenStream) TryReadGeneric() (string, bool) {
	return ts.readBetween([]TokenType{LANGLE}, []TokenType{RANGLE})
}

func (ts *TokenStream) TryReadClassSeparator() (string, bool) {
	switch ts.PeekTokenAt(0).Type {
	case HYPHEN:
		return ts.readBetween([]TokenType{HYPHEN, HYPHEN}, []TokenType{HYPHEN, HYPHEN})
	case DOT:
		return ts.readBetween([]TokenType{DOT, DOT}, []TokenType{DOT, DOT})
	case EQUALS:
		return ts.readBetween([]TokenType{EQUALS, EQUALS}, []TokenType{EQUALS, EQUALS})
	case UNDERSCORE:
		return ts.readBetween([]TokenType{UNDERSCORE, UNDERSCORE}, []TokenType{UNDERSCORE, UNDERSCORE})
	default:
		return "", false
	}
}

func (ts *TokenStream) TryReadTag() (string, bool) {
	if ts.PeekTokenAt(0).Type != DOLLAR {
		return "", false
	}
	ts.Emit() // consume $
	return ts.Emit().Literal, true
}

func (ts *TokenStream) TryReadDiagramBounds() (string, bool) {
	if ts.PeekTokenAt(0).Type != AT {
		return "", false
	}
	ts.Emit() // consume @
	if ts.PeekTokenAt(0).Type != IDENTIFIER {
		return "", false
	}
	return ts.Emit().Literal, true
}
