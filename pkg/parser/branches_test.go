package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

func TestParseEntity(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		tokType   tokenizer.TokenType
		want      *ast.Entity
		expectErr bool
	}{
		{
			name:    "simple class",
			input:   "class MyClass",
			tokType: tokenizer.CLASS,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
			},
		},
		{
			name:    "abstract class",
			input:   "abstract class MyAbstractClass",
			tokType: tokenizer.ABSTRACT,
			want: &ast.Entity{
				Identifier: "MyAbstractClass",
				Kind:       ast.AbstractClassKind,
			},
		},
		{
			name:    "class with alias",
			input:   "class MyClass as \"MC\"",
			tokType: tokenizer.CLASS,
			want: &ast.Entity{
				Identifier: "MyClass",
				Alias:      "MC",
				Kind:       ast.ClassKind,
			},
		},
		{
			name:      "class with invalid alias (not a string)",
			input:     "class MyClass as MC",
			tokType:   tokenizer.CLASS,
			want:      nil,
			expectErr: true,
		},
		{
			name:    "class with stereotype",
			input:   "class MyClass <<Service>>",
			tokType: tokenizer.CLASS,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
				Stereotype: "Service",
			},
		},
		{
			name:    "class with color",
			input:   "class MyClass #FF0000",
			tokType: tokenizer.CLASS,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
				Color:      "FF0000",
			},
		},
		{
			name:    "class with empty body",
			input:   "class MyClass {\n}",
			tokType: tokenizer.CLASS,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
			},
		},
		{
			name:    "class def with all modifiers",
			input:   "class MyClass as \"MC\" <T> <<Database>> #FF0000",
			tokType: tokenizer.CLASS,
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
			name:    "class with field member",
			input:   "class MyClass {\n  +field int\n}",
			tokType: tokenizer.CLASS,
			want: &ast.Entity{
				Identifier: "MyClass",
				Kind:       ast.ClassKind,
				Members: []ast.Member{
					ast.Field{
						Name:       "field",
						Type:       ast.TypeRef{Kind: ast.Void, Name: "int"},
						Visibility: ast.Public,
					},
				},
			},
		},
		{
			name:      "invalid alias",
			input:     "class MyClass as",
			tokType:   tokenizer.CLASS,
			want:      nil,
			expectErr: true,
		},
		{
			name:      "unclosed stereotype",
			input:     "class MyClass <<Stereo",
			tokType:   tokenizer.CLASS,
			want:      nil,
			expectErr: true,
		},
		{
			name:      "unclosed body",
			input:     "class MyClass {",
			tokType:   tokenizer.CLASS,
			want:      nil,
			expectErr: true,
		},
		{
			name:      "invalid member inside body",
			input:     "class MyClass {\n  invalidField\n}",
			tokType:   tokenizer.CLASS,
			want:      nil,
			expectErr: true,
		},
		{
			name:    "class with generic",
			input:   "class List<T>",
			tokType: tokenizer.CLASS,
			want: &ast.Entity{
				Identifier: "List",
				Kind:       ast.ClassKind,
				Generic:    "T",
			},
		},
		{
			name:    "class with multiple generics",
			input:   "class Map<K, V>",
			tokType: tokenizer.CLASS,
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
			require.Equal(t, tc.tokType, tok.Type)

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
		{
			name:        "simple field",
			input:       "name int",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: ast.Field{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "int"},
			},
		},
		{
			name:        "pointer field",
			input:       "name *MyType",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: ast.Field{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "*MyType"},
			},
		},
		{
			name:        "slice field",
			input:       "name []int",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: ast.Field{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "[]int"},
			},
		},
		{
			name:        "simple method",
			input:       "Method()",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want: ast.Method{
				Name: "Method",
			},
		},
		{
			name:        "method with parameters and return",
			input:       "Method(a int) error",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want: ast.Method{
				Name: "Method",
				Parameters: []ast.Parameter{
					{Name: "a", Type: ast.TypeRef{Kind: ast.Void, Name: "int"}},
				},
				ReturnType: []ast.TypeRef{
					{Kind: ast.Void, Name: "error"},
				},
			},
		},
		{
			name:        "method defined as method modifier but no params",
			input:       "{method} Method()",
			entryType:   tokenizer.LBRACE,
			initialName: "",
			want: ast.Method{
				Name:      "Method",
				Modifiers: []string{"method"},
			},
		},
		{
			name:        "both field and method modifiers",
			input:       "{field} {method} Name()",
			entryType:   tokenizer.LBRACE,
			initialName: "",
			want:        nil,
			expectErr:   true,
		},
		{
			name:        "invalid field (no type)",
			input:       "name",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want:        nil,
			expectErr:   true,
		},
		{
			name:        "invalid method (unclosed param parenthesis)",
			input:       "Method(a int",
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
			var initialField ast.Field
			if tc.entryType == tokenizer.LBRACE {
				mod, err := p.stream.TryReadModifier()
				require.NoError(t, err)
				initialField.Modifiers = append(initialField.Modifiers, mod)
				entryTok = tokenizer.Token{Type: tokenizer.LBRACE}
			} else {
				entryTok = p.stream.Emit()
				require.Equal(t, tc.entryType, entryTok.Type)
				if entryTok.Type == tokenizer.IDENTIFIER {
					require.Equal(t, tc.initialName, entryTok.Literal)
				}
			}

			got, err := p.parseFieldOrMethod(entryTok, initialField)
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
			input: "+field int",
			want: ast.Field{
				Name:       "field",
				Type:       ast.TypeRef{Kind: ast.Void, Name: "int"},
				Visibility: ast.Public,
			},
		},
		{
			name:  "private method member",
			input: "-Method()",
			want: ast.Method{
				Name:       "Method",
				Visibility: ast.Private,
			},
		},
		{
			name:  "field with modifier",
			input: "{field} myField int",
			want: ast.Field{
				Name:      "myField",
				Type:      ast.TypeRef{Kind: ast.Void, Name: "int"},
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
