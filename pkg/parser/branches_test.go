package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"

	"github.com/stretchr/testify/require"
)

func TestParseEntity(t *testing.T) {
	// TODO
}

func TestParseFieldOrMethod(t *testing.T) {
	// TODO
}

func TestParseEntityMember(t *testing.T) {
	tt := []struct {
		name  string
		input string
		want  ast.Member
	}{}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				dialect: dialect.NewGoDialect(),
			}
			p.Parse(tc.input)
			got, err := p.parseEntityMember()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
