package stream

import (
	"testing"

	"yur4uwe/pac/internal/fixtures"
	"yur4uwe/pac/pkg/tokenizer"
)

func BenchmarkTokenStream_Emit(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		ts := NewTokenStream(fixtures.ComprehensivePUML)
		for {
			tok := ts.Emit()
			if tok.Type == tokenizer.EOF {
				break
			}
		}
	}
}

func BenchmarkTokenStream_Peek(b *testing.B) {
	ts := NewTokenStream(fixtures.ComprehensivePUML)
	b.ReportAllocs()
	for b.Loop() {
		_ = ts.PeekTokenAt(10)
	}
}
