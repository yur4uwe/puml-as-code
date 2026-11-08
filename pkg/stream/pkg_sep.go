package stream

import (
	"strings"

	"yur4uwe/pac/pkg/tokenizer"
)

func (ts *TokenStream) MatchPackageSeparatorAt(startIdx int) (int, bool) {
	if ts.PackageSeparator == "" {
		return 0, false
	}
	var acc []tokenizer.Token
	for {
		accCount := len(acc)
		tok := ts.PeekRawTokenAt(startIdx + accCount)
		if tok.Type == tokenizer.EOF || tok.Type == tokenizer.NEWLINE {
			break
		}
		if ts.IsPackageSeparator(acc) {
			return accCount, true
		}
		acc = append(acc, tok)
	}
	return 0, false
}

func (ts *TokenStream) IsPackageSeparator(toks []tokenizer.Token) bool {
	var sb strings.Builder
	for i, tok := range toks {
		if i > 0 && tok.Span.Start.Offset != toks[i-1].Span.End.Offset {
			return false
		}
		sb.WriteString(tok.Literal)
		if sb.String() == ts.PackageSeparator {
			return true
		}
		if len(sb.String()) >= len(ts.PackageSeparator) {
			return false
		}
	}
	return false
}

func (ts *TokenStream) TryConsumePackageSeparator() (string, bool) {
	count, ok := ts.MatchPackageSeparatorAt(0)
	if !ok {
		return "", false
	}
	for range count {
		ts.Emit()
	}
	return ts.PackageSeparator, true
}
