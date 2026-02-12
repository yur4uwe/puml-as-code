package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBracketedRelationshipStyles(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected Token
	}{
		{
			name:  "bold style",
			input: "-[bold]->",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "-[bold]->",
			},
		},
		{
			name:  "dashed style",
			input: "-[dashed]->",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "-[dashed]->",
			},
		},
		{
			name:  "hidden style",
			input: "-[hidden]-->",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "-[hidden]-->",
			},
		},
		{
			name:  "color",
			input: "-[#red]->",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "-[#red]->",
			},
		},
		{
			name:  "thickness",
			input: "-[thickness=8]->",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "-[thickness=8]->",
			},
		},
		{
			name:  "mixed style",
			input: "-[#blue,dotted,thickness=4]->",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "-[#blue,dotted,thickness=4]->",
			},
		},
		{
			name:  "bidirectional with style",
			input: "<-[bold]->",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "<-[bold]->",
			},
		},
		{
			name:  "plain style shorthand",
			input: "-[hidden]>",
			expected: Token{
				Type:    RELATIONSHIP,
				Literal: "-[hidden]>",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lex := NewLexer(tc.input)
			tok := lex.NextToken()

			require.Equal(t, tc.expected.Type, tok.Type, "token type mismatch")
			require.Equal(t, tc.expected.Literal, tok.Literal, "token literal mismatch")
		})
	}
}

func TestBracketedRelationshipInContext(t *testing.T) {
	input := `foo -[bold]-> bar1
foo -[#red,dashed,thickness=2]-> bar2`

	lex := NewLexer(input)
	tokens := []Token{}
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}

	// Verify we get clean relationship tokens
	require.Equal(t, IDENTIFIER, tokens[0].Type)
	require.Equal(t, "foo", tokens[0].Literal)

	require.Equal(t, RELATIONSHIP, tokens[1].Type)
	require.Equal(t, "-[bold]->", tokens[1].Literal)

	require.Equal(t, IDENTIFIER, tokens[2].Type)
	require.Equal(t, "bar1", tokens[2].Literal)

	require.Equal(t, NEWLINE, tokens[3].Type)

	require.Equal(t, IDENTIFIER, tokens[4].Type)
	require.Equal(t, "foo", tokens[4].Literal)

	require.Equal(t, RELATIONSHIP, tokens[5].Type)
	require.Equal(t, "-[#red,dashed,thickness=2]->", tokens[5].Literal)

	require.Equal(t, IDENTIFIER, tokens[6].Type)
	require.Equal(t, "bar2", tokens[6].Literal)
}
