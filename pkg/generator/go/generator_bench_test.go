package gogenerator

import (
	"testing"

	"yur4uwe/pac/internal/fixtures"
	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"
)

func parseTableForBench(b *testing.B, input string) *resolver.SymbolTable {
	b.Helper()
	p := parser.NewParser(dialect.NewGoDialect())
	diag, err := p.Parse(input)
	if err != nil {
		b.Fatal(err)
	}
	tbl := resolver.ResolveSymbols(diag)
	if diag.HasErrors() {
		b.Fatal("diagram has errors")
	}
	return tbl
}

func BenchmarkGoGenerator_GenerateFromClassDiagram(b *testing.B) {
	tbl := parseTableForBench(b, fixtures.ComprehensivePUML)
	gen := GoCodeGenerator{}
	if err := gen.SemanticPass(tbl); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		files, err := gen.GenerateFromClassDiagram(tbl)
		if err != nil {
			b.Fatal(err)
		}
		_ = files
	}
}

func BenchmarkGoGenerator_SemanticPass(b *testing.B) {
	tbl := parseTableForBench(b, fixtures.ComprehensivePUML)
	gen := GoCodeGenerator{}

	b.ReportAllocs()
	for b.Loop() {
		if err := gen.SemanticPass(tbl); err != nil {
			b.Fatal(err)
		}
	}
}
