package tokenizer

import (
	"testing"

	"yur4uwe/pac/internal/fixtures"
)

func BenchmarkLexer(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		l := NewLexer(fixtures.ComprehensivePUML)
		for {
			tok := l.Emit()
			if tok.Type == EOF {
				break
			}
		}
	}
}
