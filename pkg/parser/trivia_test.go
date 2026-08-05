package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
)

func TestCommentTriviaAttachment(t *testing.T) {
	input := `@startuml
/' Leading comment for class '/
class Foo {
}

' Single line comment for relationship
Foo --> Bar
@enduml`

	p := &Parser{
		dialect: dialect.NewGoDialect(),
	}
	diag, err := p.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, diag)

	var fooEntity *ast.Entity
	var relStmt *ast.Relationship

	for _, stmt := range diag.Statements {
		switch s := stmt.(type) {
		case *ast.Entity:
			if s.Identifier == "Foo" {
				fooEntity = s
			}
		case ast.Entity:
			if s.Identifier == "Foo" {
				fooEntity = &s
			}
		case *ast.Relationship:
			relStmt = s
		case ast.Relationship:
			relStmt = &s
		}
	}

	require.NotNil(t, fooEntity, "Expected Foo entity to be present")
	require.NotEmpty(t, fooEntity.LeadingTrivia, "Expected leading trivia on Foo entity")
	require.Contains(t, fooEntity.LeadingTrivia[0].Literal, "Leading comment for class")

	require.NotNil(t, relStmt, "Expected relationship statement to be present")
	require.NotEmpty(t, relStmt.LeadingTrivia, "Expected leading trivia on relationship statement")
	require.Contains(t, relStmt.LeadingTrivia[0].Literal, "Single line comment for relationship")
}
