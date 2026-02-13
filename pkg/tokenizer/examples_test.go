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
	name     string
	input    string
	expected []expectedToken
}

type expectedToken struct {
	typ     TokenType
	literal string
}

type jsonExample struct {
	Name     string      `json:"n"`
	Input    string      `json:"i"`
	Expected []jsonToken `json:"e"`
}

type jsonToken struct {
	Type    string `json:"t"`
	Literal string `json:"l"`
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

		var allExamples []exampleCase

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}

			data, err := examplesFS.ReadFile(filepath.Join("examples", entry.Name()))
			if err != nil {
				panic("failed to read " + entry.Name() + ": " + err.Error())
			}

			var jsonExamples []jsonExample
			if err := json.Unmarshal(data, &jsonExamples); err != nil {
				panic("failed to unmarshal " + entry.Name() + ": " + err.Error())
			}

			// Convert JSON examples to test cases
			for _, je := range jsonExamples {
				expectedTokens := make([]expectedToken, len(je.Expected))
				for j, jt := range je.Expected {
					expectedTokens[j] = expectedToken{
						typ:     tokenTypeFromString(jt.Type),
						literal: jt.Literal,
					}
				}
				allExamples = append(allExamples, exampleCase{
					name:     je.Name,
					input:    je.Input,
					expected: expectedTokens,
				})
			}
		}

		docsExamples = allExamples
	})
	return docsExamples
}

func tokenTypeFromString(s string) TokenType {
	switch s {
	case "ILLEGAL":
		return ILLEGAL
	case "EOF":
		return EOF
	case "NEWLINE":
		return NEWLINE
	case "IDENTIFIER":
		return IDENTIFIER
	case "STRING":
		return STRING
	case "NUMBER":
		return NUMBER
	case "CLASS":
		return CLASS
	case "ABSTRACT":
		return ABSTRACT
	case "STRUCT":
		return STRUCT
	case "INTERFACE":
		return INTERFACE
	case "ENUM":
		return ENUM
	case "PACKAGE":
		return PACKAGE
	case "ANNOTATION":
		return ANNOTATION
	case "NOTE":
		return NOTE
	case "STEREOTYPE":
		return STEREOTYPE
	case "RECORD":
		return RECORD
	case "DATACLASS":
		return DATACLASS
	case "EXCEPTION":
		return EXCEPTION
	case "PROTOCOL":
		return PROTOCOL
	case "HIDE":
		return HIDE
	case "SHOW":
		return SHOW
	case "REMOVE":
		return REMOVE
	case "RESTORE":
		return RESTORE
	case "SKINPARAM":
		return SKINPARAM
	case "SET_PROPERTY":
		return SET_PROPERTY
	case "TOGETHER":
		return TOGETHER
	case "END_BLOCK":
		return END_BLOCK
	case "GENERIC":
		return GENERIC
	case "NOTE_DIRECTION":
		return NOTE_DIRECTION
	case "NOTE_POSITION":
		return NOTE_POSITION
	case "LBRACE":
		return LBRACE
	case "RBRACE":
		return RBRACE
	case "LPAREN":
		return LPAREN
	case "RPAREN":
		return RPAREN
	case "LBRACKET":
		return LBRACKET
	case "RBRACKET":
		return RBRACKET
	case "SEMICOLON":
		return SEMICOLON
	case "COLON":
		return COLON
	case "COMMA":
		return COMMA
	case "VISIBILITY":
		return VISIBILITY
	case "MODIFIER":
		return MODIFIER
	case "RELATIONSHIP":
		return RELATIONSHIP
	case "COMMENT":
		return COMMENT
	case "SEPARATOR":
		return SEPARATOR
	case "ALIAS":
		return ALIAS
	case "START":
		return START
	case "END":
		return END
	case "PREPROCESSOR":
		return PREPROCESSOR
	default:
		return ILLEGAL
	}
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
		require.Equalf(t, exp.typ, actual.Type, "token type mismatch: expected %v, got %v", exp.typ, actual.Type)
		if exp.literal != "" {
			require.Equalf(t, exp.literal, actual.Literal, "token %d literal mismatch", i)
		}
		require.NotEmptyf(t, actual.Literal, "token %d has empty literal", i)
	}
}

func TestExamplesTokenSequences(t *testing.T) {
	examples := loadDocsExamples()
	if len(examples) == 0 {
		t.Fatal("no generated docs examples found; run go generate ./pkg/tokenizer or go run ./cmd/docs-examples-gen")
	}

	for _, tc := range examples {
		t.Run(tc.name, func(t *testing.T) {
			assertTokenSequence(t, tc.input, tc.expected)
		})
	}
}
