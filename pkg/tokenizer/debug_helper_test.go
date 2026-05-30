package tokenizer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatTokenDiff_Alignment(t *testing.T) {
	exp := []expectedToken{
		{Type: IDENTIFIER, Literal: "A"},
		{Type: IDENTIFIER, Literal: "B"},
		{Type: IDENTIFIER, Literal: "C"},
		{Type: IDENTIFIER, Literal: "D"},
		{Type: IDENTIFIER, Literal: "E"},
	}

	act := []expectedToken{
		{Type: IDENTIFIER, Literal: "A"},
		{Type: IDENTIFIER, Literal: "X"}, // Substitution (Idx 1)
		{Type: IDENTIFIER, Literal: "C"}, // Match (Idx 2)
		{Type: IDENTIFIER, Literal: "Y"}, // Insertion (+3)
		{Type: IDENTIFIER, Literal: "D"}, // Match (Idx 3)
		// Missing 'E' (Idx 4)
	}

	diff := FormatTokenDiff(exp, act)

	// Verify the output contains our markers
	require.Contains(t, diff, "✓", "Should contain match markers")
	require.Contains(t, diff, "✗", "Should contain substitution markers")
	require.Contains(t, diff, "+", "Should contain insertion markers")
	require.Contains(t, diff, "-", "Should contain deletion markers")

	// Verify specific alignments we expect from the logic
	require.True(t, strings.Contains(diff, "IDENTIFIER (\"B\")     | IDENTIFIER (\"X\")     | ✗"), "Should show substitution for B -> X")
	require.True(t, strings.Contains(diff, "<none>               | IDENTIFIER (\"Y\")     | +"), "Should show insertion for Y")
	require.True(t, strings.Contains(diff, "IDENTIFIER (\"E\")     | <none>               | -"), "Should show deletion for E")

	// For manual inspection during development
	t.Logf("Aligned Diff Output:\n%s", diff)
}
