package gogenerator

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/resolver"

	"github.com/stretchr/testify/require"
)

func TestTargetTypeName(t *testing.T) {
	tt := []struct {
		name   string
		srcPkg []string
		target *resolver.EntitySymbol
		want   string
	}{
		{
			name:   "same package",
			srcPkg: []string{"models"},
			target: &resolver.EntitySymbol{
				FQN: "models.User",
				AST: &ast.Entity{
					Identifier: "User",
				},
				PackagePath: []string{"models"},
			},
			want: "User",
		},
		{
			name:   "different package",
			srcPkg: []string{"orders"},
			target: &resolver.EntitySymbol{
				FQN: "models.User",
				AST: &ast.Entity{
					Identifier: "User",
				},
				PackagePath: []string{"models"},
			},
			want: "models.User",
		},
		{
			name:   "no package",
			srcPkg: []string{},
			target: &resolver.EntitySymbol{
				FQN: "models.User",
				AST: &ast.Entity{
					Identifier: "User",
				},
				PackagePath: []string{"models"},
			},
			want: "models.User",
		},
		{
			name: "target is in root package",
			srcPkg: []string{
				"orders",
			},
			target: &resolver.EntitySymbol{
				FQN: "User",
				AST: &ast.Entity{
					Identifier: "User",
				},
				PackagePath: []string{},
			},
			want: "root.User",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resultType := targetTypeName(tc.srcPkg, tc.target)
			require.Equal(t, tc.want, resultType)
		})
	}
}
