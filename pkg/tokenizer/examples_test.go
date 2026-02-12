package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type exampleCase struct {
	name     string
	input    string
	expected []expectedToken
}

type expectedToken struct {
	typ     TokenType
	literal string
}

func collectTokens(input string) []Token {
	lex := NewLexer(input)
	tokens := make([]Token, 0, 32)
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}
	return tokens
}

func assertTokenSequence(t *testing.T, input string, expected []expectedToken) {
	t.Helper()
	tokens := collectTokens(input)
	require.Len(t, tokens, len(expected), "unexpected token count")

	for i, exp := range expected {
		actual := tokens[i]
		require.Equal(t, exp.typ, actual.Type, "token %d type mismatch", i)
		if exp.literal != "" {
			require.Equal(t, exp.literal, actual.Literal, "token %d literal mismatch", i)
		}
		require.NotEmpty(t, actual.Literal, "token %d has empty literal", i)
	}
}

func TestExamplesTokenSequences(t *testing.T) {
	if len(docsExamples) == 0 {
		t.Fatal("no generated docs examples found; run go generate ./pkg/tokenizer or go run ./cmd/docs-examples-gen")
	}

	for _, tc := range docsExamples {
		t.Run(tc.name, func(t *testing.T) {
			assertTokenSequence(t, tc.input, tc.expected)
		})
	}
}
