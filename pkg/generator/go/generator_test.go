package gogenerator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateFromClassDiagram(t *testing.T) {
	tbl := parseAndResolveTable(t, `
@startuml
class User {
  +name string
}
@enduml
	`)

	expectedGeneratedFiles := []*GeneratedFile{
		{
			Path: "/",
			Content: []byte(`package root

type User struct {
	Name string
}
`),
		},
	}

	g := GoCodeGenerator{}
	files, err := g.GenerateFromClassDiagram(tbl)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, expectedGeneratedFiles, files)
}
