package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
)

func TestParseStyleBlock(t *testing.T) {
	input := `@startuml
<style>
root {
    FontColor: #333333;
    FontName: Arial;
}
classDiagram {
    class {
        BackGroundColor: PaleGreen;
        LineColor: SeaGreen;
    }
}
</style>
class User {
}
@enduml`

	p := &Parser{
		Dialect: dialect.NewGoDialect(),
	}
	diag, err := p.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, diag)

	var styleRules []*ast.StyleRule
	for _, stmt := range diag.Statements {
		if r, ok := stmt.(*ast.StyleRule); ok {
			styleRules = append(styleRules, r)
		}
	}

	require.Len(t, styleRules, 2)

	// Rule 1: root
	require.Equal(t, []string{"root"}, styleRules[0].Selectors)
	require.False(t, styleRules[0].IsSkinparam)
	require.Equal(t, "#333333", styleRules[0].Properties["FontColor"])
	require.Equal(t, "Arial", styleRules[0].Properties["FontName"])

	// Rule 2: classDiagram -> class
	require.Equal(t, []string{"classDiagram", "class"}, styleRules[1].Selectors)
	require.False(t, styleRules[1].IsSkinparam)
	require.Equal(t, "PaleGreen", styleRules[1].Properties["BackGroundColor"])
	require.Equal(t, "SeaGreen", styleRules[1].Properties["LineColor"])
}
