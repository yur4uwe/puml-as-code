package parser

import (
	"testing"
)

func TestParseTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no title",
			input:    "@startuml\n@enduml",
			expected: "",
		},
		{
			name:     "title with spaces and special chars",
			input:    "@startuml Title-1_2   .3!\n@enduml",
			expected: "Title-1_2   .3!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			diagram, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diagram == nil {
				t.Fatalf("diagram is nil")
			}
			if diagram.Name != tt.expected {
				t.Errorf("expected title %q, got %q", tt.expected, diagram.Name)
			}
		})
	}
}
