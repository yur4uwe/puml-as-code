package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadIdentifier(t *testing.T) {
	lex := NewLexer("tag13")
	require.Equal(t, "tag13", lex.readIdentifier())

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

func TestReadLineComment(t *testing.T) {
	lex := NewLexer("' comment here\n")
	line := lex.readLineComment()
	require.Equal(t, "comment here", line)
}

func TestSetSeparator(t *testing.T) {
	lex := NewLexer("set separator ::\nclass A::B")

	// set
	tok := lex.Emit()
	require.Equal(t, SET, tok.Type)

	// separator
	tok = lex.Emit()
	require.Equal(t, IDENTIFIER, tok.Type)
	require.Equal(t, "separator", tok.Literal)

	// ::
	tok = lex.Emit()
	require.Equal(t, SEPARATOR, tok.Type)
	require.Equal(t, "::", tok.Literal)
	require.Equal(t, []rune("::"), lex.packageSeparator)

	// NEWLINE
	tok = lex.Emit()
	require.Equal(t, NEWLINE, tok.Type)

	// class
	tok = lex.Emit()
	require.Equal(t, CLASS, tok.Type)

	// A
	tok = lex.Emit()
	require.Equal(t, IDENTIFIER, tok.Type)
	require.Equal(t, "A", tok.Literal)

	// :: (as separator)
	tok = lex.Emit()
	require.Equal(t, SEPARATOR, tok.Type)
	require.Equal(t, "::", tok.Literal)

	// B
	tok = lex.Emit()
	require.Equal(t, IDENTIFIER, tok.Type)
	require.Equal(t, "B", tok.Literal)
}
