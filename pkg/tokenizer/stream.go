package tokenizer

import (
	"fmt"
	"slices"
	"strings"
	"yur4uwe/pac/pkg/parser/ast"
)

var possibleBounds = []string{
	"uml",
	"gantt",
	"mindmap",
	"def",
}

type UnexpectedTokenError struct {
	Expected Token
	Found    Token
}

func (e UnexpectedTokenError) Error() string {
	return fmt.Sprintf("token stream: unexpected token %s(%q) at %d:%d, expected %s(%q)", e.Found.Type, e.Found.Literal, e.Found.Pos.Line, e.Found.Pos.Col, e.Expected.Type, e.Expected.Literal)
}

func unexpectedTokenError(expected Token, found Token) error {
	return UnexpectedTokenError{Expected: expected, Found: found}
}

type TokenStream struct {
	lexer         *Lexer
	leadingTrivia []Token
	buffer        []Token
	sinks         []TokenSink
}

func (ts *TokenStream) Attach(sink TokenSink) func() {
	ts.sinks = append(ts.sinks, sink)
	return func() {
		for i, s := range ts.sinks {
			if s == sink {
				ts.sinks = append(ts.sinks[:i], ts.sinks[i+1:]...)
				return
			}
		}
	}
}

func NewTokenStream(input string) *TokenStream {
	return &TokenStream{
		lexer: NewLexer(input),
		sinks: make([]TokenSink, 0, 2),
	}
}

func (ts *TokenStream) AssertSeq(seq []Token) bool {
	for i, t := range seq {
		next := ts.PeekTokenAt(i)
		if next.Type != t.Type {
			return false
		}
		if t.Literal != "" && next.Literal != t.Literal {
			return false
		}
	}
	return true
}

func (ts *TokenStream) AssertSeqTypes(seq ...TokenType) bool {
	for i, t := range seq {
		next := ts.PeekTokenAt(i)
		if next.Type != t {
			return false
		}
	}
	return true
}

func (ts *TokenStream) TokensToString(toks []Token) string {
	if len(toks) == 0 {
		return ""
	}

	var sb strings.Builder
	var prevTok *Token
	for i := range toks {
		tok := &toks[i]
		if tok.Type == COMMENT {
			continue
		}

		if prevTok != nil {
			startNext := tok.Pos.Offset
			endPrev := prevTok.Pos.Offset + uint(len(prevTok.Literal))
			if startNext > endPrev {
				sb.WriteString(string(ts.lexer.input[endPrev:startNext]))
			}
		}

		sb.WriteString(tok.Literal)
		prevTok = tok
	}

	return sb.String()
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
	tok := ts.EmitRaw()
	if len(ts.leadingTrivia) > 0 && tok.Type != COMMENT {
		tok.LeadingTrivia = ts.leadingTrivia
	}
	return tok
}

func (ts *TokenStream) EmitRaw() Token {
	tok := ts.PeekTokenAt(0)
	if len(ts.buffer) > 0 {
		ts.buffer = ts.buffer[1:]
	}
	for _, sink := range ts.sinks {
		sink.Receive(tok)
	}
	return tok
}

// Assert checks if the next token matches the given token (Type and Literal if literal is set).
// Does not emit the token.
func (ts *TokenStream) Assert(token Token) bool {
	next := ts.PeekTokenAt(0)
	if next.Type != token.Type {
		return false
	}
	if token.Literal != "" && next.Literal != token.Literal {
		return false
	}
	return true
}

func (ts *TokenStream) AssertType(token TokenType) bool {
	if ts.PeekTokenAt(0).Type != token {
		return false
	}
	return true
}

func (ts *TokenStream) AssertAny(tokens ...Token) bool {
	return slices.ContainsFunc(tokens, ts.Assert)
}

func (ts *TokenStream) AssertAnyType(tokens ...TokenType) bool {
	return slices.ContainsFunc(tokens, ts.AssertType)
}

func (ts *TokenStream) TryConsume(token Token) (Token, bool) {
	if ts.Assert(token) {
		return ts.Emit(), true
	}
	return Token{}, false
}

