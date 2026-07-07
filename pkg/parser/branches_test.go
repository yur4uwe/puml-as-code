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

func TestParseRelationship(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		firstTargetTok tokenizer.Token
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var p Parser
			gotErr := p.parseRelationship(tt.firstTargetTok)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseRelationship() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseRelationship() succeeded unexpectedly")
			}
		})
	}
}
