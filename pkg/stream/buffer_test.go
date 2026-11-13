package stream

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"yur4uwe/pac/pkg/tokenizer"
)

func TestLookaheadRemainsStableThroughEOF(t *testing.T) {
	ts := NewTokenStream("A ' a comment\nB")
	eof := ts.PeekRawTokenAt(200)
	require.Equal(t, tokenizer.EOF, eof.Type)
	for range 100 {
		require.Equal(t, eof, ts.PeekRawTokenAt(200))
	}
	require.Equal(t, "A", ts.Emit().Literal)
	require.Equal(t, "B", ts.PeekTokenAt(1).Literal)
	require.Equal(t, tokenizer.NEWLINE, ts.Emit().Type)
	require.Len(t, ts.DumpCollectedTrivia(), 1)
	require.Equal(t, "B", ts.Emit().Literal)
	require.Equal(t, eof, ts.Emit())
}

func TestLookaheadOrderAfterSlidingWindow(t *testing.T) {
	var input strings.Builder
	for i := range 1000 {
		fmt.Fprintf(&input, "item%d ", i)
	}
	ts := NewTokenStream(input.String())
	for i := range 1000 {
		next := ts.PeekRawTokenAt(12)
		if i+12 < 1000 {
			require.Equal(t, fmt.Sprintf("item%d", i+12), next.Literal)
		} else {
			require.Equal(t, tokenizer.EOF, next.Type)
		}
		require.Equal(t, fmt.Sprintf("item%d", i), ts.Emit().Literal)
	}
}

func BenchmarkEmitWithLookahead(b *testing.B) {
	src := strings.Repeat("class User {\n+id string\n}\n", 128)
	for b.Loop() {
		ts := NewTokenStream(src)
		for ts.PeekTokenAt(3).Type != tokenizer.EOF {
			ts.Emit()
		}
	}
}
