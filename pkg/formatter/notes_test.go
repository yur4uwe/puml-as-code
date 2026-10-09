package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatNotes_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Inline targeted note with direction and color",
			input: `@startuml
class User {
}

note left of User : Authenticated via OAuth
note right of User #yellow : Warning: Deprecated
note top of User : Top note
note bottom of User : Bottom note
@enduml
`,
		},
		{
			name: "Inline note alias and note on link",
			input: `@startuml
A --> B
note "Active connection" as N1
note on link : TLS Encrypted
note left on link : Verified
@enduml
`,
		},
		{
			name: "Multiline block notes",
			input: `@startuml
note as N2
This is a multiline
floating note.
end note

note left of User
Detailed user instructions:
1. Register
2. Verify email
end note

note on link #cyan
Communication channel
is encrypted.
end note
@enduml
`,
		},
		{
			name: "Notes inside container with proper indentation",
			input: `@startuml
package "Auth" {
  class Session {
  }

  note right of Session : Session expires in 24h
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

			// Idempotence
			formattedAgain, err := Format(formatted)
			require.NoError(t, err)
			require.Equal(t, tt.input, formattedAgain)
		})
	}
}

func TestFormatNotes_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Irregular spacing in targeted note",
			input: `@startuml
class User {
}
note   left   of   User   #yellow   :   Warning: Deprecated
@enduml
`,
			expected: `@startuml
class User {
}

note left of User #yellow : Warning: Deprecated
@enduml
`,
		},
		{
			name: "Irregular spacing in note on link",
			input: `@startuml
A --> B
note   left   on   link   :   Verified
@enduml
`,
			expected: `@startuml
A --> B
note left on link : Verified
@enduml
`,
		},
		{
			name: "Unindented note inside container",
			input: `@startuml
package "Auth" {
  class Session {
  }
note right of Session : Session expires in 24h
}
@enduml
`,
			expected: `@startuml
package "Auth" {
  class Session {
  }

  note right of Session : Session expires in 24h
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
