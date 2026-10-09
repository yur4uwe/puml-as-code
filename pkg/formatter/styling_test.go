package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatStyling_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Single-line skinparam settings",
			input: `@startuml
skinparam backgroundColor #EEEBDC
skinparam handwritten true
skinparam shadowing false
skinparam classFontSize 12
@enduml
`,
		},
		{
			name: "Skinparam blocks with nested sections",
			input: `@startuml
skinparam class {
  BackgroundColor PaleGreen
  ArrowColor SeaGreen
  BorderColor SpringGreen

  header {
    FontSize 12
  }
}
@enduml
`,
		},
		{
			name: "CSS style blocks (<style> ... </style>)",
			input: `@startuml
<style>
  classDiagram {
    class {
      BackGroundColor: PaleGreen
      LineColor: SeaGreen
    }
  }
</style>
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

func TestFormatStyling_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Unindented CSS style blocks and rules",
			input: `@startuml
<style>
classDiagram {
class {
BackGroundColor: PaleGreen
LineColor: SeaGreen
}
}
</style>
@enduml
`,
			expected: `@startuml
<style>
  classDiagram {
    class {
      BackGroundColor: PaleGreen
      LineColor: SeaGreen
    }
  }
</style>
@enduml
`,
		},
		{
			name: "Unindented skinparam blocks with irregular header spacing",
			input: `@startuml
skinparam    class    {
BackgroundColor PaleGreen
ArrowColor SeaGreen
}
@enduml
`,
			expected: `@startuml
skinparam class {
  BackgroundColor PaleGreen
  ArrowColor SeaGreen
}
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
