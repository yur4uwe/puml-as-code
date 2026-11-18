package parser

import (
	"testing"

	"yur4uwe/pac/internal/testutil"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

func TestParseIncludeDirectives(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ast.IncludeDirective
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
			p := Parser{
				stream:  stream.NewTokenStream(tt.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			p.stream.Emit() // consume '!'
			dirTok := p.stream.Emit()
			got := p.parseIncludeDirective(dirTok)

			require.False(t, p.ast.HasErrors())
			require.NotNil(t, got)
			require.Equal(t, tt.want.Kind, got.Kind)
			require.Equal(t, tt.want.Path, got.Path)
			require.Equal(t, tt.want.Tag, got.Tag)
			testutil.SemanticEqual(t, tt.want, got)
		})
	}
}

func TestParseDirectiveRecovery(t *testing.T) {
	// That is questionable feature, but unless someone hates it, there seems to be no issue
	t.Run("unknown include kind records diagnostic but produces IncludeDirective", func(t *testing.T) {
		input := "!include_unknown myfile.puml"
		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		p.stream.Emit() // consume '!'
		dirTok := p.stream.Emit()
		got := p.parseIncludeDirective(dirTok)

		require.True(t, p.ast.HasErrors())
		require.NotNil(t, got)
		require.Equal(t, ast.IncludeOnce, got.Kind)
		require.Equal(t, "myfile.puml", got.Path)
	})

	t.Run("bare exclamation mark records diagnostic and preserves surrounding statements", func(t *testing.T) {
		input := "!\nclass User"
		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		exclTok := p.stream.Emit()
		got, err := p.parseDirective(exclTok)

		require.Error(t, err)
		require.Nil(t, got)

		p.syncToLineEnd()
		p.stream.TryConsumeType(tokenizer.NEWLINE)
		userTok := p.stream.Emit()
		userEnt, err := p.parseEntity(userTok)
		require.NoError(t, err)
		require.NotNil(t, userEnt)
		require.Equal(t, "User", userEnt.Identifier.EntityName())
	})

	t.Run("exclamation mark with space before directive name records diagnostic and preserves statement", func(t *testing.T) {
		input := "! include file.puml\nclass Order"
		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		exclTok := p.stream.Emit()
		got, err := p.parseDirective(exclTok)

		require.Error(t, err)
		require.Nil(t, got)

		p.syncToLineEnd()
		p.stream.TryConsumeType(tokenizer.NEWLINE)
		orderTok := p.stream.Emit()
		orderEnt, err := p.parseEntity(orderTok)
		require.NoError(t, err)
		require.NotNil(t, orderEnt)
		require.Equal(t, "Order", orderEnt.Identifier.EntityName())
	})
}
