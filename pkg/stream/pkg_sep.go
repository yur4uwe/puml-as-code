package stream

import (
	"strings"

	"yur4uwe/pac/pkg/tokenizer"
)

func (ts *TokenStream) MatchPackageSeparatorAt(startIdx int) (int, bool) {
	if ts.PackageSeparator == "" {
		return 0, false
	}
	var sb strings.Builder
	var lastEnd uint
	for count := 1; ; count++ {
		tok := ts.PeekRawTokenAt(startIdx + count - 1)
		if tok.Type == tokenizer.EOF || tok.Type == tokenizer.NEWLINE {
			break
		}
		if count > 1 && tok.Span.Start.Offset != lastEnd {
			break
		}
		sb.WriteString(tok.Literal)
		lastEnd = tok.Span.End.Offset
		if sb.String() == ts.PackageSeparator {
			return count, true
		}
		if sb.Len() >= len(ts.PackageSeparator) {
			break
		}
	}
	return 0, false
}

func (ts *TokenStream) IsPackageSeparator(toks []tokenizer.Token) bool {
	if len(toks) == 0 || ts.PackageSeparator == "" {
		return false
	}
	var sb strings.Builder
	for i, tok := range toks {
		if i > 0 && tok.Span.Start.Offset != toks[i-1].Span.End.Offset {
			return false
		}
		sb.WriteString(tok.Literal)
		if sb.Len() > len(ts.PackageSeparator) {
			return false
		}
	}
	return sb.String() == ts.PackageSeparator
}

func (ts *TokenStream) TryConsumePackageSeparator() (string, bool) {
	count, ok := ts.MatchPackageSeparatorAt(0)
	if !ok {
		return "", false
	}
	for range count {
		ts.EmitRaw()
	}
	return ts.PackageSeparator, true
}
