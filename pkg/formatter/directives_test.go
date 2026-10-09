package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatPragmasAndDirectives_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "pac:fmt:off and pac:fmt:on blocks",
			input: `@startuml
' pac:fmt:off
package "Unformatted" {
    class   RawClass   {
      +id:string
    }
}
' pac:fmt:on
class FormattedClass {
  +id string
}
@enduml
`,
		},
		{
			name: "pac:fmt:off disables formatting for entire remaining file",
			input: `@startuml
class A {
  +x int
}
' pac:fmt:off
class   B   {
    +y   int
}
@enduml
`,
		},
		{
			name: "Include directive once",
			input: `@startuml
!include ./common/types.puml
class Service {
}
@enduml
`,
		},
		{
			name: "Include many directive",
			input: `@startuml
!include_many ./models/user.puml
class Controller {
}
@enduml
`,
		},
		{
			name: "Include directive with tag",
			input: `@startuml
!include ./schemas.puml!USER_SCHEMA
!include_many ./schemas.puml!AUTH_SCHEMA
@enduml
`,
		},
		{
			name: "Include directive nested inside package",
			input: `@startuml
package "Core" {
  !include ./core/base.puml
}
@enduml
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := Format(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.input, formatted)

			// Idempotence check
			formattedAgain, err := Format(formatted)
			require.NoError(t, err)
			require.Equal(t, tt.input, formattedAgain)
		})
	}
}

func TestFormatPragmasAndDirectives_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Unindented include inside container",
			input: `@startuml
package "Core" {
!include ./core/base.puml
}
@enduml
`,
			expected: `@startuml
package "Core" {
  !include ./core/base.puml
}
@enduml
`,
		},
		{
			name: "Irregular spacing in include directive",
			input: `@startuml
!include_many   ./models/user.puml
@enduml
`,
			expected: `@startuml
!include_many ./models/user.puml
@enduml
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := Format(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, formatted)

			// Idempotence check
			formattedAgain, err := Format(formatted)
			require.NoError(t, err)
			require.Equal(t, tt.expected, formattedAgain)
		})
	}
}
