package tokenizer

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed examples/*.json
var examplesFS embed.FS

type exampleCase struct {
	Name     string          `json:"n"`
	Input    string          `json:"i"`
	Expected []expectedToken `json:"e"`
}

type expectedToken struct {
	Type    TokenType `json:"t"`
	Literal string    `json:"l"`
}

var (
	docsExamples     []exampleCase
	docsExamplesOnce sync.Once
)

func loadDocsExamples() []exampleCase {
	docsExamplesOnce.Do(func() {
		// Read all JSON files from the examples directory
		entries, err := examplesFS.ReadDir("examples")
		if err != nil {
			panic("failed to read examples directory: " + err.Error())
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}

			data, err := examplesFS.ReadFile(filepath.Join("examples", entry.Name()))
			if err != nil {
				panic("failed to read " + entry.Name() + ": " + err.Error())
			}

			var jsonExamples []exampleCase
			if err := json.Unmarshal(data, &jsonExamples); err != nil {
				panic("failed to unmarshal " + entry.Name() + ": " + err.Error())
			}
			docsExamples = append(docsExamples, jsonExamples...)
		}
	})
	return docsExamples
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

	// Convert actual tokens to expectedToken format for comparison
	actual := make([]expectedToken, len(tokens))
	for i, tok := range tokens {
		actual[i] = expectedToken{
			Type:    tok.Type,
			Literal: tok.Literal,
		}
	}

	require.Equal(t, expected, actual, "token mismatch for input:\n%s", input)
}

func TestExamplesTokenSequences(t *testing.T) {
	examples := loadDocsExamples()
	if len(examples) == 0 {
		t.Fatal("no generated docs examples found; run go generate ./pkg/tokenizer or go run ./cmd/docs-examples-gen")
	}

	for _, tc := range examples {
		t.Run(tc.Name, func(t *testing.T) {
			assertTokenSequence(t, tc.Input, tc.Expected)
		})
	}
}
