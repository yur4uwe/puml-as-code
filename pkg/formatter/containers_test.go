package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatContainers_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Nested packages and classes",
			input: `@startuml
package "Outer" {
  package "Inner" {
    class Service {
      +Run()
    }
  }
}
@enduml
`,
		},
		{
			name: "Various container kinds",
			input: `@startuml
namespace Core {
  class Entity {
  }
}

together {
  class A {
  }

  class B {
  }
}

folder "Storage" {
  node Server {
    database DB {
    }
  }
}

cloud AWS {
  rectangle Gateway {
  }
}

frame AppFrame {
}
@enduml
`,
		},
		{
			name: "Container with alias, stereotype, tags, and color",
			input: `@startuml
package "Core Domain" as core <<Domain>> $backend #lightgreen {
  class Model {
  }
}
@enduml
`,
		},
		{
			name: "Custom package separator with set separator .",
			input: `@startuml
set separator .

class net.http.Client {
  +Timeout int
}

class net.http.Server {
  +Port int
}
@enduml
`,
		},
		{
			name: "Custom package separator with set separator ::",
			input: `@startuml
set separator ::

class std::vector {
  +size() int
}

class std::string {
  +length() int
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

func TestFormatContainers_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Zero indentation in nested packages and entities",
			input: `@startuml
package "Outer" {
package "Inner" {
class Service {
+Run()
}
}
}
@enduml
`,
			expected: `@startuml
package "Outer" {
  package "Inner" {
    class Service {
      +Run()
    }
  }
}
@enduml
`,
		},
		{
			name: "Irregular spacing in package signature",
			input: `@startuml
package   "Core Domain"   as   core   <<Domain>>   $backend   #lightgreen   {
  class Model {
  }
}
@enduml
`,
			expected: `@startuml
package "Core Domain" as core <<Domain>> $backend #lightgreen {
  class Model {
  }
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
