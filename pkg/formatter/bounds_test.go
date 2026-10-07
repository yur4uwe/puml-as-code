package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatDiagramBounds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Bare diagram bounds",
			input: `@startuml
@enduml
`,
			expected: `@startuml
@enduml
`,
		},
		{
			name: "With trailing name",
			input: `@startuml my_diagram.puml
@enduml
`,
			expected: `@startuml my_diagram.puml
@enduml
`,
		},
		{
			name: "With ID param",
			input: `@startuml(id=CORE)
@enduml
`,
			expected: `@startuml(id=CORE)
@enduml
`,
		},
		{
			name: "With multiple params in parentheses",
			input: `@startuml(id=CORE, env=prod)
@enduml
`,
			expected: `@startuml(id=CORE, env=prod)
@enduml
`,
		},
		{
			name: "With ID param and trailing filename",
			input: `@startuml(id=CORE) diagram.png
@enduml
`,
			expected: `@startuml(id=CORE) diagram.png
@enduml
`,
		},
		{
			name: "Tool options with filename only",
			input: `@startuml{filename.puml}
@enduml
`,
			expected: `@startuml{filename.puml}
@enduml
`,
		},
		{
			name: "Tool options with filename and caption",
			input: `@startuml{filename.puml, Overview Diagram}
@enduml
`,
			expected: `@startuml{filename.puml, Overview Diagram}
@enduml
`,
		},
		{
			name: "Tool options with filename and key-value option",
			input: `@startuml{filename.puml, width=5cm}
@enduml
`,
			expected: `@startuml{filename.puml, width=5cm}
@enduml
`,
		},
		{
			name: "Tool options with filename, caption, and options",
			input: `@startuml{filename.puml, Overview Diagram, width=5cm}
@enduml
`,
			expected: `@startuml{filename.puml, Overview Diagram, width=5cm}
@enduml
`,
		},
		{
			name: "ID in parentheses followed by tool options",
			input: `@startuml(id=AUTH){filename.puml, Overview Diagram, width=5cm}
@enduml
`,
			expected: `@startuml(id=AUTH){filename.puml, Overview Diagram, width=5cm}
@enduml
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := Format(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, formatted)
		})
	}
}
