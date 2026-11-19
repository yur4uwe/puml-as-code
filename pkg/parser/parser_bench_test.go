package parser

import (
	"testing"

	"yur4uwe/pac/internal/fixtures"
	"yur4uwe/pac/pkg/parser/dialect"
)

func BenchmarkParser_GoDialect(b *testing.B) {
	d := dialect.NewGoDialect()
	b.ReportAllocs()
	for b.Loop() {
		p := NewParser(d)
		_, err := p.Parse(fixtures.ComprehensivePUML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParser_LaxDialect(b *testing.B) {
	d := dialect.LaxDialect{}
	b.ReportAllocs()
	for b.Loop() {
		p := NewParser(d)
		_, err := p.Parse(fixtures.ComprehensivePUML)
		if err != nil {
			b.Fatal(err)
		}
	}
}
