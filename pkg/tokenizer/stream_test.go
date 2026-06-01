package tokenizer

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestNewTokenStream(t *testing.T) {
	ts := NewTokenStream("")
	require.NotNil(t, ts)
	tok := ts.PeekTokenAt(0)
	require.Equal(t, EOF, tok.Type)
}

func TestStreamPeekEmitConsume(t *testing.T) {
	input := "class Foo { }"
	ts := NewTokenStream(input)

	// PeekTokenAt
	tok0 := ts.PeekTokenAt(0)
	require.Equal(t, CLASS, tok0.Type)
	require.Equal(t, "class", tok0.Literal)

	tok1 := ts.PeekTokenAt(1)
	require.Equal(t, IDENTIFIER, tok1.Type)
	require.Equal(t, "Foo", tok1.Literal)

	// Assert
	require.True(t, ts.AssertType(CLASS))
	require.False(t, ts.AssertType(LBRACE))

	// Consume
	tok, ok := ts.ConsumeType(CLASS)
	require.True(t, ok)
	require.Equal(t, "class", tok.Literal)

	// Emit
	tok = ts.Emit()
	require.Equal(t, IDENTIFIER, tok.Type)
	require.Equal(t, "Foo", tok.Literal)
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

func TestStreamAssertSeq(t *testing.T) {
	ts := NewTokenStream("class Foo {")
	
	require.True(t, ts.assertSeq([]Token{
		{Type: CLASS, Literal: "class"},
		{Type: IDENTIFIER, Literal: "Foo"},
		{Type: LBRACE},
	}))
	
	require.False(t, ts.assertSeq([]Token{
		{Type: CLASS, Literal: "class"},
		{Type: IDENTIFIER, Literal: "Bar"},
	}))

	require.False(t, ts.assertSeq([]Token{
		{Type: CLASS, Literal: "class"},
		{Type: IDENTIFIER, Literal: "Foo"},
		{Type: LBRACE},
		{Type: RBRACE},
	}))
}

func TestStreamTryReadModifier(t *testing.T) {
	ts := NewTokenStream("{abstract} {static} class")
	
	mod, err := ts.TryReadModifier()
	require.NoError(t, err)
	require.Equal(t, "abstract", mod)

	mod, err = ts.TryReadModifier()
	require.NoError(t, err)
	require.Equal(t, "static", mod)

	mod, err = ts.TryReadModifier()
	require.Error(t, err)
}

func TestStreamTryReadStereotype(t *testing.T) {
	// Note: Currently fails due to internal implementation error (missing spaces)
	ts := NewTokenStream("<<stereotype>> <<foo bar>>")
	
	stereo, err := ts.TryReadStereotype()
	require.NoError(t, err)
	require.Equal(t, "stereotype", stereo)

	stereo, err = ts.TryReadStereotype()
	require.NoError(t, err)
	require.Equal(t, "foo bar", stereo)
}

func TestStreamTryReadGeneric(t *testing.T) {
	// Note: Currently fails due to internal implementation error (missing spaces)
	ts := NewTokenStream("<T> <T, U>")
	
	gen, err := ts.TryReadGeneric()
	require.NoError(t, err)
	require.Equal(t, "T", gen)

	gen, err = ts.TryReadGeneric()
	require.NoError(t, err)
	require.Equal(t, "T, U", gen)
}

func TestStreamTryReadClassSeparator(t *testing.T) {
	ts := NewTokenStream(".. separator ..\n== sep ==")
	
	sep, err := ts.TryReadClassSeparator()
	require.NoError(t, err)
	require.Equal(t, "separator", sep)

	ts.ConsumeType(NEWLINE)

	sep, err = ts.TryReadClassSeparator()
	require.NoError(t, err)
	require.Equal(t, "sep", sep)
}

func TestStreamTryReadTag(t *testing.T) {
	t.Skip("Abandoned for now per user instruction")
	ts := NewTokenStream("$tagName $another")
	
	tag, err := ts.TryReadTag()
	require.NoError(t, err)
	require.Equal(t, "tagName", tag)

	tag, err = ts.TryReadTag()
	require.NoError(t, err)
	require.Equal(t, "another", tag)
}

func TestStreamTryReadDiagramBounds(t *testing.T) {
	// Note: Currently fails due to apparent implementation bug (Assert(AT) returns false)
	ts := NewTokenStream("@startuml\n@enduml")
	
	b, err := ts.TryReadDiagramBounds()
	require.NoError(t, err)
	require.Equal(t, "startuml", b)

	ts.ConsumeType(NEWLINE)

	b, err = ts.TryReadDiagramBounds()
	require.NoError(t, err)
	require.Equal(t, "enduml", b)
}

func TestStreamReadUntilNewline(t *testing.T) {
	t.Skip("Abandoned for now per user instruction")
	ts := NewTokenStream("foo bar /' comment '/ baz\nnext")
	
	line := ts.ReadUntilNewline()
	require.Equal(t, "foo bar  baz", line)

	ts2 := NewTokenStream("foo bar /' comment '/ baz\nnext")
	lineRaw := ts2.ReadRawUntilNewline()
	require.Equal(t, "foo bar  comment  baz", lineRaw)
}

func TestStreamReadBlock(t *testing.T) {
	// Note: Currently fails because 'end' is tokenized as IDENTIFIER instead of END_BLOCK
	input := "note right of Foo\n  This is a block\n  with multiple lines\nend note"
	ts := NewTokenStream(input)
	
	ts.ReadUntilNewline()

	block, err := ts.ReadBlock(Token{Type: END_BLOCK}, Token{Type: NOTE})
	require.NoError(t, err)
	require.Contains(t, block, "This is a block")
}

func TestStreamReadMultilineComment(t *testing.T) {
	ts := NewTokenStream("something")
	c, err := ts.ReadMultilineComment()
	require.Error(t, err)
	require.Empty(t, c)
}

func TestStreamReadBetween(t *testing.T) {
	// Note: Currently fails due to internal implementation error (missing spaces)
	ts := NewTokenStream("START foo bar END")
	
	startSeq := []Token{{Type: IDENTIFIER, Literal: "START"}}
	endSeq := []Token{{Type: IDENTIFIER, Literal: "END"}}
	
	res, err := ts.readBetween(startSeq, endSeq, ts.EmitRaw)
	require.NoError(t, err)
	require.Equal(t, " foo bar ", res)
}
