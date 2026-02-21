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

func collectTokens(lex *Lexer) []Token {
	tokens := make([]Token, 0, 32)
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}
	return tokens
}

func assertTokenSequence(t *testing.T, input string, expected []expectedToken) {
	t.Helper()
	n := len(expected)
	l := NewLexer(input)
	actual := make([]expectedToken, 0, 32)

	var i int = 0
	for tok := l.NextToken(); tok.Type != EOF; tok = l.NextToken() {
		actual = append(actual, expectedToken{
			Type:    tok.Type,
			Literal: tok.Literal,
		})
		if i < n && (tok.Type != expected[i].Type || tok.Literal != expected[i].Literal) {
			t.Errorf("token mismatch at index %d: expected %v, got %v\nneighbouting input: \n%q\nstate: %s",
				i, expected[i], tok,
				string(l.input[max(0, l.position-10):min(len(l.input), l.position+10)]),
				l.dumpState())
		}
		if i >= n {
			t.Errorf("unexpected token at index %d: %v", i, tok)
		}
		i++
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
