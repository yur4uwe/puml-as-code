package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModeStack(t *testing.T) {
	lex := NewLexer("")
	require.Equal(t, MODE_DEFAULT, lex.CurrentMode())

	lex.PushMode(MODE_NOTE)
	require.Equal(t, MODE_NOTE, lex.CurrentMode())

	lex.PushMode(MODE_CLASS)
	require.Equal(t, MODE_CLASS, lex.CurrentMode())

	require.Equal(t, MODE_CLASS, lex.PopMode())
	require.Equal(t, MODE_NOTE, lex.CurrentMode())

	require.Equal(t, MODE_NOTE, lex.PopMode())
	require.Equal(t, MODE_DEFAULT, lex.CurrentMode())
}

func TestReadIdentifier(t *testing.T) {
	lex := NewLexer("$tag13")
	require.Equal(t, "$tag13", lex.readIdentifier())

	lex = NewLexer("Foo\\-bar")
	require.Equal(t, "Foo-bar", lex.readIdentifier())

	lex = NewLexer("Foo123")
	require.Equal(t, "Foo123", lex.readIdentifier())
}

func TestReadNumber(t *testing.T) {
	lex := NewLexer("12.34")
	require.Equal(t, "12.34", lex.readNumber())

	lex = NewLexer("007")
	require.Equal(t, "007", lex.readNumber())
}

func TestReadString(t *testing.T) {
	lex := NewLexer("\"a\\\"b\"")
	require.Equal(t, "\"a\\\"b\"", lex.readString())
}

func TestReadRelationWithStyle(t *testing.T) {
	lex := NewLexer("-[bold]->")
	ok := lex.readRelation()
	require.Equal(t, RELATIONSHIP, ok.Type)
	require.Equal(t, "-[bold]->", ok.Literal)
}

func TestReadRelationWithDirection(t *testing.T) {
	lex := NewLexer("-left->")
	ok := lex.readRelation()
	require.Equal(t, RELATIONSHIP, ok.Type)
	require.Equal(t, "-left->", ok.Literal)
}

func TestReadUMLBounds(t *testing.T) {
	lex := NewLexer("@startuml")
	tok := lex.readUMLBounds()
	require.Equal(t, START, tok.Type)
	require.Equal(t, "startuml", tok.Literal)

	lex = NewLexer("@enduml")
	tok = lex.readUMLBounds()
	require.Equal(t, END, tok.Type)
	require.Equal(t, "enduml", tok.Literal)

	lex = NewLexer("@unlinked")
	tok = lex.readUMLBounds()
	require.Equal(t, IDENTIFIER, tok.Type)
	require.Equal(t, "@unlinked", tok.Literal)
}

func TestReadModifier(t *testing.T) {
	lex := NewLexer("{abstract}")
	tok := lex.readModifier()
	require.Equal(t, MODIFIER, tok.Type)
	require.Equal(t, "{abstract}", tok.Literal)
}

func TestReadStereotype(t *testing.T) {
	lex := NewLexer("<<stereotype>>")
	tok := lex.readStereotype()
	require.Equal(t, STEREOTYPE, tok.Type)
	require.Equal(t, "<<stereotype>>", tok.Literal)
}

func TestReadLabel(t *testing.T) {
	lex := NewLexer("Label text__")
	label := lex.readLabel()
	require.Equal(t, "Label text", label)
	require.Equal(t, '_', lex.ch)
}

func TestReadNoteLine(t *testing.T) {
	lex := NewLexer("  hello <b>world</b>\n")
	line := lex.readNoteLine()
	require.Equal(t, "hello <b>world</b>", line)
}

func TestReadLineComment(t *testing.T) {
	lex := NewLexer("' comment here\n")
	line := lex.readLineComment()
	require.Equal(t, "comment here", line)
}

func TestReadUntil(t *testing.T) {
	lex := NewLexer("abc:def")
	out := lex.readUntil(':')
	require.Equal(t, "abc", out)
	require.Equal(t, ':', lex.ch)
}

func TestReadProperty(t *testing.T) {
	lex := NewLexer(" separator ::\n")
	prop := lex.readProperty()
	require.Equal(t, "separator=::", prop)
	require.Equal(t, []rune("::"), lex.packageSeparator)
}

func TestIsPackageSeparator(t *testing.T) {
	lex := NewLexer("::foo")
	lex.packageSeparator = []rune("::")
	require.True(t, lex.isPackageSeparator())
	require.Equal(t, 'f', lex.ch)
}
