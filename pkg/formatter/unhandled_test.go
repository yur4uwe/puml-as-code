package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmitRawSpan(t *testing.T) {
	t.Run("canonical unhandled block inside container", func(t *testing.T) {
		input := `@startuml
package "Foo" {
  !if $foo
    class "Bar"
  !endif
}
@enduml
`

		formatted, err := Format(input)
		require.NoError(t, err)
		require.Equal(t, input, formatted)

		formattedAgain, err := Format(formatted)
		require.NoError(t, err)
		require.Equal(t, formatted, formattedAgain)
	})

	t.Run("unformatted unhandled block receives container indentation", func(t *testing.T) {
		input := `@startuml
package "Foo" {
!if $foo
  class "Bar"
!endif
}
@enduml
`
		expected := `@startuml
package "Foo" {
  !if $foo
    class "Bar"
  !endif
}
@enduml
`

		formatted, err := Format(input)
		require.NoError(t, err)
		require.Equal(t, expected, formatted)

		formattedAgain, err := Format(formatted)
		require.NoError(t, err)
		require.Equal(t, expected, formattedAgain)
	})
}
