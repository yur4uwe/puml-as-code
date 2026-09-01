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
			Path: "types.go",
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

func TestGenerateFromClassDiagram_MethodsAndInterfaces(t *testing.T) {
	tbl := parseAndResolveTable(t, `
@startuml
interface Greeter {
  +Greet(name string) string
}

class User {
  +name string
  +Greet(name string) string
}
User ..|> Greeter

enum Status {
  +Active int
  +Inactive int
}
@enduml
	`)

	expectedGeneratedFiles := []*GeneratedFile{
		{
			Path: "types.go",
			Content: []byte(`package root

type Status int

const (
	StatusActive Status = iota
	StatusInactive
)

type Greeter interface {
	Greet(name string) string
}

type User struct {
	Name string
}

var _ Greeter = (*User)(nil)

func (s *User) Greet(name string) string {
	panic("not implemented")
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

func TestGenerateFromClassDiagram_UntypedSketchMembers(t *testing.T) {
	tbl := parseAndResolveTable(t, `
@startuml
class Sketch {
  +id
  +data
  +Process(input, extra)
}
@enduml
	`)

	expectedGeneratedFiles := []*GeneratedFile{
		{
			Path: "types.go",
			Content: []byte(`package root

type Sketch struct {
	Id   any
	Data any
}

func (s *Sketch) Process(input any, extra any) {
	panic("not implemented")
}
`),
		},
	}

	g := GoCodeGenerator{}
	files, err := g.GenerateFromClassDiagram(tbl)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, string(expectedGeneratedFiles[0].Content), string(files[0].Content))
}