func (ts *TokenStream) MustConsume(token Token) Token {
	if !ts.Assert(token) {
		panic(fmt.Sprintf("expected %s, got %s", token.Type, ts.PeekTokenAt(0).Type))
	}
	return ts.Emit()
}

func (ts *TokenStream) MustConsumeType(token TokenType) Token {
	if !ts.AssertType(token) {
		panic(fmt.Sprintf("expected %s, got %s", token, ts.PeekTokenAt(0).Type))
	}
	return ts.Emit()
}

func (ts *TokenStream) TryConsumeType(token TokenType) (Token, bool) {
	if ts.AssertType(token) {
		return ts.Emit(), true
	}
	return Token{}, false
}

func (ts *TokenStream) ConsumeUntil(token ...Token) Token {
	for !ts.AssertAny(token...) && !ts.AssertType(EOF) {
		ts.Emit()
	}
	return ts.PeekTokenAt(0)
}

func (ts *TokenStream) ConsumeUntilType(token ...TokenType) Token {
	for !ts.AssertAnyType(token...) && !ts.AssertType(EOF) {
		ts.Emit()
	}
	return ts.PeekTokenAt(0)
}

// readBetween handles the common pattern of [start markers]...[end markers]
func (ts *TokenStream) readBetween(start, end []Token, emitter func() Token) (string, error) {
	// 1. Check if start markers match
	if !ts.AssertSeq(start) {
		return "", fmt.Errorf("failed to assert start marker")
	}

	// 2. Consume start markers
	for range start {
		emitter()
	}

	// 3. Read content until end markers
	var tc TokenCollector
	detach := ts.Attach(&tc)
	defer detach()

	for {
		if ts.AssertType(EOF) || ts.AssertType(NEWLINE) {
			return "", fmt.Errorf("unexpected EOF or newline")
		}

		if ts.AssertSeq(end) {
			break
		}

		emitter()
	}

	res := ts.TokensToString(tc.tokens)

	// Consume end markers
	for range end {
		emitter()
	}

	return res, nil
}

func (ts *TokenStream) TryReadModifier() (string, error) {
	start := []Token{{Type: LBRACE}}
	end := []Token{{Type: RBRACE}}
	return ts.readBetween(start, end, ts.Emit)
}

func (ts *TokenStream) TryReadStereotype() (string, error) {
	start := []Token{{Type: LANGLE}, {Type: LANGLE}}
	end := []Token{{Type: RANGLE}, {Type: RANGLE}}
	return ts.readBetween(start, end, ts.Emit)
}

func (ts *TokenStream) TryReadGeneric() (string, error) {
	start := []Token{{Type: LANGLE}}
	end := []Token{{Type: RANGLE}}
	return ts.readBetween(start, end, ts.Emit)
}

func (ts *TokenStream) TryReadClassSeparator() (string, error) {
	var start, end []Token
	switch ts.PeekTokenAt(0).Type {
	case HYPHEN:
		start = []Token{{Type: HYPHEN}, {Type: HYPHEN}}
		end = []Token{{Type: HYPHEN}, {Type: HYPHEN}}
	case DOT:
		start = []Token{{Type: DOT}, {Type: DOT}}
		end = []Token{{Type: DOT}, {Type: DOT}}
	case EQUALS:
		start = []Token{{Type: EQUALS}, {Type: EQUALS}}
		end = []Token{{Type: EQUALS}, {Type: EQUALS}}
	case UNDERSCORE:
		start = []Token{{Type: UNDERSCORE}, {Type: UNDERSCORE}}
		end = []Token{{Type: UNDERSCORE}, {Type: UNDERSCORE}}
	default:
		return "", fmt.Errorf("unexpected class separator")
	}

	// separator can have no description
	// Here we find out if its true
	if ts.AssertSeq(append(start, Token{Type: NEWLINE})) {
		return "", nil
	}

	return ts.readBetween(start, end, ts.Emit)
}

