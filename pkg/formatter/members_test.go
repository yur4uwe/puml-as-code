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
  {abstract} +Execute() error
  {static} -instance ModifierDemo
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
