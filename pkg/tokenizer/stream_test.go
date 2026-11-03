package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTokenStream(t *testing.T) {
	ts := NewTokenStream("")
	require.NotNil(t, ts)
	tok := ts.PeekRawTokenAt(0)
	require.Equal(t, EOF, tok.Type)
}

func TestPeekTokenAt(t *testing.T) {
	ts := NewTokenStream("package .\nset separator ::\npackage :: .")
	seq := []Token{
		{Type: IDENTIFIER, Literal: "package"},
		{Type: DOT, Literal: "."},
		{Type: NEWLINE, Literal: "\n"},
		{Type: IDENTIFIER, Literal: "set"},
		{Type: IDENTIFIER, Literal: "separator"},
		{Type: COLON, Literal: ":"},
		{Type: COLON, Literal: ":"},
		{Type: NEWLINE, Literal: "\n"},
		{Type: IDENTIFIER, Literal: "package"},
		{Type: COLON, Literal: ":"},
		{Type: COLON, Literal: ":"},
		{Type: DOT, Literal: "."},
	}

	for i, e := range seq {
		tok := ts.PeekRawTokenAt(i)
		require.Equal(t, e.Type, tok.Type, "Expected %d to be %s", i, e.Type.String())
		require.Equal(t, e.Literal, tok.Literal, "Expected %d to be %s", i, e.Literal)
	}
}

func TestPackageSeparatorHelpers(t *testing.T) {
	t.Run("default dot separator", func(t *testing.T) {
		ts := NewTokenStream("foo.bar")
		require.Equal(t, ".", ts.PackageSeparator)

		tok := ts.Emit()
		require.Equal(t, IDENTIFIER, tok.Type)
		require.Equal(t, "foo", tok.Literal)

		sep, ok := ts.TryConsumePackageSeparator()
		require.True(t, ok)
		require.Equal(t, ".", sep)

		tok = ts.Emit()
		require.Equal(t, IDENTIFIER, tok.Type)
		require.Equal(t, "bar", tok.Literal)
	})

	t.Run("custom double colon separator", func(t *testing.T) {
		ts := NewTokenStream("foo::bar")
		ts.PackageSeparator = "::"
		require.Equal(t, "::", ts.PackageSeparator)

		tok := ts.Emit()
		require.Equal(t, IDENTIFIER, tok.Type)

		sep, ok := ts.TryConsumePackageSeparator()
		require.True(t, ok)
		require.Equal(t, "::", sep)

		tok = ts.Emit()
		require.Equal(t, IDENTIFIER, tok.Type)
		require.Equal(t, "bar", tok.Literal)
	})

	t.Run("non-contiguous multi-token separator", func(t *testing.T) {
		ts := NewTokenStream("foo : : bar")
		ts.PackageSeparator = "::"

		tok := ts.Emit()
		require.Equal(t, IDENTIFIER, tok.Type)

		_, ok := ts.TryConsumePackageSeparator()
		require.False(t, ok)

		tok = ts.Emit()
		require.Equal(t, COLON, tok.Type)
	})

	t.Run("none separator", func(t *testing.T) {
		ts := NewTokenStream("foo.bar")
		ts.PackageSeparator = ""
		require.Equal(t, "", ts.PackageSeparator)

		tok := ts.Emit()
		require.Equal(t, IDENTIFIER, tok.Type)

		_, ok := ts.TryConsumePackageSeparator()
		require.False(t, ok)
	})
}

func TestStreamEmitRaw(t *testing.T) {
	ts := NewTokenStream("A /' comment '/ B")
	tokA := ts.EmitRaw()
	require.Equal(t, IDENTIFIER, tokA.Type)
	require.Equal(t, "A", tokA.Literal)

	tokC := ts.EmitRaw()
	require.Equal(t, COMMENT, tokC.Type)
	require.Equal(t, " comment ", tokC.Literal)

	tokB := ts.EmitRaw()
	require.Equal(t, IDENTIFIER, tokB.Type)
	require.Equal(t, "B", tokB.Literal)
}
