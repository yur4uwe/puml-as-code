package gogenerator

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"

	"github.com/stretchr/testify/require"
)

func TestGenerateFromClassDiagram(t *testing.T) {
	cd := &ast.ClassDiagram{
		Classes: map[string]*ast.Class{
			"Animal": {
				Name: "Animal",
				Attributes: []ast.Attribute{
					{Name: "name", Type: ast.TypeRef{Kind: ast.String}, Visibility: ast.Public},
					{Name: "age", Type: ast.TypeRef{Kind: ast.Int}, Visibility: ast.Private},
				},
				Methods: []ast.Method{
					{Name: "Speak", ReturnType: ast.TypeRef{Kind: ast.String}, Visibility: ast.Public},
					{Name: "eat", ReturnType: ast.TypeRef{Kind: ast.Void}, Visibility: ast.Private, Parameters: []ast.Attribute{
						{Name: "food", Type: ast.TypeRef{Kind: ast.String}},
					}},
				},
			},
		},
	}

	src, err := GoCodeGenerator{}.GenerateFromClassDiagram(cd)
	require.NoError(t, err)
	require.NotEmpty(t, src)

	require.Contains(t, src, "package generated")
	require.Contains(t, src, "type Animal struct")
	require.Regexp(t, `\bName\s+string`, src)
	require.Regexp(t, `\bage\s+int`, src)
	require.Contains(t, src, "func (a *Animal) Speak() string")
	require.Contains(t, src, "func (a *Animal) eat(food string)")
}

func TestGenerateNilDiagram(t *testing.T) {
	_, err := GoCodeGenerator{}.GenerateFromClassDiagram(nil)
	require.Error(t, err)
}
