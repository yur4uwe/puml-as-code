package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1_WildcardAsterisk(t *testing.T) {
	input := `remove *
restore $tag1`

	lex := NewLexer(input)
	tokens := []Token{}
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}

	require.Equal(t, REMOVE, tokens[0].Type)
	require.Equal(t, "remove", tokens[0].Literal)

	require.Equal(t, IDENTIFIER, tokens[1].Type)
	require.Equal(t, "*", tokens[1].Literal)

	require.Equal(t, NEWLINE, tokens[2].Type)

	require.Equal(t, RESTORE, tokens[3].Type)
	require.Equal(t, "restore", tokens[3].Literal)

	require.Equal(t, IDENTIFIER, tokens[4].Type)
	require.Equal(t, "$tag1", tokens[4].Literal)
}

func TestIssue2_Preprocessor(t *testing.T) {
	input := `!pragma layout smetana
!include file.puml`

	lex := NewLexer(input)
	tokens := []Token{}
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}

	require.Equal(t, PREPROCESSOR, tokens[0].Type)
	require.Equal(t, "!pragma", tokens[0].Literal)

	require.Equal(t, IDENTIFIER, tokens[1].Type)
	require.Equal(t, "layout", tokens[1].Literal)

	require.Equal(t, IDENTIFIER, tokens[2].Type)
	require.Equal(t, "smetana", tokens[2].Literal)

	require.Equal(t, NEWLINE, tokens[3].Type)

	require.Equal(t, PREPROCESSOR, tokens[4].Type)
	require.Equal(t, "!include", tokens[4].Literal)

	require.Equal(t, IDENTIFIER, tokens[5].Type)
	require.Equal(t, "file", tokens[5].Literal)
}

func TestIssue3_NoteHTMLTags(t *testing.T) {
	input := `note top of Foo
  <b>bold text</b>
  </color>
end note`

	lex := NewLexer(input)
	tokens := []Token{}
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}

	// Debug output
	t.Logf("Got %d tokens:", len(tokens))
	for i, tok := range tokens {
		t.Logf("  [%d] %s: %q", i, tok.Type.String(), tok.Literal)
	}

	require.Equal(t, NOTE, tokens[0].Type)
	require.Equal(t, NOTE_DIRECTION, tokens[1].Type)
	require.Equal(t, "top", tokens[1].Literal)
	require.Equal(t, NOTE_POSITION, tokens[2].Type)
	require.Equal(t, "of", tokens[2].Literal)
	require.Equal(t, IDENTIFIER, tokens[3].Type)
	require.Equal(t, "Foo", tokens[3].Literal)
	require.Equal(t, NEWLINE, tokens[4].Type)

	// All note content should be in a single IDENTIFIER token
	require.Equal(t, IDENTIFIER, tokens[5].Type)
	expectedContent := "<b>bold text</b>\n</color>"
	require.Equal(t, expectedContent, tokens[5].Literal, "All multiline note content should be in single token")

	require.Equal(t, END_BLOCK, tokens[6].Type)
	require.Equal(t, "end note", tokens[6].Literal)
}

func TestIssue4_SingleCapitalsInNamespace(t *testing.T) {
	input := `class A.B.C.D.Z {
}`

	lex := NewLexer(input)
	tokens := []Token{}
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}

	require.Equal(t, CLASS, tokens[0].Type)

	require.Equal(t, IDENTIFIER, tokens[1].Type)
	require.Equal(t, "A", tokens[1].Literal)

	require.Equal(t, SEPARATOR, tokens[2].Type)
	require.Equal(t, ".", tokens[2].Literal)

	require.Equal(t, IDENTIFIER, tokens[3].Type)
	require.Equal(t, "B", tokens[3].Literal)

	require.Equal(t, SEPARATOR, tokens[4].Type)
	require.Equal(t, ".", tokens[4].Literal)

	require.Equal(t, IDENTIFIER, tokens[5].Type)
	require.Equal(t, "C", tokens[5].Literal)

	require.Equal(t, SEPARATOR, tokens[6].Type)
	require.Equal(t, ".", tokens[6].Literal)

	require.Equal(t, IDENTIFIER, tokens[7].Type)
	require.Equal(t, "D", tokens[7].Literal)

	require.Equal(t, SEPARATOR, tokens[8].Type)
	require.Equal(t, ".", tokens[8].Literal)

	require.Equal(t, IDENTIFIER, tokens[9].Type)
	require.Equal(t, "Z", tokens[9].Literal)

	require.Equal(t, LBRACE, tokens[10].Type)
}

func TestIssue5_AtPrefixIdentifiers(t *testing.T) {
	input := `hide @unlinked
@startuml
@enduml`

	lex := NewLexer(input)
	tokens := []Token{}
	for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
		tokens = append(tokens, tok)
	}

	require.Equal(t, HIDE, tokens[0].Type)

	require.Equal(t, IDENTIFIER, tokens[1].Type)
	require.Equal(t, "@unlinked", tokens[1].Literal)

	require.Equal(t, NEWLINE, tokens[2].Type)

	require.Equal(t, START, tokens[3].Type)
	require.Equal(t, "startuml", tokens[3].Literal)

	require.Equal(t, NEWLINE, tokens[4].Type)

	require.Equal(t, END, tokens[5].Type)
	require.Equal(t, "enduml", tokens[5].Literal)
}
