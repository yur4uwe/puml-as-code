package resolver

import (
	"testing"

	"yur4uwe/pac/internal/fixtures"
	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/dialect"
)

func BenchmarkResolveSymbols(b *testing.B) {
	p := parser.NewParser(dialect.NewGoDialect())
	diag, err := p.Parse(fixtures.ComprehensivePUML)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		tbl := ResolveSymbols(diag)
		_ = tbl
	}
}

func BenchmarkResolveImports(b *testing.B) {
	fs := MapFS{
		"/project/main.puml": []byte(`@startuml
!include models.puml
class Service
Service --> User
@enduml`),
		"/project/models.puml": []byte(`class User {
  +id string
}
`),
	}

	b.ReportAllocs()
	for b.Loop() {
		p := parser.NewParser(dialect.NewGoDialect())
		diag, err := p.Parse(string(fs["/project/main.puml"]))
		if err != nil {
			b.Fatal(err)
		}
		if err := ResolveImports(diag, "/project/main.puml", "", fs, "go"); err != nil {
			b.Fatal(err)
		}
	}
}
