package stream

import "yur4uwe/pac/pkg/tokenizer"

func (ts *TokenStream) SliceInput(start, end uint) string {
	if int(start) > len(ts.source) {
		return ""
	}
	if int(end) > len(ts.source) {
		end = uint(len(ts.source))
	}
	if start > end {
		return ""
	}
	return string(ts.source[start:end])
}

func (ts *TokenStream) SliceInputEnclosingTokens(toks ...tokenizer.Token) string {
	span, ok := tokenizer.TokenSliceSpan(toks)
	if !ok {
		return ""
	}
	return ts.SliceInput(span.Start.Offset, span.End.Offset)
}

func (ts *TokenStream) ReadUntilNewline() string {
	toks := ts.ConsumeUntilType(tokenizer.NEWLINE)
	if len(toks) == 0 {
		return ""
	}

	return ts.SliceInputEnclosingTokens(toks...)
}
