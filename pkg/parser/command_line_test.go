package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"

	"github.com/stretchr/testify/require"
)

func TestLineCommandsRecoverWithoutConsumingNextLine(t *testing.T) {
	diagram, err := NewParser(dialect.LaxDialect{}).Parse(`@startuml
set
set separator
hide ' still a trailing comment
left to bottom direction
set separator ::
class domain::User
top to bottom direction
@enduml`)
	require.Error(t, err)
	require.Len(t, diagram.Diagnostics, 4)
	require.Len(t, diagram.Statements, 9)
	require.IsType(t, &ast.SetCommand{}, diagram.Statements[1])
	require.IsType(t, &ast.SetCommand{}, diagram.Statements[2])
	hide := diagram.Statements[3].(*ast.VisibilityCommand)
	require.Empty(t, hide.Target)
	require.Contains(t, hide.TrailingTrivia[0].Text, "trailing comment")
	require.IsType(t, &ast.DirectionCommand{}, diagram.Statements[4])
	require.Equal(t, "::", diagram.Statements[5].(*ast.SetCommand).Value)
	entity := diagram.Statements[6].(*ast.Entity)
	require.Equal(t, []string{"domain"}, entity.Identifier.PackagePathStrings())
	require.Equal(t, ast.TopToBottomDirection, diagram.Statements[7].(*ast.DirectionCommand).Direction)
}
