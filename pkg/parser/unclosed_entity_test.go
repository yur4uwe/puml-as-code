package parser

import (
	"errors"
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"

	"github.com/stretchr/testify/require"
)

func TestUnclosedEntityHintsAtFirstBadMember(t *testing.T) {
	for _, tt := range []struct {
		name, body, closer string
		hint, unclosed     bool
	}{
		{"at diagram end", "+id string\nname : bad\nother : bad\n", "@enduml", true, true},
		{"at eof", "+id string\nname : bad", "", true, true},
		{"valid member after bad one", "name : bad\n+id string\nother : bad\n", "@enduml", false, true},
		{"closed entity", "name : bad\n}\n", "@enduml", false, false},
		{"no bad member", "+id string\n", "@enduml", false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			diagram, _ := NewParser(dialect.NewGoDialect()).Parse("@startuml\nclass Example {\n" + tt.body + tt.closer)
			var unclosed *ast.Diagnostic
			for _, diag := range diagram.Diagnostics {
				if errors.Is(diag, ast.ErrUnclosedBlock) {
					unclosed = diag
				}
			}
			if !tt.unclosed {
				require.Nil(t, unclosed)
				return
			}
			require.NotNil(t, unclosed)
			if !tt.hint {
				require.Empty(t, unclosed.RelatedInformation)
				return
			}
			members := diagram.Statements[1].(*ast.Entity).Members
			first := members[1].(*ast.BadMember)
			require.Len(t, unclosed.RelatedInformation, 1)
			require.Equal(t, first.Span(), unclosed.RelatedInformation[0].Span)
			require.Contains(t, unclosed.RelatedInformation[0].Message, "close the block")
		})
	}
}
