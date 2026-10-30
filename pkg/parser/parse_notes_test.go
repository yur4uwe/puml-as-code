package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/internal/testutil"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

type expectedNoteBodyState byte

const (
	expectSingleLine expectedNoteBodyState = iota
	expectMultilineClosed
	expectMultilineUnclosed
)

func TestParseNote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *ast.Note
	}{
		// 1. Relative/Directed Notes
		{
			name:  "relative note left with colon",
			input: "note left : some note text\n",
			want: &ast.Note{
				Kind:      ast.NoteTargeted,
				Text:      "some note text",
				Direction: ast.DirectionLeft,
			},
		},
		{
			name:  "relative note right of target with colon",
			input: "note right of MyClass : some note text\n",
			want: &ast.Note{
				Kind:      ast.NoteTargeted,
				Text:      "some note text",
				Direction: ast.DirectionRight,
				Target:    identRefPtr("MyClass"),
			},
		},
		{
			name:  "relative note top of target with package path and colon",
			input: "note top of pkg.MyClass : some note text\n",
			want: &ast.Note{
				Kind:      ast.NoteTargeted,
				Text:      "some note text",
				Direction: ast.DirectionTop,
				Target:    identRefPtr("MyClass", "pkg"),
			},
		},
		{
			name:  "relative note top of target with color and colon",
			input: "note top of MyClass #green : some note text\n",
			want: &ast.Note{
				Kind:      ast.NoteTargeted,
				Text:      "some note text",
				Direction: ast.DirectionTop,
				Target:    identRefPtr("MyClass"),
				Color:     "#green",
			},
		},
		{
			name:  "relative note left with multiline",
			input: "note left\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:      ast.NoteTargeted,
				Text:      "some note text",
				Direction: ast.DirectionLeft,
			},
		},
		{
			name:  "relative note left of target multiline",
			input: "note left of MyClass\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:      ast.NoteTargeted,
				Text:      "some note text",
				Direction: ast.DirectionLeft,
				Target:    identRefPtr("MyClass"),
			},
		},
		{
			name:  "relative note right of target with color multiline",
			input: "note right of MyClass #blue\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:      ast.NoteTargeted,
				Text:      "some note text",
				Direction: ast.DirectionRight,
				Target:    identRefPtr("MyClass"),
				Color:     "#blue",
			},
		},

		// 2. Inline Alias Notes
		{
			name:  "inline alias note",
			input: `note "some note text" as N1` + "\n",
			want: &ast.Note{
				Kind:       ast.NoteAlias,
				Text:       "some note text",
				Identifier: identRef("N1"),
			},
		},
		{
			name:  "inline alias note with color",
			input: `note "some note text" as N1 #blue` + "\n",
			want: &ast.Note{
				Kind:       ast.NoteAlias,
				Text:       "some note text",
				Identifier: identRef("N1"),
				Color:      "#blue",
			},
		},
		{
			name:  "inline alias note with qualified identifier",
			input: `note "some note text" as pkg.N1` + "\n",
			want: &ast.Note{
				Kind:       ast.NoteAlias,
				Text:       "some note text",
				Identifier: identRef("N1", "pkg"),
			},
		},

		// 3. Multiline Alias Notes
		{
			name:  "multiline alias note with newline",
			input: "note as N1\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:       ast.NoteAlias,
				Text:       "some note text",
				Identifier: identRef("N1"),
			},
		},
		{
			name:  "multiline alias note with color",
			input: "note as N1 #red\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:       ast.NoteAlias,
				Text:       "some note text",
				Identifier: identRef("N1"),
				Color:      "#red",
			},
		},
		{
			name:  "multiline alias note with qualified identifier",
			input: "note as pkg.N1\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:       ast.NoteAlias,
				Text:       "some note text",
				Identifier: identRef("N1", "pkg"),
			},
		},

		// 4. Link Notes
		{
			name:  "link note on link with colon",
			input: "note on link : some note text\n",
			want: &ast.Note{
				Kind:   ast.NoteLink,
				Text:   "some note text",
				Target: identRefPtr("link"),
			},
		},
		{
			name:  "link note on link with color and colon",
			input: "note on link #yellow : some note text\n",
			want: &ast.Note{
				Kind:   ast.NoteLink,
				Text:   "some note text",
				Target: identRefPtr("link"),
				Color:  "#yellow",
			},
		},
		{
			name:  "link note on link multiline",
			input: "note on link\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:   ast.NoteLink,
				Text:   "some note text",
				Target: identRefPtr("link"),
			},
		},
		{
			name:  "link note on link multiline with color",
			input: "note on link #cyan\nsome note text\nend note\n",
			want: &ast.Note{
				Kind:   ast.NoteLink,
				Text:   "some note text",
				Target: identRefPtr("link"),
				Color:  "#cyan",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			require.Equal(t, keyword.Note, keyword.Classify(tok.Literal), "First token must be 'note' for test input: %q", tc.input)

			got, err := p.parseNote(tok)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.False(t, p.ast.HasErrors(), "expected no diagnostics for valid note definition")
			testutil.SemanticEqual(t, tc.want, got)
		})
	}
}

