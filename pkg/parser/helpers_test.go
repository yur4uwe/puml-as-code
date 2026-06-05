package parser

import (
	"testing"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

func TestParseSkinparamStyles(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     ast.Skinparam
		expectsError bool
	}{
		{
			name:  "Simple global skinparam",
			input: "skinparam backgroundColor #Red",
			expected: ast.Skinparam{
				"||backgroundcolor|": "#Red",
			},
		},
		{
			name:  "Target skinparam",
			input: "skinparam classBackgroundColor #Blue",
			expected: ast.Skinparam{
				"class||backgroundcolor|": "#Blue",
			},
		},
		{
			name: "Skinparam block",
			input: `skinparam class {
				BackgroundColor #Green
				HeaderFontSize 14
			}`,
			expected: ast.Skinparam{
				"class||backgroundcolor|": "#Green",
				"class|header|fontsize|":  "14",
			},
		},
		{
			name: "Nested skinparam block",
			input: `skinparam class {
				Stereotype {
					FontSize 12
					FontColor #Red
				}
				BorderColor #Black
			}`,
			expected: ast.Skinparam{
				"class|stereotype|fontsize|":  "12",
				"class|stereotype|fontcolor|": "#Red",
				"class||bordercolor|":         "#Black",
			},
		},
		{
			name:  "Skinparam with stereotype",
			input: "skinparam classBackgroundColor<<Service>> #Yellow",
			expected: ast.Skinparam{
				"class||backgroundcolor|service": "#Yellow",
			},
		},
		{
			name: "Skinparam block with stereotype",
			input: `skinparam class<<Service>> {
				BackgroundColor #Pink
			}`,
			expected: ast.Skinparam{
				"class||backgroundcolor|service": "#Pink",
			},
		},
		{
			name:         "Invalid value",
			input:        "skinparam borderThickness not-an-int",
			expectsError: true,
		},
		{
			name:         "Unknown target",
			input:        "skinparam unknownTargetBackgroundColor #Red",
			expectsError: true,
		},
		{
			name:         "Param not allowed for subtarget",
			input:        "skinparam class { Stereotype { BackgroundColor #Red } }",
			expectsError: true,
		},
		{
			name: "Deep nesting and combined names",
			input: `skinparam class {
				AttributeFontSize 10
				Stereotype {
					FontColor #Blue
				}
			}`,
			expected: ast.Skinparam{
				"class|attribute|fontsize|":  "10",
				"class|stereotype|fontcolor|": "#Blue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				stream:    tokenizer.NewTokenStream("@startuml\n" + tt.input + "\n@enduml"),
				ast:       &ast.Diagram{},
				skinparam: make(ast.Skinparam),
			}
			// Skip @startuml
			_, _ = p.stream.ReadDiagramBounds()

			var tok tokenizer.Token
			for {
				tok = p.stream.Emit()
				if tok.Type != tokenizer.NEWLINE {
					break
				}
			}
			err := p.parseStyles(tok)

			if tt.expectsError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected, p.skinparam)
		})
	}
}
