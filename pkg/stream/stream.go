package stream

import (
	"errors"
	"fmt"
	"strings"

	"yur4uwe/pac/pkg/tokenizer"
)

const maxConsecutiveEOF = 25

type UnexpectedTokenError struct {
	Expected tokenizer.Token
	Found    tokenizer.Token
}

func (e UnexpectedTokenError) Error() string {
	return fmt.Sprintf("token stream: unexpected token %s(%q) at %d:%d, expected %s(%q)", e.Found.Type, e.Found.Literal, e.Found.Span.Start.Line, e.Found.Span.Start.Col, e.Expected.Type, e.Expected.Literal)
}

func unexpectedTokenError(expected tokenizer.Token, found tokenizer.Token) error {
	return UnexpectedTokenError{Expected: expected, Found: found}
}

type TokenStream struct {
	source              []rune
	lexer               *tokenizer.Lexer
	collectedTrivia     []tokenizer.Token
	buffer              []tokenizer.Token
	sinks               []TokenSink
	rawModeTerminator   []rune
	PackageSeparator    string
	lastNonNewlineToken tokenizer.Token
	eofEmitCount        int
}

func NewTokenStream(input string) *TokenStream {
	runes := []rune(input)
	return &TokenStream{
		source:           runes,
		lexer:            tokenizer.NewLexer(runes),
		PackageSeparator: ".",
		sinks:            make([]TokenSink, 0, 2),
	}
}

func (ts *TokenStream) DumpCollectedTrivia() []tokenizer.Token {
	defer func() {
		ts.collectedTrivia = nil
	}()
	return ts.collectedTrivia
}

// EmitCommentToks consumes any immediate tokenizer.COMMENT tokens in the stream,
// appending them to collected trivia.
func (ts *TokenStream) EmitCommentToks() {
	for ts.PeekRawTokenAt(0).Type == tokenizer.COMMENT {
		ts.collectedTrivia = append(ts.collectedTrivia, ts.EmitRaw())
	}
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

// TokensToString is DEPRECATED as it can distort the original source
// Use SliceInputEnclosingTokens instead
func (ts *TokenStream) TokensToString(toks []tokenizer.Token) string {
	if len(toks) == 0 {
		return ""
	}

	var sb strings.Builder
	var prevTok *tokenizer.Token
	for i := range toks {
		tok := &toks[i]
		if tok.Type == tokenizer.COMMENT {
			continue
		}

		if prevTok != nil {
			startNext := tok.Span.Start.Offset
			endPrev := prevTok.Span.End.Offset
			if startNext > endPrev {
				sb.WriteString(string(ts.source[endPrev:startNext]))
			}
		}

		sb.WriteString(tok.Literal)
		prevTok = tok
	}

	return sb.String()
}

var (
	ErrStartMarkerNotFound = errors.New("start marker not found")
	ErrUnexpectedEOF       = errors.New("unexpected EOF")
	ErrScopeDelimiterHit   = errors.New("hit enclosing scope delimiter")
)

// ReadTokensBetween handles bracketed sequences (e.g. <<...>>, {...}, <...>),
// returning the enclosed tokens and the full source span enclosing both start and end markers.
// Optional stopTokens immediately abort collection with ErrUnexpectedEOF before consuming the stop token.
func (ts *TokenStream) ReadTokensBetween(start, end []tokenizer.TokenType, stopTokens ...tokenizer.TokenType) ([]tokenizer.Token, tokenizer.SourceSpan, error) {
	// 1. Check if start markers match
	if !ts.AssertTypeSeq(start...) {
		return nil, tokenizer.SourceSpan{}, ErrStartMarkerNotFound
	}

	startTok := ts.PeekTokenAt(0)
	// 2. Consume start markers
	for range start {
		ts.Emit()
	}

	// 3. Read content until end markers
	var collected []tokenizer.Token
	for {
		if ts.AssertType(tokenizer.EOF) || ts.AssertType(tokenizer.NEWLINE) || (len(stopTokens) > 0 && ts.AssertAnyType(stopTokens...)) {
			return nil, tokenizer.SourceSpan{}, ErrUnexpectedEOF
		}

		if ts.AssertTypeSeq(end...) {
			break
		}

		collected = append(collected, ts.Emit())
	}

	// 4. Consume end markers
	var lastEndTok tokenizer.Token
	for range end {
		lastEndTok = ts.Emit()
	}

	span := tokenizer.SourceSpan{
		Start: startTok.Span.Start,
		End:   lastEndTok.Span.End,
	}

	return collected, span, nil
}

// ReadBlock assumes that end tokens are first in the line
// So actual end markers are implicitly append([]tokenizer.Token{tokenizer.NEWLINE}, ...endMark)
func (ts *TokenStream) ReadBlock(endMark ...tokenizer.Token) (string, error) {
	tok := ts.EmitRaw()
	startPos := tok.Span.Start
	for tok.Type != tokenizer.EOF {
		if tok.Type == tokenizer.NEWLINE && ts.AssertSeq(endMark) {
			endPos := tok.Span.Start
			for range endMark {
				ts.EmitRaw() // consume end markers
			}
			return string(ts.source[startPos.Offset:endPos.Offset]), nil
		}
		tok = ts.EmitRaw()
	}
	return "", fmt.Errorf("block end not found")
}

func (ts *TokenStream) LastSemanticToken() tokenizer.Token {
	return ts.lastNonNewlineToken
}