func (ts *TokenStream) TryReadTag() (string, error) {
	if !ts.AssertType(DOLLAR) {
		return "", unexpectedTokenError(Token{Type: DOLLAR}, ts.PeekTokenAt(0))
	}
	ts.Emit() // consume $
	return ts.Emit().Literal, nil
}

func (ts *TokenStream) IsDiagramBound() bool {
	if !ts.AssertType(AT) {
		return false
	}
	tok := ts.PeekTokenAt(1)
	if tok.Type != IDENTIFIER {
		return false
	}
	return strings.HasPrefix(tok.Literal, "start") || strings.HasPrefix(tok.Literal, "end")
}

func (ts *TokenStream) TryReadDiagramBounds() (string, error) {
	if !ts.IsDiagramBound() {
		return "", fmt.Errorf("not a diagram bound")
	}
	// For tests, we want to return the literal like "startuml"
	ts.Emit() // consume @
	return ts.Emit().Literal, nil
}

func (ts *TokenStream) ReadDiagramBounds() (ast.DiagramBound, error) {
	if _, consumed := ts.TryConsumeType(AT); !consumed {
		return ast.DiagramBound{}, unexpectedTokenError(Token{Type: AT}, ts.PeekTokenAt(0))
	}

	tok, ok := ts.TryConsumeType(IDENTIFIER)
	if !ok {
		return ast.DiagramBound{}, unexpectedTokenError(Token{Type: IDENTIFIER}, ts.PeekTokenAt(0))
	}

	if !strings.HasPrefix(tok.Literal, "start") &&
		!strings.HasPrefix(tok.Literal, "end") {
		return ast.DiagramBound{}, fmt.Errorf("invalid bounding marker for diagram, expected something that starts with 'start' or 'end'")
	}

	diag := ast.DiagramBound{
		Opts: make(map[string]string),
	}

	if typ, found := strings.CutPrefix(tok.Literal, "start"); found {
		diag.IsStart = true
		diag.Type = typ
	} else if typ, found := strings.CutPrefix(tok.Literal, "end"); found {
		diag.IsStart = false
		diag.Type = typ
		// Fast return for end markers
		return diag, nil
	}

	// Check for legacy name syntax: @startuml NAME
	if !ts.AssertType(LPAREN) && !ts.AssertType(LBRACE) {
		diag.Name = ts.ReadRawUntilNewline()
		return diag, nil
	}

	readKvp := func() error {
		keyTok, ok := ts.TryConsumeType(IDENTIFIER)
		if !ok {
			return fmt.Errorf("expected a key")
		}
		if _, ok := ts.TryConsumeType(EQUALS); !ok {
			return unexpectedTokenError(Token{Type: EQUALS}, ts.PeekTokenAt(0))
		}
		valTok := ts.Emit()
		if keyTok.Literal == "id" {
			diag.ID = valTok.Literal
		} else if _, ok := diag.Opts[keyTok.Literal]; ok {
			return fmt.Errorf("duplicate key in diagram bounds: %s", keyTok.Literal)
		} else {
			diag.Opts[keyTok.Literal] = valTok.Literal
		}
		return nil
	}

	// First tag (id=TAG, key=value, ...)
	if _, consumed := ts.TryConsumeType(LPAREN); consumed {
		for !ts.AssertType(RPAREN) && !ts.AssertType(EOF) && !ts.AssertType(NEWLINE) {
			if err := readKvp(); err != nil {
				return diag, err
			}
			if _, ok := ts.TryConsumeType(COMMA); !ok {
				break
			}
		}
		if _, ok := ts.TryConsumeType(RPAREN); !ok {
			return diag, fmt.Errorf("expected closing parenthesis in diagram bounds")
		}
	}

	// Check for legacy syntax: @startuml(id=tag) NAME, again
	if !ts.AssertType(LBRACE) {
		diag.Name = ts.ReadRawUntilNewline()
		return diag, nil
	}

	// Then options {filename, caption, key=value, ...}
	if _, consumed := ts.TryConsumeType(LBRACE); consumed {
		// read filename
		buf := TokenCollector{}
		detach := ts.Attach(&buf)
		tok := ts.ConsumeUntilType(EOF, NEWLINE, COMMA, RBRACE)
		if tok.Type == EOF || tok.Type == NEWLINE {
			return diag, fmt.Errorf("unexpected EOF or newline")
		}
		detach()
		diag.Name = ts.TokensToString(buf.tokens)

		// We expect that there will not be a equal sign in the caption
		// as the second token. We assume that that is a kvp
		if ts.PeekTokenAt(2).Type != EQUALS {
			ts.TryConsumeType(COMMA)
			// read caption
			buf := TokenCollector{}
			detach := ts.Attach(&buf)
			tok := ts.ConsumeUntilType(EOF, NEWLINE, COMMA, RBRACE)
			if tok.Type == EOF || tok.Type == NEWLINE {
				return diag, fmt.Errorf("unexpected EOF or newline")
			}
			detach()
			diag.Opts["caption"] = ts.TokensToString(buf.tokens)
			fmt.Println("Caption: ", diag.Opts["caption"])
		}

		// Handle key=value pairs or more options if needed
		for {
			if !ts.AssertType(COMMA) {
				break
			}
			ts.TryConsumeType(COMMA)
			if err := readKvp(); err != nil {
				return diag, err
			}
		}

		if _, ok := ts.TryConsumeType(RBRACE); !ok {
			return diag, unexpectedTokenError(Token{Type: RBRACE}, ts.PeekTokenAt(0))
		}
	}

	if ts.AssertType(LPAREN) {
		return diag, fmt.Errorf("possibly incorrect order of tools and id")
	}

	if !ts.AssertType(NEWLINE) {
		return diag, fmt.Errorf("expected newline after diagram bounds")
	}

	return diag, nil
}

