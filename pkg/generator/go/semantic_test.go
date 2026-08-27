package gogenerator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"
)

func parseAndResolveTable(t *testing.T, input string) *resolver.SymbolTable {
	t.Helper()
	p := parser.NewParser(dialect.NewGoDialect())
	diag, err := p.Parse(input)
	require.NoError(t, err)

	tbl, err := resolver.ResolveSymbols(diag)
	require.NoError(t, err)

	return tbl
}

func TestSemanticPass_PromoteViaRealization(t *testing.T) {
	input := `
@startuml
class Reader {
    +Read() error
}
class File {
    +Read() error
}
File ..|> Reader
@enduml
`
	tbl := parseAndResolveTable(t, input)
	gen := GoCodeGenerator{}
	err := gen.SemanticPass(tbl)
	require.NoError(t, err)

	reader := tbl.Lookup("Reader")
	require.NotNil(t, reader)
	require.Equal(t, ast.EntityInterface, reader.AST.Kind, "Reader should be promoted to interface")
}

func TestSemanticPass_PromoteViaInterfaceInheritance(t *testing.T) {
	input := `
@startuml
class Reader {
    +Read() error
}
interface ReadSeeker {
    +Seek() error
}
ReadSeeker --|> Reader
@enduml
`
	tbl := parseAndResolveTable(t, input)
	gen := GoCodeGenerator{}
	err := gen.SemanticPass(tbl)
	require.NoError(t, err)

	reader := tbl.Lookup("Reader")
	require.NotNil(t, reader)
	require.Equal(t, ast.EntityInterface, reader.AST.Kind, "Reader should be promoted to interface")
}

func TestSemanticPass_ErrorIfRealizedClassHasFields(t *testing.T) {
	input := `
@startuml
class Reader {
    +buffer string
    +Read() error
}
class File
File ..|> Reader
@enduml
`
	tbl := parseAndResolveTable(t, input)
	gen := GoCodeGenerator{}
	err := gen.SemanticPass(tbl)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot realize class Reader as an interface in Go: it contains fields")
}

func TestSemanticPass_ErrorIfInterfaceDeclaresFields(t *testing.T) {
	input := `
@startuml
interface InvalidInterface {
    +buffer string
    +Read() error
}
@enduml
`
	tbl := parseAndResolveTable(t, input)
	gen := GoCodeGenerator{}
	err := gen.SemanticPass(tbl)
	require.Error(t, err)
	require.Contains(t, err.Error(), "interface or abstract class InvalidInterface cannot declare fields")
}

func TestSemanticPass_EnumMustDeclareFields(t *testing.T) {
	input := `
@startuml
enum EmptyEnum {
}
@enduml
`
	tbl := parseAndResolveTable(t, input)
	gen := GoCodeGenerator{}
	err := gen.SemanticPass(tbl)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enum EmptyEnum must declare fields")
}

func TestSemanticPass_ImplicitEntityNoPanic(t *testing.T) {
	input := `
@startuml
User --> Order
@enduml
`
	tbl := parseAndResolveTable(t, input)
	gen := GoCodeGenerator{}
	// Should not panic on implicit entities where ent.AST == nil
	require.NotPanics(t, func() {
		_ = gen.SemanticPass(tbl)
	})
}
