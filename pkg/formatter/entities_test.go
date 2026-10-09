package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatEntities_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Simple class, interface, and enum",
			input: `@startuml
class User {
}
interface Repository {
}
enum Status {
}
@enduml
`,
		},
		{
			name: "Abstract class and entity class",
			input: `@startuml
abstract class BaseEntity {
}
entity AuditRecord {
}
@enduml
`,
		},
		{
			name: "Entity with generics, stereotype, tags, and color",
			input: `@startuml
class Container <T> <<Generic>> $core $model #lightblue {
  +items T[]
}
@enduml
`,
		},
		{
			name: "Entity with alias",
			input: `@startuml
class "Custom User Service" as UserService {
}
@enduml
`,
		},
		{
			name: "Inline member declarations on entities",
			input: `@startuml
class User {
}
User : +id string
User : +GetName() string
@enduml
`,
		},
		{
			name: "Other entity kinds (struct, protocol, record, exception, dataclass)",
			input: `@startuml
struct Point {
  +x int
  +y int
}
protocol Printable {
  +Print()
}
record Person {
  +name string
}
exception NotFoundError {
  +msg string
}
dataclass Config {
  +env string
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

func TestFormatEntities_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Unindented entity body members",
			input: `@startuml
class User {
+id string
+GetName() string
}
@enduml
`,
			expected: `@startuml
class User {
  +id string
  +GetName() string
}
@enduml
`,
		},
		{
			name: "Irregular spacing in entity signature",
			input: `@startuml
class   Container   <T>   <<Generic>>   $core   $model   #lightblue   {
  +items T[]
}
@enduml
`,
			expected: `@startuml
class Container <T> <<Generic>> $core $model #lightblue {
  +items T[]
}
@enduml
`,
		},
		{
			name: "Irregular spacing in inline entity member declarations",
			input: `@startuml
class User {
}
User   :   +id string
User   :   +GetName() string
@enduml
`,
			expected: `@startuml
class User {
}
User : +id string
User : +GetName() string
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
