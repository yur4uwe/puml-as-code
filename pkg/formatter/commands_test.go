package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatCommands_Canonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Scale commands",
			input: `@startuml
scale 1.5
scale 200 width
scale 300 height
scale 200 * 100
scale max 1024 width
@enduml
`,
		},
		{
			name: "Visibility commands (hide, show, remove, restore)",
			input: `@startuml
hide empty members
show class methods
remove @unlinked
restore Component
@enduml
`,
		},
		{
			name: "Direction commands",
			input: `@startuml
left to right direction
top to bottom direction
@enduml
`,
		},
		{
			name: "Single-line text blocks (title, header, footer, legend)",
			input: `@startuml
title My Architecture Diagram
header Project Documentation
footer Confidential - Page %page% of %lastpage%
center header Centered Header
legend This is the diagram legend.
@enduml
`,
		},
		{
			name: "Multiline text blocks (title, header, footer, legend)",
			input: `@startuml
title
My Architecture diagram
with extensive title details
end title

header
Project Documentation
that spans multiple lines
end header

footer
Confidential
Page %page% of %lastpage%
end footer

header center
Centered Header
That is multiline too
end header

legend
This is the diagram legend.
described in multiple lines
end legend
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

func TestFormatCommands_Irregular(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Scale commands with irregular spacing",
			input: `@startuml
scale   max   1024   width
scale   200   *   100
scale   1.5
@enduml
`,
			expected: `@startuml
scale max 1024 width
scale 200 * 100
scale 1.5
@enduml
`,
		},
		{
			name: "Set command with irregular spacing",
			input: `@startuml
set    separator    ::
@enduml
`,
			expected: `@startuml
set separator ::
@enduml
`,
		},
		{
			name: "Direction commands with irregular spacing",
			input: `@startuml
left    to    right    direction
top    to    bottom    direction
@enduml
`,
			expected: `@startuml
left to right direction
top to bottom direction
@enduml
`,
		},
		{
			name: "Single-line title with irregular spacing",
			input: `@startuml
title    My Architecture Diagram
legend    Important diagram notes
@enduml
`,
			expected: `@startuml
title My Architecture Diagram
legend Important diagram notes
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
