package stream

import (
	"fmt"
	"slices"

	"yur4uwe/pac/pkg/tokenizer/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

// PeekRawTokenAt returns the idx-th token in the stream.
func (ts *TokenStream) PeekRawTokenAt(idx int) tokenizer.Token {
	for len(ts.buffer) <= idx {
		tok := ts.lexer.Emit()
		ts.buffer = append(ts.buffer, tok)
		if tok.Type != tokenizer.EOF {
			ts.eofEmitCount = 0
			continue
		}

		ts.eofEmitCount++
		if ts.eofEmitCount > maxConsecutiveEOF {
			panic(fmt.Sprintf(
				"tokenizer: possible infinite loop detected: emitted tokenizer.EOF %d times (at %s)",
				ts.eofEmitCount, ts.lexer.Pos(),
			))
		}
		break
	}
	if idx < len(ts.buffer) {
		return ts.buffer[idx]
	}
	return ts.buffer[len(ts.buffer)-1]
}

// PeekTokenAt returns the idx-th NON-tokenizer.COMMENT token in the stream.
func (ts *TokenStream) PeekTokenAt(idx int) tokenizer.Token {
	nonCommentCount := 0
	for rawIdx := 0; ; rawIdx++ {
		tok := ts.PeekRawTokenAt(rawIdx)
		if tok.Type != tokenizer.COMMENT {
			if nonCommentCount == idx {
				return tok
			}
			nonCommentCount++
		}
		if tok.Type == tokenizer.EOF {
			return tok
		}
	}
}

// Emit consumes the next NON-COMMENT token and returns it.
func (ts *TokenStream) Emit() tokenizer.Token {
	for ts.PeekRawTokenAt(0).Type == tokenizer.COMMENT {
		commentTok := ts.EmitRaw()
		ts.collectedTrivia = append(ts.collectedTrivia, commentTok)
	}
	tok := ts.EmitRaw()
	return tok
}

// EmitRaw consumes the next token and returns it.
func (ts *TokenStream) EmitRaw() tokenizer.Token {
	tok := ts.PeekRawTokenAt(0)
	if len(ts.buffer) > 0 {
		ts.buffer = ts.buffer[1:]
	}
	for _, sink := range ts.sinks {
		sink.Receive(tok)
	}
	if tok.Type != tokenizer.EOF && tok.Type != tokenizer.NEWLINE && tok.Type != tokenizer.COMMENT {
		ts.lastNonNewlineToken = tok
	}
	return tok
}

// Assert checks if the next token matches the given target (Type and Literal if literal is set).
// Does not emit the token.
func (ts *TokenStream) Assert(target tokenizer.Token) bool {
	next := ts.PeekTokenAt(0)
	if next.Type != target.Type {
		return false
	}
	if target.Literal != "" && next.Literal != target.Literal {
		return false
	}
	return true
}

func (ts *TokenStream) AssertType(targetType tokenizer.TokenType) bool {
	return ts.PeekTokenAt(0).Type == targetType
}

func (ts *TokenStream) AssertAny(targets ...tokenizer.Token) bool {
	return slices.ContainsFunc(targets, ts.Assert)
}

func (ts *TokenStream) AssertAnyType(targetTypes ...tokenizer.TokenType) bool {
	return slices.ContainsFunc(targetTypes, ts.AssertType)
}

func (ts *TokenStream) AssertKW(kw keyword.KeywordKind) bool {
	return keyword.Classify(ts.PeekTokenAt(0).Literal) == kw
}

func (ts *TokenStream) AssertSeq(seq []tokenizer.Token) bool {
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

func (ts *TokenStream) AssertTypeSeq(seq ...tokenizer.TokenType) bool {
	for i, t := range seq {
		next := ts.PeekTokenAt(i)
		if next.Type != t {
			return false
		}
	}
	return true
}

func (ts *TokenStream) TryConsume(target tokenizer.Token) (tokenizer.Token, bool) {
	if ts.Assert(target) {
		return ts.Emit(), true
	}
	return tokenizer.Token{}, false
}

func (ts *TokenStream) TryConsumeType(targetType tokenizer.TokenType) (tokenizer.Token, bool) {
	if ts.AssertType(targetType) {
		return ts.Emit(), true
	}
	return tokenizer.Token{}, false
}

func (ts *TokenStream) TryConsumeKW(kw keyword.KeywordKind) (tokenizer.Token, bool) {
	if ts.AssertKW(kw) {
		return ts.Emit(), true
	}
	return tokenizer.Token{}, false
}
