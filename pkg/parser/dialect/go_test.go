package dialect

import (
	"testing"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

func tokenize(input string) []tokenizer.Token {
	lex := tokenizer.NewLexer(input)
	var tokens []tokenizer.Token
	for {
		tok := lex.Emit()
		if tok.Type == tokenizer.EOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func TestGoDialect_ParseField(t *testing.T) {
	g := NewGoDialect()

	tests := []struct {
		name          string
		input         string
		expectedParam ast.Parameter
		expectError   bool
	}{
		{
			name:  "basic primitive type",
			input: "name int",
			expectedParam: ast.Parameter{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "int"},
			},
		},
		{
			name:  "pointer type",
			input: "name *MyType",
			expectedParam: ast.Parameter{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "*MyType"},
			},
		},
		{
			name:  "slice type",
			input: "name []int",
			expectedParam: ast.Parameter{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "[]int"},
			},
		},
		{
			name:  "slice of pointers",
			input: "name []*MyType",
			expectedParam: ast.Parameter{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "[]*MyType"},
			},
		},
		{
			name:  "pointer to slice",
			input: "name *[]int",
			expectedParam: ast.Parameter{
				Name: "name",
				Type: ast.TypeRef{Kind: ast.Void, Name: "*[]int"},
			},
		},
		{
			name:        "invalid primitive string token",
			input:       "name string",
			expectError: true,
		},
		{
			name:        "invalid colon separator",
			input:       "name : int",
			expectError: true,
		},
		{
			name:        "too few tokens",
			input:       "MyType",
			expectError: true,
		},
		{
			name:        "unexpected trailing tokens",
			input:       "foo bar baz",
			expectError: true,
		},
		{
			name:        "map type not supported",
			input:       "name map[string]int",
			expectError: true,
		},
		{
			name:        "chan type not supported",
			input:       "name chan bool",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks := tokenize(tt.input)
			param, err := g.ParseField(toks)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedParam, param)
			}
		})
	}
}

func TestGoDialect_ParseMethod(t *testing.T) {
	g := NewGoDialect()

	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectedRets   []ast.TypeRef
		expectedParams []ast.Parameter
		expectError    bool
	}{
		{
			name:           "no params no return",
			input:          "Method()",
			expectedName:   "Method",
			expectedRets:   nil,
			expectedParams: nil,
		},
		{
			name:         "no params single return",
			input:        "Method() int",
			expectedName: "Method",
			expectedRets: []ast.TypeRef{
				{Kind: ast.Void, Name: "int"},
			},
			expectedParams: nil,
		},
		{
			name:         "single param no return",
			input:        "Method(a int)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []ast.Parameter{
				{Name: "a", Type: ast.TypeRef{Kind: ast.Void, Name: "int"}},
			},
		},
		{
			name:         "multiple params no return",
			input:        "Method(a int, b bool)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []ast.Parameter{
				{Name: "a", Type: ast.TypeRef{Kind: ast.Void, Name: "int"}},
				{Name: "b", Type: ast.TypeRef{Kind: ast.Void, Name: "bool"}},
			},
		},
		{
			name:         "multiple params same type no return (backfilling)",
			input:        "Method(a, b int)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []ast.Parameter{
				{Name: "a", Type: ast.TypeRef{Kind: ast.Void, Name: "int"}},
				{Name: "b", Type: ast.TypeRef{Kind: ast.Void, Name: "int"}},
			},
		},
		{
			name:         "single param multiple returns",
			input:        "Method(a int) (int, error)",
			expectedName: "Method",
			expectedRets: []ast.TypeRef{
				{Kind: ast.Void, Name: "int"},
				{Kind: ast.Void, Name: "error"},
			},
			expectedParams: []ast.Parameter{
				{Name: "a", Type: ast.TypeRef{Kind: ast.Void, Name: "int"}},
			},
		},
		{
			name:         "pointer param and slice of pointers return",
			input:        "Method(a *MyType) ([]*int, error)",
			expectedName: "Method",
			expectedRets: []ast.TypeRef{
				{Kind: ast.Void, Name: "[]*int"},
				{Kind: ast.Void, Name: "error"},
			},
			expectedParams: []ast.Parameter{
				{Name: "a", Type: ast.TypeRef{Kind: ast.Void, Name: "*MyType"}},
			},
		},
		{
			name:         "mix of typed and untyped parameters",
			input:        "Method(a int, b, c bool)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []ast.Parameter{
				{Name: "a", Type: ast.TypeRef{Kind: ast.Void, Name: "int"}},
				{Name: "b", Type: ast.TypeRef{Kind: ast.Void, Name: "bool"}},
				{Name: "c", Type: ast.TypeRef{Kind: ast.Void, Name: "bool"}},
			},
		},
		{
			name:         "unnamed single parameter",
			input:        "Method(int)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []ast.Parameter{
				{Name: "int", Type: ast.TypeRef{Kind: ast.Void, Name: ""}},
			},
		},
		{
			name:         "unnamed multiple parameters",
			input:        "Method(int, bool)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []ast.Parameter{
				{Name: "int", Type: ast.TypeRef{Kind: ast.Void, Name: ""}},
				{Name: "bool", Type: ast.TypeRef{Kind: ast.Void, Name: ""}},
			},
		},
		{
			name:         "multiple returns in parentheses",
			input:        "Method() (int, bool)",
			expectedName: "Method",
			expectedRets: []ast.TypeRef{
				{Kind: ast.Void, Name: "int"},
				{Kind: ast.Void, Name: "bool"},
			},
			expectedParams: nil,
		},
		{
			name:        "receiver not allowed",
			input:       "(r *MyStruct) Method()",
			expectError: true,
		},
		{
			name:        "missing parentheses",
			input:       "Method",
			expectError: true,
		},
		{
			name:        "unclosed parenthesis",
			input:       "Method(a int",
			expectError: true,
		},
		{
			name:        "missing parameter comma",
			input:       "Method(a int b bool)",
			expectError: true,
		},
		{
			name:        "trailing comma not supported",
			input:       "Method(a int, )",
			expectError: true,
		},
		{
			name:        "leading comma in parameters",
			input:       "Method(, a int)",
			expectError: true,
		},
		{
			name:        "double commas in parameters",
			input:       "Method(a int,, b bool)",
			expectError: true,
		},
		{
			name:        "trailing comma in returns",
			input:       "Method() (int, bool,)",
			expectError: true,
		},
		{
			name:        "double commas in returns",
			input:       "Method() (int,, bool)",
			expectError: true,
		},
		{
			name:        "unclosed return parenthesis",
			input:       "Method() (int",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks := tokenize(tt.input)
			name, rets, params, err := g.ParseMethod(toks)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedName, name)
				require.Equal(t, tt.expectedRets, rets)
				require.Equal(t, tt.expectedParams, params)
			}
		})
	}
}
