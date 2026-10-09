package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatDiagramBounds_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Bare diagram bounds",
			input: `@startuml
@enduml
`,
		},
		{
			name: "With trailing name",
			input: `@startuml my_diagram.puml
@enduml
`,
		},
		{
			name: "With ID param",
			input: `@startuml(id=CORE)
@enduml
`,
		},
		{
			name: "With multiple params in parentheses",
			input: `@startuml(id=CORE, env=prod)
@enduml
`,
		},
		{
			name: "With ID param and trailing filename",
			input: `@startuml(id=CORE) diagram.png
@enduml
`,
		},
		{
			name: "Tool options with filename only",
			input: `@startuml{filename.puml}
@enduml
`,
		},
		{
			name: "Tool options with filename and caption",
			input: `@startuml{filename.puml, Overview Diagram}
@enduml
`,
		},
		{
			name: "Tool options with filename and key-value option",
			input: `@startuml{filename.puml, width=5cm}
@enduml
`,
		},
		{
			name: "Tool options with filename, caption, and options",
			input: `@startuml{filename.puml, Overview Diagram, width=5cm}
@enduml
`,
		},
		{
			name: "ID in parentheses followed by tool options",
			input: `@startuml(id=AUTH){filename.puml, Overview Diagram, width=5cm}
@enduml
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := Format(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.input, formatted)

			// Idempotence
			formattedAgain, err := Format(formatted)
			require.NoError(t, err)
			require.Equal(t, tt.input, formattedAgain)
		})
	}
}

func TestFormatDiagramBounds_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Irregular spacing in ID parentheses",
			input: `@startuml ( id = CORE )
@enduml
`,
			expected: `@startuml(id=CORE)
@enduml
`,
		},
		{
			name: "Irregular spacing in multiple params",
			input: `@startuml ( id = CORE ,   env = prod )
@enduml
`,
			expected: `@startuml(id=CORE, env=prod)
@enduml
`,
		},
		{
			name: "Irregular spacing in tool options",
			input: `@startuml {  filename.puml  ,   Overview Diagram  ,   width=5cm  }
@enduml
`,
			expected: `@startuml{filename.puml, Overview Diagram, width=5cm}
@enduml
`,
		},
		{
			name: "Excessive whitespace before trailing filename",
			input: `@startuml    my_diagram.puml
@enduml
`,
			expected: `@startuml my_diagram.puml
@enduml
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := Format(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, formatted)

			// Idempotence
			formattedAgain, err := Format(formatted)
			require.NoError(t, err)
			require.Equal(t, tt.expected, formattedAgain)
		})
	}
}
