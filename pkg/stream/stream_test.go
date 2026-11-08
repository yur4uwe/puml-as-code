package stream

import (
	"testing"

	"github.com/stretchr/testify/require"

	"yur4uwe/pac/pkg/tokenizer"
)

func TestNewTokenStream(t *testing.T) {
	ts := NewTokenStream("")
	require.NotNil(t, ts)
	tok := ts.PeekRawTokenAt(0)
	require.Equal(t, tokenizer.EOF, tok.Type)
}

func TestPeekTokenAt(t *testing.T) {
	ts := NewTokenStream("package .\nset separator ::\npackage :: .")
	seq := []tokenizer.Token{
		{Type: tokenizer.IDENTIFIER, Literal: "package"},
		{Type: tokenizer.DOT, Literal: "."},
		{Type: tokenizer.NEWLINE, Literal: "\n"},
		{Type: tokenizer.IDENTIFIER, Literal: "set"},
		{Type: tokenizer.IDENTIFIER, Literal: "separator"},
		{Type: tokenizer.COLON, Literal: ":"},
		{Type: tokenizer.COLON, Literal: ":"},
		{Type: tokenizer.NEWLINE, Literal: "\n"},
		{Type: tokenizer.IDENTIFIER, Literal: "package"},
		{Type: tokenizer.COLON, Literal: ":"},
		{Type: tokenizer.COLON, Literal: ":"},
		{Type: tokenizer.DOT, Literal: "."},
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
		require.Equal(t, tokenizer.IDENTIFIER, tok.Type)
		require.Equal(t, "foo", tok.Literal)

		sep, ok := ts.TryConsumePackageSeparator()
		require.True(t, ok)
		require.Equal(t, ".", sep)

		tok = ts.Emit()
		require.Equal(t, tokenizer.IDENTIFIER, tok.Type)
		require.Equal(t, "bar", tok.Literal)
	})

	t.Run("custom double colon separator", func(t *testing.T) {
		ts := NewTokenStream("foo::bar")
		ts.PackageSeparator = "::"
		require.Equal(t, "::", ts.PackageSeparator)

		tok := ts.Emit()
		require.Equal(t, tokenizer.IDENTIFIER, tok.Type)

		sep, ok := ts.TryConsumePackageSeparator()
		require.True(t, ok)
		require.Equal(t, "::", sep)

		tok = ts.Emit()
		require.Equal(t, tokenizer.IDENTIFIER, tok.Type)
		require.Equal(t, "bar", tok.Literal)
	})

	t.Run("non-contiguous multi-token separator", func(t *testing.T) {
		ts := NewTokenStream("foo : : bar")
		ts.PackageSeparator = "::"

		tok := ts.Emit()
		require.Equal(t, tokenizer.IDENTIFIER, tok.Type)

		_, ok := ts.TryConsumePackageSeparator()
		require.False(t, ok)

		tok = ts.Emit()
		require.Equal(t, tokenizer.COLON, tok.Type)
	})

	t.Run("none separator", func(t *testing.T) {
		ts := NewTokenStream("foo.bar")
		ts.PackageSeparator = ""
		require.Equal(t, "", ts.PackageSeparator)

		tok := ts.Emit()
		require.Equal(t, tokenizer.IDENTIFIER, tok.Type)

		_, ok := ts.TryConsumePackageSeparator()
		require.False(t, ok)
	})

	t.Run("separator followed by newline", func(t *testing.T) {
		ts := NewTokenStream("foo.\nbar")
		tok := ts.Emit()
		require.Equal(t, "foo", tok.Literal)

		count, ok := ts.MatchPackageSeparatorAt(0)
		require.True(t, ok)
		require.Equal(t, 1, count)

		sep, ok := ts.TryConsumePackageSeparator()
		require.True(t, ok)
		require.Equal(t, ".", sep)
	})

	t.Run("separator at EOF", func(t *testing.T) {
		ts := NewTokenStream(".")
		count, ok := ts.MatchPackageSeparatorAt(0)
		require.True(t, ok)
		require.Equal(t, 1, count)

		sep, ok := ts.TryConsumePackageSeparator()
		require.True(t, ok)
		require.Equal(t, ".", sep)
	})

	t.Run("multi-token separator at EOF", func(t *testing.T) {
		ts := NewTokenStream("::")
		ts.PackageSeparator = "::"
		count, ok := ts.MatchPackageSeparatorAt(0)
		require.True(t, ok)
		require.Equal(t, 2, count)

		sep, ok := ts.TryConsumePackageSeparator()
		require.True(t, ok)
		require.Equal(t, "::", sep)
	})

	t.Run("IsPackageSeparator trailing tokens rejected", func(t *testing.T) {
		ts := NewTokenStream("")
		ts.PackageSeparator = "::"
		toks := []tokenizer.Token{
			{Literal: ":", Span: tokenizer.SourceSpan{Start: tokenizer.Pos{Offset: 0}, End: tokenizer.Pos{Offset: 1}}},
			{Literal: ":", Span: tokenizer.SourceSpan{Start: tokenizer.Pos{Offset: 1}, End: tokenizer.Pos{Offset: 2}}},
			{Literal: "x", Span: tokenizer.SourceSpan{Start: tokenizer.Pos{Offset: 2}, End: tokenizer.Pos{Offset: 3}}},
		}
		require.False(t, ts.IsPackageSeparator(toks))
	})
}

func TestStreamEmitRaw(t *testing.T) {
	ts := NewTokenStream("A /' comment '/ B")
	tokA := ts.EmitRaw()
	require.Equal(t, tokenizer.IDENTIFIER, tokA.Type)
	require.Equal(t, "A", tokA.Literal)

	tokC := ts.EmitRaw()
	require.Equal(t, tokenizer.COMMENT, tokC.Type)
	require.Equal(t, " comment ", tokC.Literal)

	tokB := ts.EmitRaw()
	require.Equal(t, tokenizer.IDENTIFIER, tokB.Type)
	require.Equal(t, "B", tokB.Literal)
}
