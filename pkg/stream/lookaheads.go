package stream

import (
	"errors"
	"strings"

	"yur4uwe/pac/pkg/tokenizer"
)

// ConsumeTextBlock consumes all tokens until a matching closer line is found.
// It handles single-token ("endheader") and multi-token ("end header") closers case-insensitively,
// preserves leading indentation on the body lines, strips indentation before the closer line,
// prevents false-positive matches mid-line
func (ts *TokenStream) ConsumeTextBlock(delimiterLiterals ...string) (string, error) {
	if len(delimiterLiterals) == 0 {
		return "", errors.New("empty closers for ConsumeTextBlock")
	}

	delim := strings.ToLower(strings.Join(delimiterLiterals, ""))

	var startBodyOffset uint
	if ts.AssertType(tokenizer.NEWLINE) {
		nlTok := ts.Emit()
		startBodyOffset = nlTok.Span.End.Offset
	} else if !ts.AssertType(tokenizer.EOF) {
		startBodyOffset = ts.PeekRawTokenAt(0).Span.Start.Offset
	}

	var closerStartOffset uint

	for {
		if ts.AssertType(tokenizer.EOF) {
			return "", ErrUnexpectedEOF
		}

		// Collect tokens on the current line
		curLineToks := ts.ConsumeUntilType(tokenizer.NEWLINE)
		if len(curLineToks) <= 0 {
			ts.TryConsumeType(tokenizer.NEWLINE)
			continue
		}

		if tokenizer.MatchTokenLine(curLineToks, delim) {
			closerStartOffset = curLineToks[0].Span.Start.Offset
			ts.TryConsumeType(tokenizer.NEWLINE)
			break
		} else if tokenizer.MatchTokenLine(curLineToks, "@enduml") {
			return "", ErrUnexpectedEOF
		}
	}

	rawBody := ts.SliceInput(startBodyOffset, closerStartOffset)
	if lastNL := strings.LastIndex(rawBody, "\n"); lastNL != -1 {
		rawBody = rawBody[:lastNL]
		rawBody = strings.TrimSuffix(rawBody, "\r")
	} else {
		rawBody = ""
	}

	return rawBody, nil
}

// HasCloserBeforeBoundary peeks forward in the stream without consuming tokens to check
// whether a line matching delimiterLiterals exists before encountering tokenizer.EOF or @enduml.
func (ts *TokenStream) HasCloserBeforeBoundary(delimiterLiterals ...string) bool {
	delim := strings.ToLower(strings.Join(delimiterLiterals, ""))
	var curLine []tokenizer.Token
	for i := 0; ; i++ {
		tok := ts.PeekTokenAt(i)
		if tok.Type == tokenizer.EOF {
			break
		}
		if tok.Type == tokenizer.NEWLINE {
			if len(curLine) > 0 {
				if tokenizer.MatchTokenLine(curLine, delim) {
					return true
				}
				if tokenizer.MatchTokenLine(curLine, "@enduml") {
					return false
				}
			}
			curLine = curLine[:0]
			continue
		}
		curLine = append(curLine, tok)
	}
	if len(curLine) > 0 && tokenizer.MatchTokenLine(curLine, delim) {
		return true
	}
	return false
}

// ConsumeUntil consumes tokens until it encounters one of the specified target tokens or tokenizer.EOF.
// Returns the slice of consumed tokens.
// NOTE: The terminator token itself is NOT consumed and NOT included in the returned slice;
// it remains at the head of the stream for callers to assert (e.g. via ts.Assert / ts.AssertType).
func (ts *TokenStream) ConsumeUntil(targets ...tokenizer.Token) []tokenizer.Token {
	var consumed []tokenizer.Token
	for !ts.AssertAny(targets...) && !ts.AssertType(tokenizer.EOF) {
		consumed = append(consumed, ts.Emit())
	}
	return consumed
}

// ConsumeUntilType consumes tokens until it encounters one of the specified token types or tokenizer.EOF.
// Returns the slice of consumed tokens.
// NOTE: The terminator token itself is NOT consumed and NOT included in the returned slice;
// it remains at the head of the stream for callers to assert (e.g. via ts.Assert / ts.AssertType).
func (ts *TokenStream) ConsumeUntilType(targetTypes ...tokenizer.TokenType) []tokenizer.Token {
	var consumed []tokenizer.Token
	for !ts.AssertAnyType(targetTypes...) && !ts.AssertType(tokenizer.EOF) {
		consumed = append(consumed, ts.Emit())
	}
	return consumed
}