func (ts *TokenStream) collectUntilNewline(emitter func() Token) []Token {
	if ts.AssertType(EOF) || ts.AssertType(NEWLINE) {
		return nil
	}
	buf := []Token{}
	for {
		tok := emitter()
		buf = append(buf, tok)
		if tok.Type == NEWLINE || tok.Type == EOF {
			break
		}
	}
	return buf
}

func (ts *TokenStream) ReadUntilNewline() string {
	toks := ts.collectUntilNewline(ts.EmitRaw)
	if len(toks) == 0 {
		return ""
	}

	var filtered []Token
	for _, tok := range toks {
		if tok.Type == COMMENT {
			break
		}
		if tok.Type == NEWLINE || tok.Type == EOF {
			continue
		}
		filtered = append(filtered, tok)
	}

	return strings.TrimSpace(ts.TokensToString(filtered))
}

func (ts *TokenStream) ReadRawUntilNewline() string {
	toks := ts.collectUntilNewline(ts.EmitRaw)
	if len(toks) == 0 {
		return ""
	}

	var filtered []Token
	for _, tok := range toks {
		if tok.Type == NEWLINE || tok.Type == EOF {
			continue
		}
		filtered = append(filtered, tok)
	}

	if len(filtered) == 0 {
		return ""
	}

	start := filtered[0].Pos.Offset
	end := filtered[len(filtered)-1].Pos.Offset + uint(len(filtered[len(filtered)-1].Literal))
	str := string(ts.lexer.input[start:end])
	return strings.TrimSpace(str)
}

func (ts *TokenStream) ReadMultilineComment() (string, error) {
	return "", fmt.Errorf("multiline comment not found")
}

// ReadBlock assumes that end tokens are first in the line
// So actual end markers are implicitly append([]Token{NEWLINE}, ...endMark)
func (ts *TokenStream) ReadBlock(endMark ...Token) (string, error) {
	tok := ts.EmitRaw()
	startPos := tok.Pos
	for tok.Type != EOF {
		if tok.Type == NEWLINE && ts.AssertSeq(endMark) {
			endPos := tok.Pos
			for range endMark {
				ts.EmitRaw() // consume end markers
			}
			return string(ts.lexer.input[startPos.Offset:endPos.Offset]), nil
		}
		tok = ts.EmitRaw()
	}
	return "", fmt.Errorf("block end not found")
}
