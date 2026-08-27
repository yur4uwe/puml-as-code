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
		want      ast.Statement
		expectErr bool
	}{
		{
			name:   "simple class",
			input:  "class MyClass",
			kwType: keyword.Class,
			want: ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.EntityClass,
			},
		},
		{
			name:   "abstract class",
			input:  "abstract class MyAbstractClass",
			kwType: keyword.Abstract,
			want: ast.Entity{
				Identifier: "MyAbstractClass",
				Kind:       ast.EntityAbstractClass,
			},
		},
		{
			name:   "class with alias",
			input:  "class MyClass as \"MC\"",
			kwType: keyword.Class,
			want: ast.Entity{
				Identifier: "MyClass",
				Alias:      "MC",
				Kind:       ast.EntityClass,
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
			want: ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.EntityClass,
				Stereotype: "Service",
			},
		},
		{
			name:   "class with color",
			input:  "class MyClass #FF0000",
			kwType: keyword.Class,
			want: ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.EntityClass,
				Color:      "FF0000",
			},
		},
		{
			name:   "class with empty body",
			input:  "class MyClass {\n}",
			kwType: keyword.Class,
			want: ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.EntityClass,
			},
		},
		{
			name:   "class def with all modifiers",
			input:  "class MyClass as \"MC\" <T> <<Database>> #FF0000",
			kwType: keyword.Class,
			want: ast.Entity{
				Identifier: "MyClass",
				Alias:      "MC",
				Kind:       ast.EntityClass,
				Generic:    "T",
				Stereotype: "Database",
				Color:      "FF0000",
			},
		},
		{
			name:   "class with field member",
			input:  "class MyClass {\n  +field int\n}",
			kwType: keyword.Class,
			want: ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.EntityClass,
				Members: []ast.Member{
					&dialect.GoField{
						Name:       "field",
						Type:       dialect.NamedRef("int"),
						Visibility: ast.VisibilityPublic,
					},
				},
			},
		},
		{
			name:   "class with package separator",
			input:  "class net.http.Client\n",
			kwType: keyword.Class,
			want: ast.Container{
				Identifier: "net",
				Statements: []ast.Statement{
					ast.Container{
						Identifier: "http",
						Statements: []ast.Statement{
							ast.Entity{
								Identifier: "Client",
								Kind:       ast.EntityClass,
							},
						},
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
			want: ast.Entity{
				Identifier: "List",
				Kind:       ast.EntityClass,
				Generic:    "T",
			},
		},
		{
			name:   "class with multiple generics",
			input:  "class Map<K, V>",
			kwType: keyword.Class,
			want: ast.Entity{
				Identifier: "Map",
				Kind:       ast.EntityClass,
				Generic:    "K, V",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
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
				Dialect: dialect.NewGoDialect(),
			}

			var entryTok tokenizer.Token
			var mod *string
			vis := ast.VisibilityUnknown

			if tc.entryType == tokenizer.LBRACE {
				m, err := p.tryReadModifier()
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

			got, err := p.parseFieldOrMethod(mod, vis, entryTok, nil)
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
				Visibility: ast.VisibilityPublic,
			},
		},
		{
			name:  "private method member",
			input: "-Method()\n",
			want: &dialect.GoMethod{
				Name:       "Method",
				Visibility: ast.VisibilityPrivate,
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
				Dialect: dialect.NewGoDialect(),
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
				Body:    '-',
				RArrow:  '>',
				TypeRHS: ast.RelationAssociation,
			},
		},
		{
			name:  "simple solid right arrow short",
			input: "->",
			want: &ast.Relationship{
				Body:    '-',
				RArrow:  '>',
				TypeRHS: ast.RelationAssociation,
			},
		},
		{
			name:  "simple dotted right arrow",
			input: "..>",
			want: &ast.Relationship{
				Body:    '.',
				RArrow:  '>',
				TypeRHS: ast.RelationDependency,
			},
		},
		{
			name:  "double-headed arrow solid",
			input: "<-->",
			want: &ast.Relationship{
				LArrow:  '<',
				Body:    '-',
				RArrow:  '>',
				TypeLHS: ast.RelationAssociation,
				TypeRHS: ast.RelationAssociation,
			},
		},
		{
			name:  "double-headed arrow dotted",
			input: "<..>",
			want: &ast.Relationship{
				LArrow:  '<',
				Body:    '.',
				RArrow:  '>',
				TypeLHS: ast.RelationDependency,
				TypeRHS: ast.RelationDependency,
			},
		},
		{
			name:  "left extension solid right arrow",
			input: "<|-->",
			want: &ast.Relationship{
				LArrow:  '|',
				Body:    '-',
				RArrow:  '>',
				TypeLHS: ast.RelationInheritance,
				TypeRHS: ast.RelationAssociation,
			},
		},
		{
			name:  "solid right extension arrow",
			input: "--|>",
			want: &ast.Relationship{
				Body:    '-',
				RArrow:  '|',
				TypeRHS: ast.RelationInheritance,
			},
		},
		{
			name:  "double extension solid",
			input: "<|--|>",
			want: &ast.Relationship{
				LArrow:  '|',
				Body:    '-',
				RArrow:  '|',
				TypeLHS: ast.RelationInheritance,
				TypeRHS: ast.RelationInheritance,
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
				LArrow:  'o',
				Body:    '-',
				RArrow:  'o',
				TypeLHS: ast.RelationAggregation,
				TypeRHS: ast.RelationAggregation,
			},
		},
		{
			name:  "asterisk left/right",
			input: "*--*",
			want: &ast.Relationship{
				LArrow:  '*',
				Body:    '-',
				RArrow:  '*',
				TypeLHS: ast.RelationComposition,
				TypeRHS: ast.RelationComposition,
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
				Direction: ast.DirectionLeft,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "right direction",
			input: "-right->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.DirectionRight,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "up direction",
			input: "-up->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.DirectionTop,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "down direction",
			input: "-down->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.DirectionBottom,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "single attribute",
			input: "-[foo]->",
			want: &ast.Relationship{
				Body:    '-',
				Attrs:   []string{"foo"},
				RArrow:  '>',
				TypeRHS: ast.RelationAssociation,
			},
		},
		{
			name:  "multiple attributes",
			input: "-[foo,bar]->",
			want: &ast.Relationship{
				Body:    '-',
				Attrs:   []string{"foo", "bar"},
				RArrow:  '>',
				TypeRHS: ast.RelationAssociation,
			},
		},
		{
			name:  "attributes and direction",
			input: "-[foo]left->",
			want: &ast.Relationship{
				Body:      '-',
				Attrs:     []string{"foo"},
				Direction: ast.DirectionLeft,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "direction and attributes",
			input: "-left[foo]->",
			want: &ast.Relationship{
				Body:      '-',
				Attrs:     []string{"foo"},
				Direction: ast.DirectionLeft,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "multitoken attributes",
			input: "-[#foo,%bar]->",
			want: &ast.Relationship{
				Body:    '-',
				Attrs:   []string{"#foo", "%bar"},
				RArrow:  '>',
				TypeRHS: ast.RelationAssociation,
			},
		},
		{
			name:  "no arrowheads solid",
			input: "--",
			want: &ast.Relationship{
				Body:    '-',
				TypeLHS: ast.RelationAssociation,
				TypeRHS: ast.RelationAssociation,
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
			input: "note left : some note text\n",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.DirectionLeft,
			},
		},
		{
			name:  "relative note right of target with colon",
			input: "note right of MyClass : some note text\n",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.DirectionRight,
				Target:    &ast.TargetRef{Entity: "MyClass"},
			},
		},
		{
			name:  "relative note top of target with color and colon",
			input: "note top of MyClass #green : some note text\n",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.DirectionTop,
				Target:    &ast.TargetRef{Entity: "MyClass"},
			},
		},
		{
			name:  "relative note bottom on link with colon",
			input: "note bottom on link : some note text\n",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.DirectionBottom,
				Target:    &ast.TargetRef{Entity: "link"},
			},
		},
		{
			name:  "relative note left with multiline",
			input: "note left\nsome note text\nend note\n",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.DirectionLeft,
			},
		},
		{
			name:  "relative note left of target multiline",
			input: "note left of MyClass\nsome note text\nend note",
			want: &ast.Note{
				Text:      "some note text",
				Direction: ast.DirectionLeft,
				Target:    &ast.TargetRef{Entity: "MyClass"},
			},
		},
		{
			name:        "relative note invalid identifier after direction",
			input:       "note left invalid : text\n",
			expectErr:   true,
			errContains: "Unexpected identifier after direction",
		},
		{
			name:        "relative note expected identifier for target",
			input:       "note left of : text\n",
			expectErr:   true,
			errContains: "Expected ':' or newline after note definition",
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
				Text:       "some note text",
				Identifier: "N1",
			},
		},
		{
			name:  "inline alias note with color",
			input: `note "some note text" as N1 #blue`,
			want: &ast.Note{
				Text:       "some note text",
				Identifier: "N1",
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
			input: "note on link : some note text\n",
			want: &ast.Note{
				Text: "some note text",
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
			errContains: "Unexpected identifier after direction",
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

			note, err := p.parseNote()
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
		want        ast.Container
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
			want: ast.Container{
				Kind: ast.ContainerTogether,
			},
		},
		{
			name:   "together empty body",
			input:  "together {}",
			kwType: keyword.Together,
			want: ast.Container{
				Kind: ast.ContainerTogether,
			},
		},
		{
			name:   "together with single class",
			input:  "together { class A }",
			kwType: keyword.Together,
			want: ast.Container{
				Kind: ast.ContainerTogether,
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "package simple identifier",
			input:  "package mypkg\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "mypkg",
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package string identifier and alias",
			input:  `package "My Package" as mypkg` + "\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "mypkg",
				Alias:      "My Package",
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package identifier and string alias",
			input:  `package mypkg as "My Package"` + "\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "mypkg",
				Alias:      "My Package",
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package identical identifier and alias",
			input:  "package mypkg as otherpkg\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "otherpkg",
				Alias:      "mypkg",
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package with stereotype",
			input:  "package mypkg <<Service>>\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "mypkg",
				Stereotype: "Service",
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package with stereotype and color",
			input:  "package mypkg <<Service>> #green\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "mypkg",
				Stereotype: "Service",
				Color:      "green",
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package body with relationship and nested package",
			input:  "package mypkg { class A together { class B } A -> B\n}",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "mypkg",
				Kind:       ast.ContainerPackage,
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
					ast.Container{
						Kind: ast.ContainerTogether,
						Statements: []ast.Statement{
							ast.Entity{
								Identifier: "B",
								Kind:       ast.EntityClass,
							},
						},
					},
					ast.Relationship{
						LHS:     ast.TargetRef{Entity: "A"},
						RHS:     ast.TargetRef{Entity: "B"},
						Body:    '-',
						RArrow:  '>',
						TypeRHS: ast.RelationAssociation,
					},
				},
			},
		},
		{
			name:   "package with the keyword as a name in relationship",
			input:  "package p { class folder {} folder --> p }",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "p",
				Kind:       ast.ContainerPackage,
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "folder",
						Kind:       ast.EntityClass,
					},
					ast.Relationship{
						LHS:     ast.TargetRef{Entity: "folder"},
						RHS:     ast.TargetRef{Entity: "p"},
						Body:    '-',
						RArrow:  '>',
						TypeRHS: ast.RelationAssociation,
					},
				},
			},
		},
		{
			name:   "expect to correctly parse nested containers",
			input:  "package p.p {}",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: "p",
				Kind:       ast.ContainerPackage,
				Statements: []ast.Statement{
					ast.Container{
						Identifier: "p",
						Kind:       ast.ContainerPackage,
					},
				},
			},
		},
		{
			name:        "package incomplete alias",
			input:       "package mypkg as\n",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "Expected identifier for container name",
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
			name:   "package with same line body and color",
			input:  "package mypkg #red { class A }",
			kwType: keyword.Package,
			want: ast.Container{
				Kind:       ast.ContainerPackage,
				Identifier: "mypkg",
				Color:      "red",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:        "package body class error",
			input:       "package mypkg { class MyClass as }",
			kwType:      keyword.Package,
			expectErr:   true,
			errContains: "Expected token for entity identifier or alias",
		},
		// --- Container keyword coverage ---
		{
			name:   "folder with body",
			input:  "folder myfolder { class A }",
			kwType: keyword.Folder,
			want: ast.Container{
				Kind:       ast.ContainerFolder,
				Identifier: "myfolder",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "frame with body",
			input:  "frame myframe { class A }",
			kwType: keyword.Frame,
			want: ast.Container{
				Kind:       ast.ContainerFrame,
				Identifier: "myframe",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "rectangle with body",
			input:  "rectangle myrect { class A }",
			kwType: keyword.Rectangle,
			want: ast.Container{
				Kind:       ast.ContainerRectangle,
				Identifier: "myrect",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "cloud with body",
			input:  "cloud mycloud { class A }",
			kwType: keyword.Cloud,
			want: ast.Container{
				Kind:       ast.ContainerCloud,
				Identifier: "mycloud",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "database with body",
			input:  "database mydb { class A }",
			kwType: keyword.Database,
			want: ast.Container{
				Kind:       ast.ContainerDatabase,
				Identifier: "mydb",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "node with body",
			input:  "node mynode { class A }",
			kwType: keyword.Node,
			want: ast.Container{
				Kind:       ast.ContainerNode,
				Identifier: "mynode",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "namespace with body",
			input:  "namespace myns { class A }",
			kwType: keyword.Namespace,
			want: ast.Container{
				Kind:       ast.ContainerNamespace,
				Identifier: "myns",
				Statements: []ast.Statement{
					ast.Entity{
						Identifier: "A",
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "folder empty body",
			input:  "folder myfolder {}",
			kwType: keyword.Folder,
			want: ast.Container{
				Kind:       ast.ContainerFolder,
				Identifier: "myfolder",
			},
		},
		{
			name:   "namespace with alias and stereotype",
			input:  `namespace myns as "My Namespace" <<API>>` + "\n",
			kwType: keyword.Namespace,
			want: ast.Container{
				Kind:       ast.ContainerNamespace,
				Identifier: "myns",
				Alias:      "My Namespace",
				Stereotype: "API",
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
			require.Equal(t, tc.kwType, keyword.Classify(tok.Literal))

			got, err := p.parseContainer(tok)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestParseTargetRef(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ast.TargetRef
	}{
		{
			name:  "simple entity",
			input: "Client",
			want:  ast.TargetRef{Entity: "Client"},
		},
		{
			name:  "package path and entity",
			input: "net.http.Client",
			want:  ast.TargetRef{PackagePath: []string{"net", "http"}, Entity: "Client"},
		},
		{
			name:  "entity and simple member",
			input: "Client::Do",
			want:  ast.TargetRef{Entity: "Client", Member: "Do"},
		},
		{
			name:  "entity and method parens",
			input: "Client::Do()",
			want:  ast.TargetRef{Entity: "Client", Member: "Do()"},
		},
		{
			name:  "entity and method with parameters",
			input: `Client::"Do(req Request)"`,
			want:  ast.TargetRef{Entity: "Client", Member: "Do(req Request)"},
		},
		{
			name:  "package path entity and method in quotes",
			input: `net.http.Client::"Do(Context)"`,
			want:  ast.TargetRef{PackagePath: []string{"net", "http"}, Entity: "Client", Member: "Do(Context)"},
		},
		{
			name:  "package separator follows the entity but not a part of it",
			input: "Foo ..|> Bar",
			want:  ast.TargetRef{Entity: "Foo"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			firstTok := p.stream.Emit()
			got, err := p.parseTargetRef(firstTok)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParseInlineMember(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantEntity     string
		wantMethod     bool
		wantField      bool
		wantModifiers  []string
		wantVisibility ast.VisibilityKind
		wantMember     func(t *testing.T, member ast.Member)
	}{
		{
			name:           "field with visibility",
			input:          "Foo : +id string\n",
			wantEntity:     "Foo",
			wantField:      true,
			wantVisibility: ast.VisibilityPublic,
		},
		{
			name:           "method with visibility and params",
			input:          "Foo : +DoWork(ctx Context) error\n",
			wantEntity:     "Foo",
			wantMethod:     true,
			wantVisibility: ast.VisibilityPublic,
		},
		{
			name:       "field with static modifier",
			input:      "Foo : {static} count int\n",
			wantEntity: "Foo",
			wantField:  true,
			wantModifiers: []string{
				"static",
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
			for tok.Type == tokenizer.NEWLINE {
				tok = p.stream.Emit()
			}
			got, err := p.parseInlineMember(tok)
			require.NoError(t, err)
			ent, ok := got.(ast.Entity)
			require.True(t, ok)
			require.Equal(t, tc.wantEntity, ent.Identifier)
			require.Len(t, ent.Members, 1)
			if tc.wantMethod {
				method, ok := ent.Members[0].(*dialect.GoMethod)
				require.True(t, ok)
				require.Equal(t, tc.wantModifiers, method.Modifiers)
				require.Equal(t, tc.wantVisibility, method.Visibility)
			} else {
				field, ok := ent.Members[0].(*dialect.GoField)
				require.True(t, ok)
				require.Equal(t, tc.wantModifiers, field.Modifiers)
				require.Equal(t, tc.wantVisibility, field.Visibility)
			}
		})
	}
}

func TestParseEntityAndContainerTags(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStereo string
		wantTags   []string
	}{
		{
			name:       "pre-stereotype tags",
			input:      "class Foo $tag1 $tag2 <<stereo>>\n",
			wantStereo: "stereo",
			wantTags:   []string{"tag1", "tag2"},
		},
		{
			name:       "post-stereotype tags",
			input:      "class Foo <<stereo>> $tag3\n",
			wantStereo: "stereo",
			wantTags:   []string{"tag3"},
		},
		{
			name:       "precedence: pre-stereotype tags override post-stereotype tags",
			input:      "class Foo $pre <<stereo>> $post\n",
			wantStereo: "stereo",
			wantTags:   []string{"pre"},
		},
		{
			name:       "standalone tag without stereotype",
			input:      "class Foo $alone\n",
			wantStereo: "",
			wantTags:   []string{"alone"},
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
			got, err := p.parseEntity(tok)
			require.NoError(t, err)
			ent, ok := got.(ast.Entity)
			require.True(t, ok)
			require.Equal(t, tc.wantStereo, ent.Stereotype)
			require.Equal(t, tc.wantTags, ent.Tags)
		})
	}
}
