package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatPartialPreservesBrokenSource(t *testing.T) {
	for _, tt := range []struct{ name, input, want string }{
		{"unclosed entity", "@startuml\nclass User {\n+id string\n@enduml\n", "@startuml\nclass User {\n  +id string\n@enduml\n"},
		{"unclosed container", "@startuml\npackage domain {\nclass User\n@enduml\n", "@startuml\npackage domain {\n  class User\n@enduml\n"},
		{"malformed arrow", "@startuml\nA -unknown- B\n@enduml", "@startuml\nA -unknown- B\n@enduml\n"},
		{"missing target", "@startuml\nA -->\n@enduml", "@startuml\nA -->\n@enduml\n"},
		{"bad member", "@startuml\nclass User {\n@oops\n}\n@enduml", "@startuml\nclass User {\n  @oops\n}\n@enduml\n"},
		{"bad statement", "@startuml\n?oops\n@enduml", "@startuml\n?oops\n@enduml\n"},
		{"unclosed note", "@startuml\nnote right of User\nkeep this\n@enduml", "@startuml\nnote right of User\nkeep this\n@enduml\n"},
		{"unclosed style", "@startuml\n<style>\nclassDiagram {\nFontColor: red\n}\n@enduml", "@startuml\n<style>\n  classDiagram {\n    FontColor: red\n  }\n@enduml\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := FormatPartial(tt.input)
			require.Error(t, err)
			require.Equal(t, tt.want, out)
			again, err := FormatPartial(out)
			require.Error(t, err)
			require.Equal(t, out, again)
			strict, err := Format(tt.input)
			require.Error(t, err)
			require.Empty(t, strict)
		})
	}
}

func TestFormatPartialKeepsInputWithoutParsedBounds(t *testing.T) {
	for _, input := range []string{"@star", "class Thing", "@startuml\nclass User\n@enduml\n@startuml\nclass Other\n@enduml"} {
		out, _ := FormatPartial(input)
		require.Equal(t, input, out)
	}
}
