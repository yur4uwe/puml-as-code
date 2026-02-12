package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// assertTokenType checks token type with descriptive error message
func assertTokenType(t *testing.T, expected, actual TokenType) {
	t.Helper()
	require.Equal(t, expected, actual, "expected %s, got %s", expected.String(), actual.String())
}

func TestResolveUnambiguousToken_EOF(t *testing.T) {
	l := NewLexer("")
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, EOF, tok.Type)
}

func TestResolveUnambiguousToken_String(t *testing.T) {
	l := NewLexer(`"hello world"`)
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, STRING, tok.Type)
	require.Equal(t, `"hello world"`, tok.Literal)
}

func TestResolveUnambiguousToken_Brackets(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected TokenType
		literal  string
	}{
		{"left bracket", "[", LBRACKET, "["},
		{"right bracket", "]", RBRACKET, "]"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLexer(tc.input)
			tok, resolved := ResolveUnambiguousToken(l)

			require.True(t, resolved)
			assertTokenType(t, tc.expected, tok.Type)
			require.Equal(t, tc.literal, tok.Literal)
		})
	}
}

func TestResolveUnambiguousToken_Comma(t *testing.T) {
	l := NewLexer(",")
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, COMMA, tok.Type)
}

func TestResolveUnambiguousToken_Semicolon(t *testing.T) {
	l := NewLexer(";")
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, SEMICOLON, tok.Type)
}

func TestResolveUnambiguousToken_Comment(t *testing.T) {
	l := NewLexer("' this is a comment")
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, COMMENT, tok.Type)
}

func TestResolveUnambiguousToken_Generic(t *testing.T) {
	l := NewLexer("<T>")
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, GENERIC, tok.Type)
	require.Equal(t, "<T>", tok.Literal)
}

func TestResolveUnambiguousToken_Stereotype(t *testing.T) {
	l := NewLexer("<<interface>>")
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, STEREOTYPE, tok.Type)
	require.Equal(t, "<<interface>>", tok.Literal)
}

func TestResolveUnambiguousToken_UMLBounds(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{"startuml", "@startuml", START},
		{"enduml", "@enduml", END},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLexer(tc.input)
			tok, resolved := ResolveUnambiguousToken(l)

			require.True(t, resolved)
			assertTokenType(t, tc.expected, tok.Type)
		})
	}
}

func TestResolveUnambiguousToken_Unresolved(t *testing.T) {
	l := NewLexer("foo")
	_, resolved := ResolveUnambiguousToken(l)

	require.False(t, resolved, "identifiers should not be resolved as unambiguous")
}
