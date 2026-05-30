package tokenizer

import (
	"fmt"
	"strings"
)

type expectedToken struct {
	Type    TokenType `json:"t"`
	Literal string    `json:"l"`
}

type diffOp int

const (
	opMatch diffOp = iota
	opSubstitute
	opInsert
	opDelete
)

type diffItem struct {
	op     diffOp
	expIdx int // -1 if insert
	actIdx int // -1 if delete
}

func computeTokenDiff(expected, actual []expectedToken) []diffItem {
	n := len(expected)
	m := len(actual)

	// dp[i][j] stores the minimum edit distance
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 0; i <= n; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 2
			if expected[i-1].Type == actual[j-1].Type && expected[i-1].Literal == actual[j-1].Literal {
				cost = 0
			}

			replaceCost := dp[i-1][j-1] + cost
			deleteCost := dp[i-1][j] + 1
			insertCost := dp[i][j-1] + 1

			minCost := replaceCost
			if deleteCost < minCost {
				minCost = deleteCost
			}
			if insertCost < minCost {
				minCost = insertCost
			}

			dp[i][j] = minCost
		}
	}

	// Backtrack to find alignment
	var alignment []diffItem
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 {
			cost := 2
			if expected[i-1].Type == actual[j-1].Type && expected[i-1].Literal == actual[j-1].Literal {
				cost = 0
			}

			if dp[i][j] == dp[i-1][j-1]+cost {
				if cost == 0 {
					alignment = append(alignment, diffItem{opMatch, i - 1, j - 1})
				} else {
					alignment = append(alignment, diffItem{opSubstitute, i - 1, j - 1})
				}
				i--
				j--
				continue
			}
		}

		if i > 0 && dp[i][j] == dp[i-1][j]+1 {
			alignment = append(alignment, diffItem{opDelete, i - 1, -1})
			i--
		} else if j > 0 && dp[i][j] == dp[i][j-1]+1 {
			alignment = append(alignment, diffItem{opInsert, -1, j - 1})
			j--
		}
	}

	// Reverse alignment
	for l, r := 0, len(alignment)-1; l < r; l, r = l+1, r-1 {
		alignment[l], alignment[r] = alignment[r], alignment[l]
	}

	return alignment
}

// FormatTokenDiff returns a string showing expected vs actual tokens side-by-side.
func FormatTokenDiff(expected []expectedToken, actual []expectedToken) string {
	alignment := computeTokenDiff(expected, actual)

	var sb strings.Builder
	sb.WriteString("\nToken Sequence Diff:\n")
	fmt.Fprintf(&sb, "%-5s | %-20s | %-20s | %s\n", "Idx", "Expected Type", "Actual Type", "Match")
	sb.WriteString(strings.Repeat("-", 70) + "\n")

	for _, item := range alignment {
		expStr := "<none>"
		actStr := "<none>"
		match := " "
		idxStr := ""

		switch item.op {
		case opMatch:
			expStr = fmt.Sprintf("%s (%q)", expected[item.expIdx].Type.String(), expected[item.expIdx].Literal)
			actStr = fmt.Sprintf("%s (%q)", actual[item.actIdx].Type.String(), actual[item.actIdx].Literal)
			match = "✓"
			idxStr = fmt.Sprintf("%d", item.expIdx)
		case opSubstitute:
			expStr = fmt.Sprintf("%s (%q)", expected[item.expIdx].Type.String(), expected[item.expIdx].Literal)
			actStr = fmt.Sprintf("%s (%q)", actual[item.actIdx].Type.String(), actual[item.actIdx].Literal)
			match = "✗"
			idxStr = fmt.Sprintf("%d", item.expIdx)
		case opDelete:
			expStr = fmt.Sprintf("%s (%q)", expected[item.expIdx].Type.String(), expected[item.expIdx].Literal)
			match = "-"
			idxStr = fmt.Sprintf("%d", item.expIdx)
		case opInsert:
			actStr = fmt.Sprintf("%s (%q)", actual[item.actIdx].Type.String(), actual[item.actIdx].Literal)
			match = "+"
			idxStr = fmt.Sprintf("+%d", item.actIdx)
		}

		fmt.Fprintf(&sb, "%-5s | %-20s | %-20s | %s\n", idxStr, truncate(expStr, 20), truncate(actStr, 20), match)

		if item.op == opSubstitute {
			fmt.Fprintf(&sb, "        exp: %s %q\n", expected[item.expIdx].Type, expected[item.expIdx].Literal)
			fmt.Fprintf(&sb, "        act: %s %q\n", actual[item.actIdx].Type, actual[item.actIdx].Literal)
		} else if item.op == opDelete {
			fmt.Fprintf(&sb, "        missing token\n")
		} else if item.op == opInsert {
			fmt.Fprintf(&sb, "        unexpected token\n")
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

func (l *Lexer) DebugState() {
	fmt.Printf("DEBUG: pos=%d, ch=%q, sep=%q, isDefault=%v, expecting=%v\n",
		l.position, l.ch, string(l.packageSeparator), l.isDefaultSeparator, l.expectingSeparatorValue)
}
