package tokenizer

import (
	"fmt"
	"strings"
)

type expectedToken struct {
	Type    TokenType `json:"t"`
	Literal string    `json:"l"`
}

// FormatTokenDiff returns a string showing expected vs actual tokens side-by-side.
func FormatTokenDiff(expected []expectedToken, actual []expectedToken) string {
	var sb strings.Builder
	sb.WriteString("\nToken Sequence Diff:\n")
	fmt.Fprintf(&sb, "%-3s | %-20s | %-20s | %s\n", "Idx", "Expected Type", "Actual Type", "Match")
	sb.WriteString(strings.Repeat("-", 70) + "\n")

	maxLen := max(len(actual), len(expected))

	for i := range maxLen {
		expStr := "<none>"
		actStr := "<none>"
		match := " "

		if i < len(expected) {
			expStr = fmt.Sprintf("%s (%q)", expected[i].Type.String(), expected[i].Literal)
		}
		if i < len(actual) {
			actStr = fmt.Sprintf("%s (%q)", actual[i].Type.String(), actual[i].Literal)
		}

		if i < len(expected) && i < len(actual) {
			if expected[i].Type == actual[i].Type && expected[i].Literal == actual[i].Literal {
				match = "✓"
			} else {
				match = "✗"
			}
		} else {
			match = "!"
		}

		fmt.Fprintf(&sb, "%-3d | %-20s | %-20s | %s\n", i, truncate(expStr, 20), truncate(actStr, 20), match)

		// If it's a mismatch, we might want to see the full literals if they were truncated
		if match != "✓" {
			if i < len(expected) {
				fmt.Fprintf(&sb, "    exp: %s %q\n", expected[i].Type, expected[i].Literal)
			}
			if i < len(actual) {
				fmt.Fprintf(&sb, "    act: %s %q\n", actual[i].Type, actual[i].Literal)
			}
		}
	}

	return sb.String()
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-3] + "..."
	}
	return s
}
