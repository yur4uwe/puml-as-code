package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"

	"github.com/stretchr/testify/require"
)

func TestParseIncludeDirectives(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ast.IncludeDirective
		expectErr bool
	}{
		{
			name:  "include simple path",
			input: "!include common.puml",
			want: ast.IncludeDirective{
				Kind: ast.IncludeOnce,
				Path: "common.puml",
			},
		},
		{
			name:  "include_once with path",
			input: "!include_once ./types/models.puml",
			want: ast.IncludeDirective{
				Kind: ast.IncludeOnce,
				Path: "./types/models.puml",
			},
		},
		{
			name:  "include_many with path",
			input: "!include_many base.puml",
			want: ast.IncludeDirective{
				Kind: ast.IncludeMany,
				Path: "base.puml",
			},
		},
		{
			name:  "include with diagram tag",
			input: "!include common.puml!diagram1",
			want: ast.IncludeDirective{
				Kind: ast.IncludeOnce,
				Path: "common.puml",
				Tag:  "diagram1",
			},
		},
		{
			name:  "include_many with diagram tag index",
			input: "!include_many common.puml!0",
			want: ast.IncludeDirective{
				Kind: ast.IncludeMany,
				Path: "common.puml",
				Tag:  "0",
			},
		},
		{
			name:  "incomplete include without path",
			input: "!include",
			want: ast.IncludeDirective{
				Kind: ast.IncludeOnce,
				Path: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(dialect.NewGoDialect())
			p.IsBoundless = true
			diagram, err := p.Parse(tt.input)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, diagram.Statements, 1)

			got, ok := diagram.Statements[0].(*ast.IncludeDirective)
			require.True(t, ok, "statement should be *ast.IncludeDirective")
			require.Equal(t, tt.want.Kind, got.Kind)
			require.Equal(t, tt.want.Path, got.Path)
			require.Equal(t, tt.want.Tag, got.Tag)
		})
	}
}

func TestParseDirectiveRecovery(t *testing.T) {
	t.Run("unknown include kind records diagnostic but produces IncludeDirective", func(t *testing.T) {
		input := "!include_unknown myfile.puml"
		p := NewParser(dialect.NewGoDialect())
		p.IsBoundless = true
		diagram, err := p.Parse(input)
		require.Error(t, err)
		require.NotNil(t, diagram)
		require.True(t, diagram.HasErrors())
		require.Len(t, diagram.Statements, 1)

		got, ok := diagram.Statements[0].(*ast.IncludeDirective)
		require.True(t, ok)
		require.Equal(t, ast.IncludeOnce, got.Kind)
		require.Equal(t, "myfile.puml", got.Path)
	})

	t.Run("bare exclamation mark records diagnostic and preserves surrounding statements", func(t *testing.T) {
		input := `@startuml
!
class User
@enduml`
		p := NewParser(dialect.NewGoDialect())
		diagram, err := p.Parse(input)
		require.Error(t, err)
		require.NotNil(t, diagram)
		require.True(t, diagram.HasErrors())
		require.Len(t, diagram.Statements, 4)

		uh, ok := diagram.Statements[1].(*ast.UnhandledStatement)
		require.True(t, ok)
		require.Equal(t, "!", uh.Raw)

		user, ok := diagram.Statements[2].(*ast.Entity)
		require.True(t, ok)
		require.Equal(t, "User", user.Identifier.EntityName())
	})

	t.Run("exclamation mark with space before directive name records diagnostic and preserves statement", func(t *testing.T) {
		input := `@startuml
! include file.puml
class Order
@enduml`
		p := NewParser(dialect.NewGoDialect())
		diagram, err := p.Parse(input)
		require.Error(t, err)
		require.NotNil(t, diagram)
		require.True(t, diagram.HasErrors())
		require.Len(t, diagram.Statements, 4)

		uh, ok := diagram.Statements[1].(*ast.UnhandledStatement)
		require.True(t, ok)
		require.Equal(t, "! include", uh.Raw)

		order, ok := diagram.Statements[2].(*ast.Entity)
		require.True(t, ok)
		require.Equal(t, "Order", order.Identifier.EntityName())
	})
}
