package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatRelationships_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Basic arrow kinds",
			input: `@startuml
A --> B
C <|-- D
E *-- F
G o-- H
I ..> J
K <|.. L
@enduml
`,
		},
		{
			name: "Relationships with labels and cardinalities",
			input: `@startuml
Order "*" --> "1" Customer : placed by
Company "1" *-- "0..*" Department : contains
Teacher "1..*" -- "1..*" Student : teaches
@enduml
`,
		},
		{
			name: "Arrow orientations (up, down, left, right)",
			input: `@startuml
A -up-> B
C -down-> D
E -left-> F
G -right-> H
@enduml
`,
		},
		{
			name: "Arrow attributes (thickness, color, dashed)",
			input: `@startuml
A -[bold]-> B
C -[dashed,#red]-> D
E -[#blue,thickness=2]-> F
@enduml
`,
		},
		{
			name: "Relationships with qualified package names and members",
			input: `@startuml
net.http.Client --> net.http.Response : creates
auth.User::id --> db.Record::pk
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

func TestFormatRelationships_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Compact relationship with multiplicities and label",
			input: `@startuml
Order"*"-->"1"Customer:placed by
Company"1"*--"0..*"Department:contains
@enduml
`,
			expected: `@startuml
Order "*" --> "1" Customer : placed by
Company "1" *-- "0..*" Department : contains
@enduml
`,
		},
		{
			name: "Excessive whitespace in relationships",
			input: `@startuml
Teacher    "1..*"    --    "1..*"    Student    :    teaches
A    -->    B
@enduml
`,
			expected: `@startuml
Teacher "1..*" -- "1..*" Student : teaches
A --> B
@enduml
`,
		},
		{
			name: "Irregular spacing in arrow attributes",
			input: `@startuml
C   -[ dashed , #red ]->   D
E   -[ #blue , thickness=2 ]->   F
@enduml
`,
			expected: `@startuml
C -[dashed,#red]-> D
E -[#blue,thickness=2]-> F
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
