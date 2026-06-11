package tokenizer

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestReadUntilNewline_Comments(t *testing.T) {
	input := "Red /' this is a comment '/\n"
	ts := NewTokenStream(input)
	val := ts.ReadUntilNewline()
	require.Equal(t, "Red", val)

	input2 := "Green ' the line with this comment is ignored\n"
	ts2 := NewTokenStream(input2)
	val2 := ts2.ReadUntilNewline()
	require.Equal(t, "Green", val2)
}
