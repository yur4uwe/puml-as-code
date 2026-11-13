package parser

import (
	"testing"

	"yur4uwe/pac/internal/testutil"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/stream"

	"github.com/stretchr/testify/require"
)

func TestParseScaleCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ast.ScaleCommand
		expectErr bool
	}{
		// bare integer factor
		{
			name:  "scale 200",
			input: "scale 200",
			want:  ast.ScaleCommand{Lhs: "200"},
		},
		// decimal factors, digits preserved verbatim
		{
			name:  "scale 1.5",
			input: "scale 1.5",
			want:  ast.ScaleCommand{Lhs: "1", Sep: ".", Rhs: "5"},
		},
		{
			name:  "scale 1.50 keeps trailing zero",
			input: "scale 1.50",
			want:  ast.ScaleCommand{Lhs: "1", Sep: ".", Rhs: "50"},
		},
		{
			name:  "scale 1.05 keeps leading zero",
			input: "scale 1.05",
			want:  ast.ScaleCommand{Lhs: "1", Sep: ".", Rhs: "05"},
		},
		// fraction factor
		{
			name:  "scale 2/3",
			input: "scale 2/3",
			want:  ast.ScaleCommand{Lhs: "2", Sep: "/", Rhs: "3"},
		},
		// single dimension
		{
			name:  "scale 200 width",
			input: "scale 200 width",
			want:  ast.ScaleCommand{Lhs: "200", Unit: "width"},
		},
		{
			name:  "scale 200 height",
			input: "scale 200 height",
			want:  ast.ScaleCommand{Lhs: "200", Unit: "height"},
		},
		// boxes, both separators
		{
			name:  "scale 200*100",
			input: "scale 200*100",
			want:  ast.ScaleCommand{Lhs: "200", Sep: "*", Rhs: "100"},
		},
		{
			name:  "scale 200x100",
			input: "scale 200x100",
			want:  ast.ScaleCommand{Lhs: "200", Sep: "x", Rhs: "100"},
		},
		// scale with 'x' box edge cases
		{
			name:  "scale 200x 100",
			input: "scale 200x 100",
			want:  ast.ScaleCommand{Lhs: "200", Sep: "x", Rhs: "100"},
		},
		{
			name:  "scale 200 x100",
			input: "scale 200 x100",
			want:  ast.ScaleCommand{Lhs: "200", Sep: "x", Rhs: "100"},
		},
		// max variants
		{
			name:  "scale max 300*200",
			input: "scale max 300*200",
			want:  ast.ScaleCommand{IsMax: true, Lhs: "300", Sep: "*", Rhs: "200"},
		},
		{
			name:  "scale max 300x200",
			input: "scale max 300x200",
			want:  ast.ScaleCommand{IsMax: true, Lhs: "300", Sep: "x", Rhs: "200"},
		},
		{
			name:  "scale max 1024 width",
			input: "scale max 1024 width",
			want:  ast.ScaleCommand{IsMax: true, Lhs: "1024", Unit: "width"},
		},
		{
			name:  "scale max 800 height",
			input: "scale max 800 height",
			want:  ast.ScaleCommand{IsMax: true, Lhs: "800", Unit: "height"},
		},
		// whitespace tolerance
		{
			name:  "extra spaces",
			input: "scale   max   1024   width",
			want:  ast.ScaleCommand{IsMax: true, Lhs: "1024", Unit: "width"},
		},
		{name: "leading dot", input: "scale .5", want: ast.ScaleCommand{Lhs: ".5"}},
		{name: "trailing dot", input: "scale 1.", want: ast.ScaleCommand{Lhs: "1."}},
		{name: "trailing dot unit", input: "scale 1. width", want: ast.ScaleCommand{Lhs: "1.", Unit: "width"}},
		{name: "decimal unit", input: "scale 1.5 height", want: ast.ScaleCommand{Lhs: "1.5", Unit: "height"}},
		{name: "leading zeros", input: "scale 007", want: ast.ScaleCommand{Lhs: "007"}},
		{name: "zero width", input: "scale 0 width", want: ast.ScaleCommand{Lhs: "0", Unit: "width"}},
		{name: "spaced box", input: "scale 200 * 100", want: ast.ScaleCommand{Lhs: "200", Sep: "*", Rhs: "100"}},
		{name: "spaced x box", input: "scale 200 x 100", want: ast.ScaleCommand{Lhs: "200", Sep: "x", Rhs: "100"}},
		{name: "zero in spaced x box", input: "scale 0 x100", want: ast.ScaleCommand{Lhs: "0", Sep: "x", Rhs: "100"}},

		// errors
		{name: "no operand", input: "scale", expectErr: true},
		{name: "max without operand", input: "scale max", expectErr: true},
		{name: "non-numeric", input: "scale abc", expectErr: true},
		{name: "unknown unit", input: "scale 200 depth", expectErr: true},
		{name: "box with unit", input: "scale 200*100 width", expectErr: true},
		{name: "compact box with unit", input: "scale 200x100 width", expectErr: true},
		{name: "decimal box", input: "scale 200 x 100.5", expectErr: true},
		{name: "dangling decimal point", input: "scale .", expectErr: true},
		{name: "dangling slash", input: "scale 2/", expectErr: true},
		{name: "dangling box separator", input: "scale 200*", expectErr: true},
		{name: "two decimal points", input: "scale 1.2.3", expectErr: true},
		{name: "mixed separators", input: "scale 2/3*4", expectErr: true},
		{name: "hex", input: "scale 0x10", expectErr: true},
		{name: "uppercase hex", input: "scale 0X10", expectErr: true},
		{name: "octal", input: "scale 0o12", expectErr: true},
		{name: "scientific", input: "scale 1e3", expectErr: true},
		{name: "binary", input: "scale 0b11 width", expectErr: true},
		{name: "max integer", input: "scale max 200", expectErr: true},
		{name: "max decimal", input: "scale max 1.5", expectErr: true},
		{name: "fraction unit", input: "scale 2/3 width", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream: stream.NewTokenStream(tc.input),
				ast:    &ast.Diagram{},
			}
			got := p.parseScale(p.stream.Emit())
			if tc.expectErr {
				require.Len(t, p.ast.Diagnostics, 1)
				require.Same(t, got, p.ast.Diagnostics[0].Node)
			} else {
				require.Empty(t, p.ast.Diagnostics)
				testutil.SemanticEqual(t, tc.want, got)
			}
		})
	}
}

func TestScaleRecoversAtNextStatement(t *testing.T) {
	p := NewParser(dialect.LaxDialect{})
	diagram, err := p.Parse("@startuml\nscale max 200x\nclass Next\n@enduml\n")
	require.Error(t, err)
	require.Len(t, diagram.Diagnostics, 1)
	require.Len(t, diagram.Statements, 4)
	scale := diagram.Statements[1].(*ast.ScaleCommand)
	require.True(t, scale.IsMax)
	require.Equal(t, "200", scale.Lhs)
	require.Equal(t, "x", scale.Sep)
	require.Same(t, scale, diagram.Diagnostics[0].Node)
	require.Equal(t, "Next", diagram.Statements[2].(*ast.Entity).Identifier.EntityName())
}