func TestParseNote_Errors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantText  string
		wantDiags []expectedDiag
	}{
		// 1. Directed note errors
		{
			name:     "relative note invalid identifier after direction",
			input:    "note left invalid : text\n",
			wantText: "text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Unexpected identifier after direction",
				},
			},
		},
		{
			name:     "relative note expected identifier for target",
			input:    "note left of : text\n",
			wantText: "text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected identifier for target",
				},
			},
		},
		{
			name:     "relative note unexpected identifier for link target",
			input:    "note left on MyClass : text\n",
			wantText: "text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Unexpected identifier for a note link target",
				},
			},
		},
		{
			name:     "relative note warning on link direction ignored",
			input:    "note left on link : text\n",
			wantText: "text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityWarning,
					Msg:      "direction \"left\" is ignored for note on links",
				},
			},
		},

		// 2. Inline alias note errors
		{
			name:     "inline alias note missing as",
			input:    "note \"some note text\" N1\n",
			wantText: "some note text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "expected alias keyword after note text",
				},
			},
		},
		{
			name:     "inline alias note missing identifier",
			input:    "note \"some note text\" as\n",
			wantText: "some note text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected identifier after alias keyword",
				},
			},
		},
		{
			name:     "inline alias note unexpected tokens after alias",
			input:    "note \"some note text\" as N1 extra tokens\n",
			wantText: "some note text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Unexpected tokens after inline alias note",
				},
			},
		},

		// 3. Multiline alias note errors
		{
			name:     "multiline alias note missing identifier with closed body",
			input:    "note as\nsome note text\nend note\n",
			wantText: "some note text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected identifier after alias keyword",
				},
			},
		},
		{
			name:  "multiline alias note missing identifier and unclosed body",
			input: "note as\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected identifier after alias keyword",
				},
				{
					Severity: ast.SeverityError,
					Msg:      "unclosed note: expected end note",
				},
			},
		},
		{
			name:     "multiline alias note with colon (invalid newline) and closed body",
			input:    "note as N1 : some note text\nbody\nend note\n",
			wantText: "body",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected newline after alias keyword",
				},
			},
		},
		{
			name:     "multiline alias note with colon and color (invalid) and closed body",
			input:    "note as N1 #red : some note text\nbody\nend note\n",
			wantText: "body",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected newline after alias keyword",
				},
			},
		},

		// 4. Link note errors
		{
			name:     "unexpected identifier after note (link note invalid position)",
			input:    "note of link : text\n",
			wantText: "text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Unexpected identifier after 'note'",
				},
			},
		},
		{
			name:     "link note expected link after note on with colon",
			input:    "note on invalid : text\n",
			wantText: "text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected 'link' after 'note on'",
				},
			},
		},
		{
			name:     "link note expected link after note on with block",
			input:    "note on invalid\ntext\nend note\n",
			wantText: "text",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected 'link' after 'note on'",
				},
			},
		},
		{
			name:  "link note missing target and unclosed body",
			input: "note on\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected 'link' after 'note on'",
				},
				{
					Severity: ast.SeverityError,
					Msg:      "unclosed note: expected end note",
				},
			},
		},

		// 5. Entry point & Body errors
		{
			name:  "parseNote unexpected starting keyword",
			input: "note class : text\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "expected direction (left, right etc.), position or alias (as) after 'note', got class",
				},
			},
		},
		{
			name:  "parseNote unexpected starting token type",
			input: "note 123 : text\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unexpected token in note definition",
				},
			},
		},
		{
			name:  "parseNoteBody unexpected body definition token",
			input: "note left @invalid\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected ':' or newline after note definition",
				},
			},
		},
		{
			name:  "parseNoteBody unclosed multiline block",
			input: "note left\nsome unclosed text\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unclosed note: expected end note",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			require.Equal(t, keyword.Note, keyword.Classify(tok.Literal))

			got, err := p.parseNote(tok)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.True(t, p.ast.HasErrors(), "expected parser to record diagnostics on bad syntax")
			assertDiagsEqual(t, tc.wantDiags, p.ast.Diagnostics, got)
			if tc.wantText != "" {
				require.Equal(t, tc.wantText, got.Text)
			}
		})
	}
}

