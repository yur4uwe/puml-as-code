package tokenizer

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
)

//go:embed examples/*.json
var examplesFS embed.FS

type exampleCase struct {
	Name     string          `json:"n"`
	Input    string          `json:"i"`
	Expected []expectedToken `json:"e"`
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

func collectTokens(lex *Lexer) []Token {
	tokens := make([]Token, 0, 32)
	for tok := lex.Emit(); tok.Type != EOF; tok = lex.Emit() {
		tokens = append(tokens, tok)
	}
	return tokens
}

func assertTokenSequence(t *testing.T, input string, expected []expectedToken) {
	t.Helper()
	l := NewLexer(input)
	actual := make([]expectedToken, 0, len(expected))

	for tok := l.Emit(); tok.Type != EOF; tok = l.Emit() {
		actual = append(actual, expectedToken{
			Type:    tok.Type,
			Literal: tok.Literal,
		})
	}

	if len(expected) != len(actual) {
		t.Errorf("token count mismatch: expected %d, got %d\n%s", len(expected), len(actual), FormatTokenDiff(expected, actual))
		return
	}

	for i := range expected {
		if expected[i].Type != actual[i].Type || expected[i].Literal != actual[i].Literal {
			t.Errorf("token mismatch at index %d\n%s", i, FormatTokenDiff(expected, actual))
			return
		}
	}
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
