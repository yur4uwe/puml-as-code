package stream

import (
	"errors"

	"yur4uwe/pac/pkg/tokenizer"
)

const maxConsecutiveEOF = 25

type TokenStream struct {
	source              []rune
	lexer               *tokenizer.Lexer
	collectedTrivia     []tokenizer.Token
	buffer              []tokenizer.Token
	bufferHead          int
	sinks               []TokenSink
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

var (
	ErrStartMarkerNotFound = errors.New("start marker not found")
	ErrUnexpectedEOF       = errors.New("unexpected EOF")
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

func (ts *TokenStream) LastSemanticToken() tokenizer.Token {
	return ts.lastNonNewlineToken
}
