package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

func TestParseEntity(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		kwType    keyword.KeywordKind
		want      *ast.Entity
		expectErr bool
	}{
		{
			name:   "simple class",
			input:  "class MyClass",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
			},
		},
		{
			name:   "abstract class",
			input:  "abstract class MyAbstractClass",
			kwType: keyword.Abstract,
			want: &ast.Entity{
				Identifier: "MyAbstractClass",
				Kind:       ast.AbstractClassKind,
			},
		},
		{
			name:   "class with alias",
			input:  "class MyClass as \"MC\"",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "MyClass",
				Alias:      "MC",
				Kind:       ast.ClassKind,
			},
		},
		{
			name:      "class with invalid alias (not a string)",
			input:     "class MyClass as MC",
			kwType:    keyword.Class,
			want:      nil,
			expectErr: true,
		},
		{
			name:   "class with stereotype",
			input:  "class MyClass <<Service>>",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
				Stereotype: "Service",
			},
		},
		{
			name:   "class with color",
			input:  "class MyClass #FF0000",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
				Color:      "FF0000",
			},
		},
		{
			name:   "class with empty body",
			input:  "class MyClass {\n}",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
			},
		},
		{
			name:   "class def with all modifiers",
			input:  "class MyClass as \"MC\" <T> <<Database>> #FF0000",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "MyClass",
				Alias:      "MC",
				Kind:       ast.ClassKind,
				Generic:    "T",
				Stereotype: "Database",
				Color:      "FF0000",
			},
		},
		{
			name:   "class with field member",
			input:  "class MyClass {\n  +field int\n}",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
				Members: []ast.Member{
					&dialect.GoField{
						Name:       "field",
						Type:       dialect.NamedRef("int"),
						Visibility: ast.Public,
					},
				},
			},
		},
		{
			name:      "invalid alias",
			input:     "class MyClass as",
			kwType:    keyword.Class,
			want:      nil,
			expectErr: true,
		},
		{
			name:      "unclosed stereotype",
			input:     "class MyClass <<Stereo",
			kwType:    keyword.Class,
			want:      nil,
			expectErr: true,
		},
		{
			name:      "unclosed body",
			input:     "class MyClass {",
			kwType:    keyword.Class,
			want:      nil,
			expectErr: true,
		},
		{
			name:      "invalid member inside body",
			input:     "class MyClass {\n  invalidField\n}",
			kwType:    keyword.Class,
			want:      nil,
			expectErr: true,
		},
		{
			name:   "class with generic",
			input:  "class List<T>",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "List",
				Kind:       ast.ClassKind,
				Generic:    "T",
			},
		},
		{
			name:   "class with multiple generics",
			input:  "class Map<K, V>",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: "Map",
				Kind:       ast.ClassKind,
				Generic:    "K, V",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				dialect: dialect.NewGoDialect(),
			}
			tok := p.stream.Emit()
			require.Equal(t, tc.kwType, keyword.Classify(tok.Literal))

			got, err := p.parseEntity(tok)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestParseFieldOrMethod(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		entryType   tokenizer.TokenType
		initialName string
		want        ast.Member
		expectErr   bool
	}{
		// Terminate fields with a newline because
		// parser will only stop collection when it sees a newline
		{
			name:        "simple field",
			input:       "name int\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: &dialect.GoField{
				Name: "name",
				Type: dialect.NamedRef("int"),
			},
		},
		{
			name:        "pointer field",
			input:       "name *MyType\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: &dialect.GoField{
				Name: "name",
				Type: dialect.PointerTo(dialect.NamedRef("MyType")),
			},
		},
		{
			name:        "slice field",
			input:       "name []int\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: &dialect.GoField{
				Name: "name",
				Type: dialect.SliceOf(dialect.NamedRef("int")),
			},
		},
		{
			name:        "simple method",
			input:       "Method()\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want: &dialect.GoMethod{
				Name: "Method",
			},
		},
		{
			name:        "method with parameters and return",
			input:       "Method(a int) error\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want: &dialect.GoMethod{
				Name: "Method",
				Parameters: []dialect.GoParameter{
					{Name: "a", Type: dialect.NamedRef("int")},
				},
				ReturnType: []dialect.GoParameter{
					{Type: dialect.NamedRef("error")},
				},
			},
		},
		{
			name:        "method defined as method modifier but no params",
			input:       "{method} Method()\n",
			entryType:   tokenizer.LBRACE,
			initialName: "",
			want: &dialect.GoMethod{
				Name:      "Method",
				Modifiers: []string{"method"},
			},
		},
		{
			name:        "both field and method modifiers",
			input:       "{field} {method} Name()\n",
			entryType:   tokenizer.LBRACE,
			initialName: "",
			want:        nil,
			expectErr:   true,
		},
		{
			name:        "invalid field (no type)",
			input:       "name\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want:        nil,
			expectErr:   true,
		},
		{
			name:        "invalid method (unclosed param parenthesis)",
			input:       "Method(a int\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want:        nil,
			expectErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				dialect: dialect.NewGoDialect(),
			}

			var entryTok tokenizer.Token
			var mod *string
			vis := ast.UnknownVisibility

			if tc.entryType == tokenizer.LBRACE {
				m, err := p.stream.TryReadModifier()
				require.NoError(t, err)
				mod = &m
				entryTok = tokenizer.Token{Type: tokenizer.LBRACE}
			} else {
				entryTok = p.stream.Emit()
				require.Equal(t, tc.entryType, entryTok.Type)
				if entryTok.Type == tokenizer.IDENTIFIER {
					require.Equal(t, tc.initialName, entryTok.Literal)
				}
			}

			got, err := p.parseFieldOrMethod(mod, vis, entryTok)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestParseEntityMember(t *testing.T) {
	tt := []struct {
		name      string
		input     string
		want      ast.Member
		expectErr bool
	}{
		{
			name:  "class separator dots",
			input: ".. separator ..",
			want: ast.ClassSeparator{
				Label: "separator",
				Type:  '.',
			},
		},
		{
			name:  "class separator hyphens",
			input: "-- section --",
			want: ast.ClassSeparator{
				Label: "section",
				Type:  '-',
			},
		},
		{
			name:  "public field member",
			input: "+field int\n",
			want: &dialect.GoField{
				Name:       "field",
				Type:       dialect.NamedRef("int"),
				Visibility: ast.Public,
			},
		},
		{
			name:  "private method member",
			input: "-Method()\n",
			want: &dialect.GoMethod{
				Name:       "Method",
				Visibility: ast.Private,
			},
		},
		{
			name:  "field with modifier",
			input: "{field} myField int\n",
			want: &dialect.GoField{
				Name:      "myField",
				Type:      dialect.NamedRef("int"),
				Modifiers: []string{"field"},
			},
		},
		{
			name:      "unclosed modifier",
			input:     "{field myField int",
			expectErr: true,
		},
		{
			name:      "unexpected token in body",
			input:     "@invalid",
			expectErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				dialect: dialect.NewGoDialect(),
			}
			got, err := p.parseEntityMember()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestParseArrowTokens(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        *ast.Relationship
		expectErr   bool
		errContains string
	}{
		{
			name:  "simple solid right arrow",
			input: "-->",
			want: &ast.Relationship{
				Body:   '-',
				RArrow: '>',
			},
		},
		{
			name:  "simple solid right arrow short",
			input: "->",
			want: &ast.Relationship{
				Body:   '-',
				RArrow: '>',
			},
		},
		{
			name:  "simple dotted right arrow",
			input: "..>",
			want: &ast.Relationship{
				Body:   '.',
				RArrow: '>',
			},
		},
		{
			name:  "double-headed arrow solid",
			input: "<-->",
			want: &ast.Relationship{
				LArrow: '<',
				Body:   '-',
				RArrow: '>',
			},
		},
		{
			name:  "double-headed arrow dotted",
			input: "<..>",
			want: &ast.Relationship{
				LArrow: '<',
				Body:   '.',
				RArrow: '>',
			},
		},
		{
			name:  "left extension solid right arrow",
			input: "<|-->",
			want: &ast.Relationship{
				LArrow: '|',
				Body:   '-',
				RArrow: '>',
			},
		},
		{
			name:  "solid right extension arrow",
			input: "--|>",
			want: &ast.Relationship{
				Body:   '-',
				RArrow: '|',
			},
		},
		{
			name:  "double extension solid",
			input: "<|--|>",
			want: &ast.Relationship{
				LArrow: '|',
				Body:   '-',
				RArrow: '|',
			},
		},
		{
			name:  "curly braces left/right",
			input: "}--{",
			want: &ast.Relationship{
				LArrow: '}',
				Body:   '-',
				RArrow: '{',
			},
		},
		{
			name:  "curly braces dotted left/right",
			input: "}..{",
			want: &ast.Relationship{
				LArrow: '}',
				Body:   '.',
				RArrow: '{',
			},
		},
		{
			name:  "lolipop right arrowhead",
			input: "--()",
			want: &ast.Relationship{
				Body:   '-',
				RArrow: '(',
			},
		},
		{
			name:        "less-than pipe greater-than diamond/extension",
			input:       "<|>",
			expectErr:   true,
			errContains: "Unexpected token as the relationship body",
			want: &ast.Relationship{
				LArrow: '|',
			},
		},
		{
			name:  "custom 'x' left/right",
			input: "x--x",
			want: &ast.Relationship{
				LArrow: 'x',
				Body:   '-',
				RArrow: 'x',
			},
		},
		{
			name:  "custom 'o' left/right",
			input: "o--o",
			want: &ast.Relationship{
				LArrow: 'o',
				Body:   '-',
				RArrow: 'o',
			},
		},
		{
			name:  "asterisk left/right",
			input: "*--*",
			want: &ast.Relationship{
				LArrow: '*',
				Body:   '-',
				RArrow: '*',
			},
		},
		{
			name:  "plus left/right",
			input: "+--+",
			want: &ast.Relationship{
				LArrow: '+',
				Body:   '-',
				RArrow: '+',
			},
		},
		{
			name:  "caret left/right",
			input: "^--^",
			want: &ast.Relationship{
				LArrow: '^',
				Body:   '-',
				RArrow: '^',
			},
		},
		{
			name:  "hash left/right",
			input: "#--#",
			want: &ast.Relationship{
				LArrow: '#',
				Body:   '-',
				RArrow: '#',
			},
		},
		{
			name:  "left direction",
			input: "-left->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.Left,
				RArrow:    '>',
			},
		},
		{
			name:  "right direction",
			input: "-right->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.Right,
				RArrow:    '>',
			},
		},
		{
			name:  "up direction",
			input: "-up->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.Top,
				RArrow:    '>',
			},
		},
		{
			name:  "down direction",
			input: "-down->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.Bottom,
				RArrow:    '>',
			},
		},
		{
			name:  "single attribute",
			input: "-[foo]->",
			want: &ast.Relationship{
				Body:   '-',
				Attrs:  []string{"foo"},
				RArrow: '>',
			},
		},
		{
			name:  "multiple attributes",
			input: "-[foo,bar]->",
			want: &ast.Relationship{
				Body:   '-',
				Attrs:  []string{"foo", "bar"},
				RArrow: '>',
			},
		},
		{
			name:  "attributes and direction",
			input: "-[foo]left->",
			want: &ast.Relationship{
				Body:      '-',
				Attrs:     []string{"foo"},
				Direction: ast.Left,
				RArrow:    '>',
			},
		},
		{
			name:  "direction and attributes",
			input: "-left[foo]->",
			want: &ast.Relationship{
				Body:      '-',
				Attrs:     []string{"foo"},
				Direction: ast.Left,
				RArrow:    '>',
			},
		},
		{
			name:  "multitoken attributes",
			input: "-[#foo,%bar]->",
			want: &ast.Relationship{
				Body:   '-',
				Attrs:  []string{"#foo", "%bar"},
				RArrow: '>',
			},
		},
		{
			name:  "no arrowheads solid",
			input: "--",
			want: &ast.Relationship{
				Body: '-',
			},
		},
		{
			name:  "no arrowheads dotted",
			input: "..",
			want: &ast.Relationship{
				Body: '.',
			},
		},
		{
			name:        "empty input error",
			input:       "",
			expectErr:   true,
			errContains: "Unexpected token at the start of relationship definition",
		},
		{
			name:        "invalid starting token error",
			input:       "@",
			expectErr:   true,
			errContains: "Unexpected token at the start of relationship definition",
		},
		{
			name:        "missing trailing body rune error",
			input:       "-up>",
			expectErr:   true,
			errContains: "Unexpected token in body relationship definition",
		},
		{
			name:        "invalid top direction value error",
			input:       "-top->",
			expectErr:   true,
			errContains: "Unexpected direction in relationship",
		},
		{
			name:        "invalid bottom direction value error",
			input:       "-bottom->",
			expectErr:   true,
			errContains: "Unexpected direction in relationship",
		},
		{
			name:        "invalid identifier error",
			input:       "-foo->",
			expectErr:   true,
			errContains: "Unexpected identifier in relationship definition",
		},
		{
			name:        "unexpected separator inside direction error",
			input:       "-left-[foo]->",
			expectErr:   true,
			errContains: "Cannot separate direction and attributes with a body token",
		},
		{
			name:        "unclosed pipe extension error",
			input:       "--|",
			expectErr:   true,
			errContains: "Expected '|>' after relationship",
		},
		{
			name:        "lolipop right after direction error",
			input:       "-left-()",
			expectErr:   true,
			errContains: "Lolipop interface cannot contain direction or attributes",
		},
		{
			name:        "lolipop right after attributes error",
			input:       "-[foo]-()",
			expectErr:   true,
			errContains: "Lolipop interface cannot contain direction or attributes",
		},
		{
			name:        "mixed body types error",
			input:       "-.-",
			expectErr:   true,
			errContains: "Different body type runes in relationship",
		},
		{
			name:        "unclosed attributes at EOF error",
			input:       "-[foo",
			expectErr:   true,
			errContains: "Unexpected break in relationship attribute container",
		},
		{
			name:        "unclosed attributes at newline error",
			input:       "-[foo\n",
			expectErr:   true,
			errContains: "Unexpected break in relationship attribute container",
		},
		{
			name:        "missing attribute (double comma)",
			input:       "-[foo,,bar]->",
			expectErr:   true,
			errContains: "Unexpected comma in relationship attribute container",
			want: &ast.Relationship{
				Body:  '-',
				Attrs: []string{"foo"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream: tokenizer.NewTokenStream(tc.input),
			}
			var got ast.Relationship
			err := p.parseArrowTokens(&got)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tc.want != nil {
				require.Equal(t, *tc.want, got)
			}
		})
	}
}

