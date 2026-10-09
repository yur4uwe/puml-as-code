package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatOptions(t *testing.T) {
	input := `@startuml
package "Core" {
  class User {
    +id string
    +GetName() string
  }
}
@enduml
`

	t.Run("Default options (2 spaces)", func(t *testing.T) {
		expected := `@startuml
package "Core" {
  class User {
    +id string
    +GetName() string
  }
}
@enduml
`
		formatted, err := Format(input)
		require.NoError(t, err)
		require.Equal(t, expected, formatted)
	})

	t.Run("Custom indent size (4 spaces)", func(t *testing.T) {
		expected := `@startuml
package "Core" {
    class User {
        +id string
        +GetName() string
    }
}
@enduml
`
		formatted, err := Format(input, FormatOptions{
			IndentSize: 4,
			UseTabs:    false,
		})
		require.NoError(t, err)
		require.Equal(t, expected, formatted)

		// Idempotence with options
		formattedAgain, err := Format(formatted, FormatOptions{
			IndentSize: 4,
			UseTabs:    false,
		})
		require.NoError(t, err)
		require.Equal(t, formatted, formattedAgain)
	})

	t.Run("Use tabs indentation", func(t *testing.T) {
		expected := "@startuml\npackage \"Core\" {\n\tclass User {\n\t\t+id string\n\t\t+GetName() string\n\t}\n}\n@enduml\n"
		formatted, err := Format(input, FormatOptions{
			UseTabs: true,
		})
		require.NoError(t, err)
		require.Equal(t, expected, formatted)

		// Idempotence with tabs
		formattedAgain, err := Format(formatted, FormatOptions{
			UseTabs: true,
		})
		require.NoError(t, err)
		require.Equal(t, formatted, formattedAgain)
	})
}
