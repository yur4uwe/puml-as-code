package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmitRawSpan(t *testing.T) {
	input := `@startuml
' pac:fmt:off
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
}
