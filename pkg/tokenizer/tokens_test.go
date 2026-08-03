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
	require.Equal(t, `hello world`, tok.Literal)
}

func TestResolveUnambiguousToken_Physical(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{"left bracket", "[", LBRACKET},
		{"right bracket", "]", RBRACKET},
		{"left paren", "(", LPAREN},
		{"right paren", ")", RPAREN},
		{"left brace", "{", LBRACE},
		{"right brace", "}", RBRACE},
		{"left angle", "<", LANGLE},
		{"right angle", ">", RANGLE},
		{"comma", ",", COMMA},
		{"semicolon", ";", SEMICOLON},
		{"colon", ":", COLON},
		{"dot", ".", DOT},
		{"equals", "=", EQUALS},
		{"plus", "+", PLUS},
		{"hyphen", "-", DASH},
		{"tilde", "~", TILDE},
		{"hash", "#", HASH},
		{"vbar", "|", PIPE},
		{"asterisk", "*", ASTERISK},
		{"slash", "/", SLASH},
		{"backslash", "\\", BACKSLASH},
		{"caret", "^", CARET},
		{"dollar", "$", DOLLAR},
		{"percent", "%", PERCENT},
		{"at", "@", AT},
		{"exclamation", "!", EXCLAMATION},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLexer(tc.input)
			tok, resolved := ResolveUnambiguousToken(l)

			require.True(t, resolved)
			assertTokenType(t, tc.expected, tok.Type)
			require.Equal(t, tc.input, tok.Literal)
		})
	}
}

func TestResolveUnambiguousToken_Comment(t *testing.T) {
	l := NewLexer("' this is a comment")
	tok, resolved := ResolveUnambiguousToken(l)

	require.True(t, resolved)
	assertTokenType(t, COMMENT, tok.Type)
}

func TestResolveUnambiguousToken_Unresolved(t *testing.T) {
	l := NewLexer("foo")
	_, resolved := ResolveUnambiguousToken(l)

	require.False(t, resolved, "identifiers should not be resolved as unambiguous")
}

func TestResolveAmbiguousToken_Identifier(t *testing.T) {
	l := NewLexer("class")
	tok := ResolveAmbiguousToken(l)
	assertTokenType(t, IDENTIFIER, tok.Type)
	require.Equal(t, "class", tok.Literal)

	l = NewLexer("foo_bar")
	tok = ResolveAmbiguousToken(l)
	assertTokenType(t, IDENTIFIER, tok.Type)
	require.Equal(t, "foo_bar", tok.Literal)
}

func TestResolveAmbiguousToken_Number(t *testing.T) {
	l := NewLexer("123")
	tok := ResolveAmbiguousToken(l)
	assertTokenType(t, NUMBER, tok.Type)
	require.Equal(t, "123", tok.Literal)

	l = NewLexer("123.456")
	tok = ResolveAmbiguousToken(l)
	assertTokenType(t, NUMBER, tok.Type)
	require.Equal(t, "123.456", tok.Literal)

	l = NewLexer("0x123")
	tok = ResolveAmbiguousToken(l)
	assertTokenType(t, NUMBER, tok.Type)
	require.Equal(t, "0x123", tok.Literal)

	l = NewLexer("0o123")
	tok = ResolveAmbiguousToken(l)
	assertTokenType(t, NUMBER, tok.Type)
	require.Equal(t, "0o123", tok.Literal)

	l = NewLexer("0b101")
	tok = ResolveAmbiguousToken(l)
	assertTokenType(t, NUMBER, tok.Type)
	require.Equal(t, "0b101", tok.Literal)
}
