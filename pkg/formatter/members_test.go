package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatMembers_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Visibility symbols (+, -, #, ~)",
			input: `@startuml
class VisibilityDemo {
  +publicField string
  -privateField int
  #protectedField bool
  ~packageField byte

  +PublicMethod()
  -privateMethod()
  #protectedMethod()
  ~packageMethod()
}
@enduml
`,
		},
		{
			name: "Member modifiers ({static}, {abstract})",
			input: `@startuml
class ModifierDemo {
  {static} +DefaultTimeout int
  {static} -instance ModifierDemo

  {abstract} +Execute() error
}
@enduml
`,
		},
		{
			name: "Method parameters, multiple returns, and complex types",
			input: `@startuml
class ServiceDemo {
  +Process(id string, count int) (Result, error)
  +Fetch(filter Map<String, String>) List<Item>
  +Execute(ctx Context)
}
@enduml
`,
		},
		{
			name: "Class separators (-- and == and .. and __)",
			input: `@startuml
class SeparatorDemo {
  +id string

  --

  +active bool

  -- Status Section --

  +status Status

  ==

  == Methods ==

  +Run()

  ..

  .. Events ..

  +OnError()

  __

  __ Internal __

  -internalKey string
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

func TestFormatMembers_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Modifiers with internal spaces and irregular member spacing",
			input: `@startuml
class ModifierDemo {
  { static }   +DefaultTimeout int
  { abstract }   +Execute() error
}
@enduml
`,
			expected: `@startuml
class ModifierDemo {
  {static} +DefaultTimeout int

  {abstract} +Execute() error
}
@enduml
`,
		},
		{
			name: "Method parameters and return types with irregular spacing",
			input: `@startuml
class ServiceDemo {
  +Process( id string , count int ) ( Result , error )
  +Execute( ctx Context )
}
@enduml
`,
			expected: `@startuml
class ServiceDemo {
  +Process(id string, count int) (Result, error)
  +Execute(ctx Context)
}
@enduml
`,
		},
		{
			name: "Separators with irregular spaces around labels",
			input: `@startuml
class SeparatorDemo {
  +id string
  --   Status Section   --
  +status Status
  ==   Methods   ==
  +Run()
}
@enduml
`,
			expected: `@startuml
class SeparatorDemo {
  +id string

  -- Status Section --

  +status Status

  == Methods ==

  +Run()
}
@enduml
`,
		},
		{
			name: "Interleaved fields and methods separated by blank lines",
			input: `@startuml
class User {
  +id string
  +SetID(id string)
  +name string
  +GetDisplayName() string
  +Validate() error
}
@enduml
`,
			expected: `@startuml
class User {
  +id string

  +SetID(id string)

  +name string

  +GetDisplayName() string
  +Validate() error
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