func TestParseNote_Recovery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		bs       expectedNoteBodyState
		wantText string
	}{
		{
			name:     "restore on invalid identifier after direction with colon",
			input:    "note left invalid : text\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "text",
		},
		{
			name:     "restore on invalid identifier after direction with multiline block",
			input:    "note left invalid\nsome multiline text\nend note\nclass NextClass",
			bs:       expectMultilineClosed,
			wantText: "some multiline text",
		},
		{
			name:     "restore on missing target identifier after of with colon",
			input:    "note left of : text\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "text",
		},
		{
			name:     "restore on missing target identifier after of with multiline block",
			input:    "note left of\nsome multiline text\nend note\nclass NextClass",
			bs:       expectMultilineClosed,
			wantText: "some multiline text",
		},
		{
			name:     "restore on unexpected identifier for link target with colon",
			input:    "note left on MyClass : text\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "text",
		},
		{
			name:     "restore on unexpected identifier for link target with multiline block",
			input:    "note left on MyClass\nsome multiline text\nend note\nclass NextClass",
			bs:       expectMultilineClosed,
			wantText: "some multiline text",
		},
		{
			name:     "restore on invalid alias for inline note",
			input:    "note \"some text\" as 123\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "some text",
		},
		{
			name:     "restore on missing as keyword for inline note",
			input:    "note \"some text\" N1\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "some text",
		},
		{
			name:     "restore on extra tokens after inline alias note",
			input:    "note \"some text\" as N1 extra tokens\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "some text",
		},
		{
			name:     "restore on missing identifier for multiline alias note with block",
			input:    "note as 123\nsome note text\nend note\nclass NextClass",
			bs:       expectMultilineClosed,
			wantText: "some note text",
		},
		{
			name:     "restore on unexpected tokens before newline for multiline alias note",
			input:    "note as N1 : invalid\nsome note text\nend note\nclass NextClass",
			bs:       expectMultilineClosed,
			wantText: "some note text",
		},
		{
			name:     "restore on expected link after note on with colon",
			input:    "note on invalid : text\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "text",
		},
		{
			name:     "restore on expected link after note on with multiline block",
			input:    "note on invalid\nsome multiline text\nend note\nclass NextClass",
			bs:       expectMultilineClosed,
			wantText: "some multiline text",
		},
		{
			name:     "restore on unexpected token in note definition",
			input:    "note 123 : text\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "",
		},
		{
			name:     "restore on unexpected identifier after note without body",
			input:    "note class\nclass NextClass",
			bs:       expectSingleLine,
			wantText: "",
		},
		{
			name:     "restore on unclosed multiline note block",
			input:    "note left\nsome unclosed text\n",
			bs:       expectMultilineUnclosed,
			wantText: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			require.Equal(t, keyword.Note, keyword.Classify(tok.Literal))

			got, err := p.parseNote(tok)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.True(t, p.ast.HasErrors(), "expected parser to record diagnostics on bad syntax")

			// Verify note text was parsed/retained properly
			require.Equal(t, tc.wantText, got.Text, "expected note text to match after recovery")

			if tc.bs == expectMultilineUnclosed {
				return
			}

			if tc.bs == expectSingleLine {
				require.Equal(t, tokenizer.NEWLINE, p.stream.PeekTokenAt(0).Type, "expected next token to be NEWLINE after recovery")
				p.stream.Emit() // consume newline
			}

			// Canary assertion: next token must parse cleanly as NextClass
			nextTok := p.stream.Emit()
			require.Equal(t, keyword.Class, keyword.Classify(nextTok.Literal))
			nextEnt := p.parseEntity(nextTok)
			require.NotNil(t, nextEnt)
			require.Equal(t, "NextClass", nextEnt.Identifier.EntityName())
		})
	}
}

func TestParseNoteBody(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantText  string
		wantDiags []expectedDiag
	}{
		{
			name:     "single-line colon body",
			input:    ": simple text note\n",
			wantText: "simple text note",
		},
		{
			name:     "multiline body with end note closer",
			input:    "\nfirst line\nsecond line\nend note\n",
			wantText: "first line\nsecond line",
		},
		{
			name:  "unclosed multiline body",
			input: "\nfirst line\nsecond line without closer\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unclosed note: expected end note",
				},
			},
		},
		{
			name:  "unexpected token at body start",
			input: "@unexpected\n",
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected ':' or newline after note definition",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			note := &ast.Note{}
			p.parseNoteBody(note)

			if len(tc.wantDiags) > 0 {
				require.True(t, p.ast.HasErrors())
				assertDiagsEqual(t, tc.wantDiags, p.ast.Diagnostics, note)
			} else {
				require.False(t, p.ast.HasErrors())
				require.Equal(t, tc.wantText, note.Text)
			}
		})
	}
}
