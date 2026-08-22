package resolver

import (
	"testing"

	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"

	"github.com/stretchr/testify/require"
)

func parseAndResolve(t *testing.T, mainFile string, files MapFS) *ast.Diagram {
	t.Helper()

	content, ok := files[mainFile]
	require.True(t, ok, "main file not found in test MapFS: %s", mainFile)

	p := parser.NewParser(dialect.NewGoDialect())
	diag, err := p.Parse(string(content))
	require.NoError(t, err, "failed to parse main diagram")

	err = ResolveImports(diag, mainFile, "", files, "go")
	require.NoError(t, err, "failed to resolve imports")

	return diag
}

// extractStatementSummary inspects diagram.Statements in order, ensures no unresolved
// IncludeDirectives remain, verifies outer boundary markers, and returns an ordered
// summary of statements for exact AST structural assertions.
func extractStatementSummary(t *testing.T, diagram *ast.Diagram) []string {
	t.Helper()

	require.NotEmpty(t, diagram.Statements, "diagram statements should not be empty")

	// First and last statement must be diagram bounds
	startBound, ok := diagram.Statements[0].(ast.DiagramBound)
	require.True(t, ok && startBound.IsStart, "first statement must be @startuml DiagramBound")

	endBound, ok := diagram.Statements[len(diagram.Statements)-1].(ast.DiagramBound)
	require.True(t, ok && !endBound.IsStart, "last statement must be @enduml DiagramBound")

	var summary []string
	for _, stmt := range diagram.Statements[1 : len(diagram.Statements)-1] {
		switch s := stmt.(type) {
		case *ast.Entity:
			summary = append(summary, s.Kind.String()+":"+s.Identifier)
		case ast.Relationship:
			summary = append(summary, "rel:"+s.LHS.Entity+"->"+s.RHS.Entity)
		case ast.IncludeDirective:
			t.Fatalf("unexpected un-spliced IncludeDirective remaining in AST: %+v", stmt)
		case ast.DiagramBound:
			t.Fatalf("unexpected inner DiagramBound in spliced statement list: %+v", stmt)
		default:
			summary = append(summary, "stmt")
		}
	}
	return summary
}

func TestResolveImports_BasicInclude(t *testing.T) {
	fs := MapFS{
		"/project/main.puml": []byte(`
@startuml
!include models.puml
class MainService
MainService --> User
@enduml
`),
		"/project/models.puml": []byte(`
@startuml
class User
class Order
User --> Order
@enduml
`),
	}

	diag := parseAndResolve(t, "/project/main.puml", fs)
	expected := []string{
		"class:User",
		"class:Order",
		"rel:User->Order",
		"class:MainService",
		"rel:MainService->User",
	}
	require.Equal(t, expected, extractStatementSummary(t, diag))
}

func TestResolveImports_DiamondIncludeOnce(t *testing.T) {
	// Root includes A and B. Both A and B include Common via !include (default include_once).
	// Common.puml defines CommonModel. It must appear only ONCE in the spliced AST.
	fs := MapFS{
		"/project/root.puml": []byte(`
@startuml
!include a.puml
!include b.puml
class RootApp
@enduml
`),
		"/project/a.puml": []byte(`
@startuml
!include common.puml
class ServiceA
@enduml
`),
		"/project/b.puml": []byte(`
@startuml
!include_once common.puml
class ServiceB
@enduml
`),
		"/project/common.puml": []byte(`
@startuml
class CommonModel
@enduml
`),
	}

	diag := parseAndResolve(t, "/project/root.puml", fs)
	expected := []string{
		"class:CommonModel",
		"class:ServiceA",
		"class:ServiceB",
		"class:RootApp",
	}
	require.Equal(t, expected, extractStatementSummary(t, diag))
}

func TestResolveImports_IncludeMany(t *testing.T) {
	// Using !include_many allows repeated inclusion of snippet.puml
	fs := MapFS{
		"/project/main.puml": []byte(`
@startuml
!include_many snippet.puml
!include_many snippet.puml
class Main
@enduml
`),
		"/project/snippet.puml": []byte(`
@startuml
class ReusedItem
@enduml
`),
	}

	diag := parseAndResolve(t, "/project/main.puml", fs)
	expected := []string{
		"class:ReusedItem",
		"class:ReusedItem",
		"class:Main",
	}
	require.Equal(t, expected, extractStatementSummary(t, diag))
}

func TestResolveImports_CircularCycleError(t *testing.T) {
	fs := MapFS{
		"/project/a.puml": []byte(`
@startuml
!include b.puml
class A
@enduml
`),
		"/project/b.puml": []byte(`
@startuml
!include a.puml
class B
@enduml
`),
	}

	p := parser.NewParser(dialect.NewGoDialect())
	diag, err := p.Parse(string(fs["/project/a.puml"]))
	require.NoError(t, err)

	err = ResolveImports(diag, "/project/a.puml", "", fs, "go")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrIncludeCycle)
}

func TestResolveImports_TaggedBlock(t *testing.T) {
	fs := MapFS{
		"/project/main.puml": []byte(`
@startuml
!include lib.puml!AUTH
class App
@enduml
`),
		"/project/lib.puml": []byte(`
@startuml(id=CORE)
class CoreEntity
@enduml

@startuml(id=AUTH)
class AuthToken
@enduml
`),
	}

	diag := parseAndResolve(t, "/project/main.puml", fs)
	expected := []string{
		"class:AuthToken",
		"class:App",
	}
	require.Equal(t, expected, extractStatementSummary(t, diag))
}

func TestResolveImports_StdlibUnimplemented(t *testing.T) {
	fs := MapFS{
		"/project/main.puml": []byte(`
@startuml
!include <C4/C4_Container>
class App
@enduml
`),
	}

	p := parser.NewParser(dialect.NewGoDialect())
	diag, err := p.Parse(string(fs["/project/main.puml"]))
	require.NoError(t, err)

	err = ResolveImports(diag, "/project/main.puml", "", fs, "go")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrStdlibNotImplemented)
}

func TestResolveImports_RemoteIncludeUnimplemented(t *testing.T) {
	fs := MapFS{
		"/project/main.puml": []byte(`
@startuml
!include https://raw.githubusercontent.com/plantuml/plantuml/master/test.puml
class App
@enduml
`),
	}

	p := parser.NewParser(dialect.NewGoDialect())
	diag, err := p.Parse(string(fs["/project/main.puml"]))
	require.NoError(t, err)

	err = ResolveImports(diag, "/project/main.puml", "", fs, "go")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrRemoteIncludeUnimplemented)
}
