package gogenerator

import "slices"

type NotesView struct {
	Notes []string
}

type TriviaView struct {
	LeadingTrivia  []string
	TrailingTrivia []string
}

// FileView represents a single .go file being generated
type FileView struct {
	PackageName string
	ImportPath  string
	Imports     []string
	Structs     []StructView
	Interfaces  []InterfaceView
	Enums       []EnumView
}

func (f *FileView) AddImport(importPath string) {
	// Potentially can improve search performace with binary search
	if importPath == "" || importPath == f.PackageName || importPath == f.ImportPath {
		return
	}
	// Deduplicate
	if !slices.Contains(f.Imports, importPath) {
		f.Imports = append(f.Imports, importPath)
	}
}

type StructView struct {
	Name       string // e.g. "User"
	Embeds     []string
	Implements []string
	Fields     []FieldView
	Methods    []MethodView
	NotesView
	TriviaView
}

type InterfaceView struct {
	Name    string
	Embeds  []string
	Methods []MethodView
	NotesView
	TriviaView
}

type EnumView struct {
	Name   string
	Values []string
	NotesView
	TriviaView
}

type FieldView struct {
	Name    string // Pre-formatted with visibility: "ID" vs "password"
	Type    string // e.g. "string", "*Order", "[]Item"
	Comment string // e.g. "// protected" or empty
	NotesView
	TriviaView
}

type MethodView struct {
	Name      string // Pre-formatted name, e.g. "Read"
	Signature string // e.g. "(p []byte) (n int, err error)"
	NotesView
	TriviaView
}