func TestParseNote(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        *ast.Note
		expectErr   bool
		errContains string
	}{
		// --- Relative Notes (parseReltiveNote) ---
		{
			name:  "relative note left with colon",
			input: "note left : some note text",
			want: &ast.Note{
				Text:      " some note text",
				Direction: ast.Left,
			},
		},
		{
			name:  "relative note right of target with colon",
			input: "note right of MyClass : some note text",
			want: &ast.Note{
				Text:      " some note text",
				Direction: ast.Right,
				Target:    "MyClass",
			},
		},
		{
			name:  "relative note top of target with color and colon",
			input: "note top of MyClass #green : some note text",
			want: &ast.Note{
				Text:      " some note text",
				Direction: ast.Top,
				Target:    "MyClass",
			},
		},
		{
			name:  "relative note bottom on link with colon",
			input: "note bottom on link : some note text",
			want: &ast.Note{
				Text:      " some note text",
				Direction: ast.Bottom,
				Target:    "link",
			},
		},
		{
			name:  "relative note left with multiline",
			input: "note left\nsome note text\nend note",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.Left,
			},
		},
		{
			name:  "relative note left of target multiline",
			input: "note left of MyClass\nsome note text\nend note",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.Left,
				Target:    "MyClass",
			},
		},
		{
			name:        "relative note invalid identifier after direction",
			input:       "note left invalid : text",
			expectErr:   true,
			errContains: "Expected ':' or newline after note definition",
		},
		{
			name:        "relative note expected identifier for target",
			input:       "note left of : text",
			expectErr:   true,
			errContains: "Expected identifier for a note target",
		},
		{
			name:        "relative note unexpected identifier for link target",
			input:       "note left on MyClass : text",
			expectErr:   true,
			errContains: "Unexpected identifier for a note link target",
		},

		// --- Inline Alias Notes (parseInlineAliasNote) ---
		{
			name:  "inline alias note",
			input: `note "some note text" as N1`,
			want: &ast.Note{
				Text:  "some note text",
				Alias: "N1",
			},
		},
		{
			name:  "inline alias note with color",
			input: `note "some note text" as N1 #blue`,
			want: &ast.Note{
				Text:  "some note text",
				Alias: "N1",
			},
		},
		{
			name:        "inline alias note missing as",
			input:       `note "some note text" N1`,
			expectErr:   true,
			errContains: "Expected alias keyword after note text",
		},
		{
			name:        "inline alias note missing identifier",
			input:       `note "some note text" as`,
			expectErr:   true,
			errContains: "Expected identifier after alias keyword",
		},

		// --- Multiline Alias Notes (parseMultilineAliasNote) ---
		{
			name:        "multiline alias note with colon (invalid)",
			input:       "note as N1 : some note text",
			expectErr:   true,
			errContains: "Expected newline after alias keyword",
		},
		{
			name:        "multiline alias note with colon and color (invalid)",
			input:       "note as N1 #red : some note text",
			expectErr:   true,
			errContains: "Expected newline after alias keyword",
		},
		{
			name:        "multiline alias note missing identifier",
			input:       "note as",
			expectErr:   true,
			errContains: "Expected identifier after alias keyword",
		},
		{
			name:  "multiline alias note with newline",
			input: "note as N1\nsome note text\nend note",
			want: &ast.Note{
				Text: "some note text",
			},
		},

		// --- Link Notes (parseLinkNote) ---
		{
			name:  "link note on link with colon",
			input: "note on link : some note text",
			want: &ast.Note{
				Text: " some note text",
			},
		},
		{
			name:  "link note on link multiline",
			input: "note on link\nsome note text\nend note",
			want: &ast.Note{
				Text: "some note text",
			},
		},
		{
			name:        "unexpected identifier after note",
			input:       "note invalid link : text",
			expectErr:   true,
			errContains: "expected direction, string, note position or alias after 'note'",
		},
		{
			name:        "link note expected link after note on",
			input:       "note on invalid : text",
			expectErr:   true,
			errContains: "Expected 'link' after 'note on'",
		},

		// --- Generic Parser / parseNoteBody Errors ---
		{
			name:        "parseNote unexpected starting token",
			input:       "note class : text",
			expectErr:   true,
			errContains: "expected direction, string, note position or alias after 'note'",
		},
		{
			name:        "parseNoteBody unexpected body definition token",
			input:       "note left class : text",
			expectErr:   true,
			errContains: "Expected ':' or newline after note definition",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream: tokenizer.NewTokenStream(tc.input),
				ast:    &ast.Diagram{},
			}
			tok := p.stream.Emit()
			require.Equal(t, keyword.Note, keyword.Classify(tok.Literal), "First token must be 'note' for test input: %q", tc.input)

			note, err := p.parseNote(tok)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, *tc.want, note)
			}
		})
	}
}

