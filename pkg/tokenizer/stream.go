package tokenizer

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"yur4uwe/pac/pkg/parser/keyword"
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
	lexer             *Lexer
	leadingTrivia     []Token
	buffer            []Token
	sinks             []TokenSink
	rawModeTerminator []rune
	PackageSeparator  string
}

func NewTokenStream(input string) *TokenStream {
	return &TokenStream{
		lexer:            NewLexer(input),
		PackageSeparator: ".",
		sinks:            make([]TokenSink, 0, 2),
	}
}

func (ts *TokenStream) matchPackageSeparatorAt(startIdx int) (int, bool) {
	if ts.PackageSeparator == "" {
		return 0, false
	}
	var sb strings.Builder
	count := 0
	var prevTok Token
	hasPrev := false
	for {
		tok := ts.PeekTokenAt(startIdx + count)
		if tok.Type == EOF || tok.Type == NEWLINE {
			break
		}
		if hasPrev {
			if tok.Pos.Offset != prevTok.Pos.Offset+uint(len([]rune(prevTok.Literal))) {
				break
			}
		}
		sb.WriteString(tok.Literal)
		count++
		prevTok = tok
		hasPrev = true
		if sb.String() == ts.PackageSeparator {
			return count, true
		}
		if len(sb.String()) >= len(ts.PackageSeparator) {
			break
		}
	}
	return 0, false
}

func (ts *TokenStream) TryConsumePackageSeparator() (string, bool) {
	count, ok := ts.matchPackageSeparatorAt(0)
	if !ok {
		return "", false
	}
	for range count {
		ts.Emit()
	}
	return ts.PackageSeparator, true
}

func (ts *TokenStream) PeekTokenAt(idx int) Token {
	for len(ts.buffer) <= idx {
		tok := ts.lexer.Emit()
		ts.buffer = append(ts.buffer, tok)
		if tok.Type == EOF {
			break
		}
	}
	if idx < len(ts.buffer) {
		return ts.buffer[idx]
	}
	return ts.buffer[len(ts.buffer)-1]
}

func (ts *TokenStream) Emit() Token {
	for ts.PeekTokenAt(0).Type == COMMENT {
		commentTok := ts.EmitRaw()
		ts.leadingTrivia = append(ts.leadingTrivia, commentTok)
	}
	tok := ts.EmitRaw()
	if len(ts.leadingTrivia) > 0 && tok.Type != NEWLINE {
		tok.LeadingTrivia = ts.leadingTrivia
		ts.leadingTrivia = nil
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

func (ts *TokenStream) ConsumeTextBlock(endSequence []Token) (string, error) {
	if len(endSequence) == 0 {
		panic("empty end sequence for raw mode")
	}
	tok := ts.EmitRaw()
	startOff := tok.Pos.Offset
	for !ts.AssertSeq(endSequence) {
		tok = ts.EmitRaw()
		if tok.Type == EOF {
			return "", ErrUnexpectedEOF
		}
	}
	endOff := tok.Pos.Offset + uint(len(tok.Literal))
	for range endSequence {
		ts.EmitRaw() // consume end markers
	}

	return string(ts.lexer.input[startOff:endOff]), nil
}

// Assert checks if the next token matches the given target (Type and Literal if literal is set).
// Does not emit the token.
func (ts *TokenStream) Assert(target Token) bool {
	next := ts.PeekTokenAt(0)
	if next.Type != target.Type {
		return false
	}
	if target.Literal != "" && next.Literal != target.Literal {
		return false
	}
	return true
}

func (ts *TokenStream) AssertType(targetType TokenType) bool {
	return ts.PeekTokenAt(0).Type == targetType
}

func (ts *TokenStream) AssertAny(targets ...Token) bool {
	return slices.ContainsFunc(targets, ts.Assert)
}

func (ts *TokenStream) AssertAnyType(targetTypes ...TokenType) bool {
	return slices.ContainsFunc(targetTypes, ts.AssertType)
}

func (ts *TokenStream) AssertKW(kw keyword.KeywordKind) bool {
	return keyword.Classify(ts.PeekTokenAt(0).Literal) == kw
}

// MustConsumeType is ONLY for internal use as it panics
func (ts *TokenStream) MustConsumeType(targetType TokenType) Token {
	if !ts.AssertType(targetType) {
		panic(fmt.Sprintf("expected %s, got %s", targetType, ts.PeekTokenAt(0).Type))
	}
	return ts.Emit()
}

func (ts *TokenStream) TryConsume(target Token) (Token, bool) {
	if ts.Assert(target) {
		return ts.Emit(), true
	}
	return Token{}, false
}

func (ts *TokenStream) TryConsumeType(targetType TokenType) (Token, bool) {
	if ts.AssertType(targetType) {
		return ts.Emit(), true
	}
	return Token{}, false
}

func (ts *TokenStream) TryConsumeKW(kw keyword.KeywordKind) (Token, bool) {
	if ts.AssertKW(kw) {
		return ts.Emit(), true
	}
	return Token{}, false
}

// ConsumeUntil consumes tokens until it encounters one of the specified target tokens or EOF.
// Returns the slice of consumed tokens.
// NOTE: The terminator token itself is NOT consumed and NOT included in the returned slice;
// it remains at the head of the stream for callers to assert (e.g. via ts.Assert / ts.AssertType).
func (ts *TokenStream) ConsumeUntil(targets ...Token) []Token {
	var consumed []Token
	for !ts.AssertAny(targets...) && !ts.AssertType(EOF) {
		consumed = append(consumed, ts.Emit())
	}
	return consumed
}

// ConsumeUntilType consumes tokens until it encounters one of the specified token types or EOF.
// Returns the slice of consumed tokens.
// NOTE: The terminator token itself is NOT consumed and NOT included in the returned slice;
// it remains at the head of the stream for callers to assert (e.g. via ts.Assert / ts.AssertType).
func (ts *TokenStream) ConsumeUntilType(targetTypes ...TokenType) []Token {
	var consumed []Token
	for !ts.AssertAnyType(targetTypes...) && !ts.AssertType(EOF) {
		consumed = append(consumed, ts.Emit())
	}
	return consumed
}

var (
	ErrStartMarkerNotFound = errors.New("start marker not found")
	ErrUnexpectedEOF       = errors.New("unexpected EOF")
)

// ReadBetween handles the common pattern of [start markers]...[end markers]
func (ts *TokenStream) ReadBetween(start, end []Token) (string, error) {
	// 1. Check if start markers match
	if !ts.AssertSeq(start) {
		return "", ErrStartMarkerNotFound
	}

	// 2. Consume start markers
	for range start {
		ts.Emit()
	}

	// 3. Read content until end markers
	var collected []Token
	for {
		if ts.AssertType(EOF) || ts.AssertType(NEWLINE) {
			return "", ErrUnexpectedEOF
		}

		if ts.AssertSeq(end) {
			break
		}

		collected = append(collected, ts.Emit())
	}

	res := ts.TokensToString(collected)

	// Consume end markers
	for range end {
		ts.Emit()
	}

	return res, nil
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
