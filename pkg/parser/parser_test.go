package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/dialect"
	"github.com/stretchr/testify/require"
)

func TestParseTitle(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "no title",
			input:    "@startuml\n@enduml",
			expected: "",
		},
		{
			name:     "title with spaces and special chars",
			input:    "@startuml Title-1_2   .3!\n@enduml",
			expected: "Title-1_2   .3!",
		},
		{
			name:        "Uml bounds don't match types",
			input:       "@startuml\n@endgantt",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{dialect: dialect.NewGoDialect()}
			diagram, err := p.Parse(tt.input)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, diagram)
			require.NotNil(t, diagram.Statements)
			require.Equal(t, tt.expected, diagram.Name)
		})
	}
}
