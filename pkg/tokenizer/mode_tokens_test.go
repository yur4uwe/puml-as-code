package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveDefaultModeToken_Newline(t *testing.T) {
	l := NewLexer("\n")
	tok, resolved := resolveDefaultModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, NEWLINE, tok.Type)
	require.Equal(t, "\n", tok.Literal)
}

func TestResolveDefaultModeToken_LollipopInterface(t *testing.T) {
	l := NewLexer("()--")
	tok, resolved := resolveDefaultModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, RELATIONSHIP, tok.Type)
	require.Equal(t, "()--", tok.Literal)
}

func TestResolveDefaultModeToken_ParenthesisNotRelation(t *testing.T) {
	l := NewLexer("(foo")
	tok, resolved := resolveDefaultModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, LPAREN, tok.Type)
	require.Equal(t, "(", tok.Literal)
}

func TestResolveClassDefModeToken_Newline(t *testing.T) {
	l := NewLexer("\n")
	l.PushMode(MODE_CLASS_DEF)
	tok, resolved := resolveClassDefModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, NEWLINE, tok.Type)
	require.Equal(t, MODE_DEFAULT, l.CurrentMode(), "should pop MODE_CLASS_DEF on newline")
}

func TestResolveClassDefModeToken_OpenBrace(t *testing.T) {
	l := NewLexer("{")
	l.PushMode(MODE_CLASS_DEF)
	tok, resolved := resolveClassDefModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, LBRACE, tok.Type)
	require.Equal(t, MODE_CLASS, l.CurrentMode(), "should transition to MODE_CLASS")
}

func TestResolveClassDefModeToken_Extends(t *testing.T) {
	l := NewLexer("extends Foo")
	l.PushMode(MODE_CLASS_DEF)
	tok, resolved := resolveClassDefModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, RELATIONSHIP, tok.Type)
	require.Equal(t, "extends", tok.Literal)
}

func TestResolveClassDefModeToken_Implements(t *testing.T) {
	l := NewLexer("implements IFoo")
	l.PushMode(MODE_CLASS_DEF)
	tok, resolved := resolveClassDefModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, RELATIONSHIP, tok.Type)
	require.Equal(t, "implements", tok.Literal)
}

func TestResolveClassModeToken_Newline(t *testing.T) {
	l := NewLexer("\n")
	l.PushMode(MODE_CLASS)
	tok, resolved := resolveClassModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, NEWLINE, tok.Type)
	require.Equal(t, MODE_CLASS, l.CurrentMode(), "should stay in MODE_CLASS on newline")
}

func TestResolveClassModeToken_CloseBrace(t *testing.T) {
	// needs to have a character before the close brace to pop the mode
	l := NewLexer(" }")
	l.PushMode(MODE_CLASS)
	l.jumpToPosition(1)
	tok, resolved := resolveClassModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, RBRACE, tok.Type)
	require.Equal(t, MODE_DEFAULT, l.CurrentMode(), "should pop MODE_CLASS on close brace")
}

func TestResolveClassModeToken_Separator(t *testing.T) {
	l := NewLexer("--")
	l.PushMode(MODE_CLASS)
	tok, resolved := resolveClassModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, SEPARATOR, tok.Type)
	require.Equal(t, "--", tok.Literal)
	require.Equal(t, MODE_LABEL, l.CurrentMode(), "should push MODE_LABEL after separator")
}

func TestResolveClassModeToken_Visibility(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"public", "+field", "+"},
		{"private", "-field", "-"},
		{"protected", "#field", "#"},
		{"package", "~field", "~"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLexer(tc.input)
			l.PushMode(MODE_CLASS)
			tok, resolved := resolveClassModeToken(l)

			require.True(t, resolved)
			assertTokenType(t, VISIBILITY, tok.Type)
			require.Equal(t, tc.expected, tok.Literal)
		})
	}
}

func TestResolveClassModeToken_Identifier(t *testing.T) {
	l := NewLexer("field")
	l.PushMode(MODE_CLASS)
	tok, resolved := resolveClassModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, IDENTIFIER, tok.Type)
	require.Equal(t, "field", tok.Literal)
}

func TestResolveLabelModeToken_Newline(t *testing.T) {
	l := NewLexer("\n")
	l.PushMode(MODE_LABEL)
	tok, resolved := resolveLabelModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, NEWLINE, tok.Type)
	require.Equal(t, MODE_DEFAULT, l.CurrentMode(), "should pop MODE_LABEL on newline")
}

func TestResolveLabelModeToken_Separator(t *testing.T) {
	l := NewLexer("--")
	l.PushMode(MODE_LABEL)
	tok, resolved := resolveLabelModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, SEPARATOR, tok.Type)
	require.Equal(t, "--", tok.Literal)
	require.Equal(t, MODE_DEFAULT, l.CurrentMode(), "should pop MODE_LABEL on separator")
}

func TestResolveLabelModeToken_Identifier(t *testing.T) {
	l := NewLexer("some label text")
	l.PushMode(MODE_LABEL)
	tok, resolved := resolveLabelModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, IDENTIFIER, tok.Type)
	require.NotEmpty(t, tok.Literal)
}

func TestResolveNoteModeToken_NewlineSingleLine(t *testing.T) {
	l := NewLexer("\n")
	l.PushMode(MODE_NOTE)
	l.isMultilineNote = false
	tok, resolved := resolveNoteModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, NEWLINE, tok.Type)
	require.Equal(t, MODE_DEFAULT, l.CurrentMode(), "should pop MODE_NOTE on newline for single-line note")
}

func TestResolveNoteModeToken_NewlineMultiLine(t *testing.T) {
	l := NewLexer("\n")
	l.PushMode(MODE_NOTE)
	l.isMultilineNote = true
	tok, resolved := resolveNoteModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, NEWLINE, tok.Type)
	require.Equal(t, MODE_NOTE, l.CurrentMode(), "should stay in MODE_NOTE on newline for multi-line note")
}

func TestResolveNoteModeToken_Colon(t *testing.T) {
	l := NewLexer(": note text")
	l.PushMode(MODE_NOTE)
	tok, resolved := resolveNoteModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, COLON, tok.Type)
	require.Equal(t, MODE_LABEL, l.CurrentMode(), "should transition to MODE_LABEL after colon")
	require.False(t, l.isMultilineNote, "should mark as single-line note")
}

func TestResolveNoteModeToken_EndBlock(t *testing.T) {
	l := NewLexer("end note")
	l.PushMode(MODE_NOTE)
	l.isMultilineNote = true
	tok, resolved := resolveNoteModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, END_BLOCK, tok.Type)
	require.Equal(t, "end note", tok.Literal)
	require.Equal(t, MODE_DEFAULT, l.CurrentMode(), "should pop MODE_NOTE on end note")
	require.False(t, l.isMultilineNote, "should clear multiline flag")
}

func TestResolveNoteModeToken_NotePosition(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"of", "of Foo"},
		{"on", "on link"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLexer(tc.input)
			l.PushMode(MODE_NOTE)
			l.isMultilineNote = false
			tok, resolved := resolveNoteModeToken(l)

			require.True(t, resolved)
			assertTokenType(t, NOTE_POSITION, tok.Type)
			require.True(t, l.isMultilineNote, "should mark as multi-line note")
		})
	}
}

func TestResolveNoteModeToken_Content(t *testing.T) {
	l := NewLexer("this is note content\n")
	l.PushMode(MODE_NOTE)
	tok, resolved := resolveNoteModeToken(l)

	require.True(t, resolved)
	assertTokenType(t, IDENTIFIER, tok.Type)
	require.Equal(t, "this is note content", tok.Literal)
}
