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
		Dialect: dialect.NewGoDialect(),
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

func TestMemberAndTrailingTriviaAttachment(t *testing.T) {
	input := `@startuml
class Example {
  ' Field leading comment
  +id string ' Field trailing comment
  -- Methods --
  ' Method leading comment
  +DoWork() error ' Method trailing comment
} ' Class closing trailing comment
@enduml`

	p := &Parser{
		Dialect: dialect.NewGoDialect(),
	}
	diag, err := p.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, diag)

	var exampleEntity *ast.Entity
	for _, stmt := range diag.Statements {
		if ent, ok := stmt.(*ast.Entity); ok && ent.Identifier == "Example" {
			exampleEntity = ent
			break
		}
	}
	require.NotNil(t, exampleEntity, "Expected Example entity")
	require.NotEmpty(t, exampleEntity.TrailingTrivia, "Expected trailing trivia on Example entity closing")
	require.Contains(t, exampleEntity.TrailingTrivia[0].Literal, "Class closing trailing comment")

	require.Len(t, exampleEntity.Members, 3, "Expected 3 members: field, separator, method")

	// 1. Field
	field, ok := exampleEntity.Members[0].(*dialect.GoField)
	require.True(t, ok, "Expected GoField")
	require.NotEmpty(t, field.LeadingTrivia, "Expected leading trivia on field")
	require.Contains(t, field.LeadingTrivia[0].Literal, "Field leading comment")
	require.NotEmpty(t, field.TrailingTrivia, "Expected trailing trivia on field")
	require.Contains(t, field.TrailingTrivia[0].Literal, "Field trailing comment")

	// 2. Separator
	sep, ok := exampleEntity.Members[1].(ast.ClassSeparator)
	require.True(t, ok, "Expected ClassSeparator")
	require.Equal(t, "Methods", sep.Label)

	// 3. Method
	method, ok := exampleEntity.Members[2].(*dialect.GoMethod)
	require.True(t, ok, "Expected GoMethod")
	require.NotEmpty(t, method.LeadingTrivia, "Expected leading trivia on method")
	require.Contains(t, method.LeadingTrivia[0].Literal, "Method leading comment")
	require.NotEmpty(t, method.TrailingTrivia, "Expected trailing trivia on method")
	require.Contains(t, method.TrailingTrivia[0].Literal, "Method trailing comment")
}
