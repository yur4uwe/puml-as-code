package formatter

import (
	"testing"

	"yur4uwe/pac/internal/fixtures"
)

func BenchmarkFormat(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		out, err := Format(fixtures.ComprehensivePUML)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}

func BenchmarkFormat_Tabs(b *testing.B) {
	opts := FormatOptions{
		UseTabs:    true,
		IndentSize: 1,
	}
	b.ReportAllocs()
	for b.Loop() {
		out, err := Format(fixtures.ComprehensivePUML, opts)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}