func TestParseContainer(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		kwType      keyword.KeywordKind
		want        *ast.Container
		expectErr   bool
		errContains string
	}{
		{
			name:        "together empty error",
			input:       "together",
			kwType:      keyword.Together,
			expectErr:   true,
			errContains: "Expected container body to end",
		},
		{
			name:   "together empty newline",
			input:  "together\n",
			kwType: keyword.Together,
			want:   &ast.Container{},
		},
		{
			name:   "together empty body",
			input:  "together {}",
			kwType: keyword.Together,
			want:   &ast.Container{},
		},
		{
			name:   "together with single class",
			input:  "together { class A }",
			kwType: keyword.Together,
			want: &ast.Container{
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: "A",
						Kind:       ast.ClassKind,
					},
				},
			},
		},
		{
			name:   "package simple identifier",
			input:  "package mypkg\n",
			kwType: keyword.Package,
			want: &ast.Container{
				Identifier: "mypkg",
			},
		},
		{
			name:   "package string identifier and alias",
			input:  `package "My Package" as mypkg` + "\n",
			kwType: keyword.Package,
			want: &ast.Container{
				Identifier: "mypkg",
				Alias:      "My Package",
			},
		},
		{
			name:   "package identifier and string alias",
			input:  `package mypkg as "My Package"` + "\n",
			kwType: keyword.Package,
			want: &ast.Container{
				Identifier: "mypkg",
				Alias:      "My Package",
			},
		},
		{
			name:   "package identical identifier and alias",
			input:  "package mypkg as otherpkg\n",
			kwType: keyword.Package,
			want: &ast.Container{
				Identifier: "otherpkg",
				Alias:      "mypkg",
			},
		},
		{
			name:   "package with stereotype",
			input:  "package mypkg <<Service>>\n",
			kwType: keyword.Package,
			want: &ast.Container{
				Identifier: "mypkg",
				Stereotype: "Service",
			},
		},
		{
			name:   "package with stereotype and color",
			input:  "package mypkg <<Service>> #green\n",
			kwType: keyword.Package,
			want: &ast.Container{
				Identifier: "mypkg",
				Stereotype: "Service",
				Color:      "green",
			},
		},
		{
			name:   "package body with relationship and nested package",
			input:  "package mypkg { class A together { class B } A -> B\n}",
			kwType: keyword.Package,
			want: &ast.Container{
				Identifier: "mypkg",
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: "A",
						Kind:       ast.ClassKind,
					},
					ast.Container{
						Statements: []ast.Statement{
							&ast.Entity{
								Identifier: "B",
								Kind:       ast.ClassKind,
							},
						},
					},
					ast.Relationship{
						LHS:    "A",
						RHS:    "B",
						Body:   '-',
						RArrow: '>',
					},
				},
			},
		},
		{
			name:        "package incomplete alias",
			input:       "package mypkg as\n",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "Expected container name or alias",
		},
		{
			name:        "package unclosed stereotype",
			input:       "package mypkg <<Service\n",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "unexpected EOF",
		},
		{
			name:        "package unclosed body",
			input:       "package mypkg { class A",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "unexpected EOF",
		},
		{
			name:        "package unexpected token in body",
			input:       "package mypkg { @invalid }",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "Expected a statement in a container body",
		},
		{
			name:        "package body newline error",
			input:       "package mypkg {\n}",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "Expected a statement in a container body",
		},
		{
			name:        "package with same line body and color without colon/newline",
			input:       "package mypkg #red { class A }",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "Expected container body to end",
		},
		{
			name:        "package body class error",
			input:       "package mypkg { class MyClass as }",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "Expected token for entity identifier or alias",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			require.Equal(t, tc.kwType, keyword.Classify(tok.Literal))

			got, err := p.parseContainer(tok)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, *tc.want, got)
			}
		})
	}
}
