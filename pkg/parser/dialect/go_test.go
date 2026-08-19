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
		expectedParam GoParameter
		expectError   bool
	}{
		{
			name:  "basic primitive type",
			input: "name int",
			expectedParam: GoParameter{
				Name: "name",
				Type: NamedRef("int"),
			},
		},
		{
			name:  "pointer type",
			input: "name *MyType",
			expectedParam: GoParameter{
				Name: "name",
				// Type: GoTypeRef{
				// 	Typ: Pointer,
				// 	Base: &GoTypeRef{
				// 		Typ:  Named,
				// 		Name: "MyType",
				// 	},
				// },
				Type: PointerTo(NamedRef("MyType")),
			},
		},
		{
			name:  "slice type",
			input: "name []int",
			expectedParam: GoParameter{
				Name: "name",
				// Type: GoTypeRef{
				// 	Typ: Slice,
				// 	Base: &GoTypeRef{
				// 		Typ:  Named,
				// 		Name: "int",
				// 	},
				// },
				Type: SliceOf(NamedRef("int")),
			},
		},
		{
			name:  "slice of pointers",
			input: "name []*MyType",
			expectedParam: GoParameter{
				Name: "name",
				Type: SliceOf(PointerTo(NamedRef("MyType"))),
			},
		},
		{
			name:  "pointer to slice",
			input: "name *[]int",
			expectedParam: GoParameter{
				Name: "name",
				Type: PointerTo(SliceOf(NamedRef("int"))),
			},
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
			field, err := g.ParseField(toks, &MemberOptions{Visibility: ast.VisibilityUnknown})
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			goField, ok := field.(*GoField)
			require.True(t, ok, "Returned field is not of type GoField")
			require.Equal(t, tt.expectedParam.Type, goField.Type)
			require.Equal(t, tt.expectedParam.Name, goField.Name)
		})
	}
}

func TestGoDialect_ParseMethod(t *testing.T) {
	g := NewGoDialect()

	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectedRets   []GoParameter
		expectedParams []GoParameter
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
			expectedRets: []GoParameter{
				{Name: "", Type: NamedRef("int")},
			},
			expectedParams: nil,
		},
		{
			name:         "single param no return",
			input:        "Method(a int)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []GoParameter{
				{Name: "a", Type: NamedRef("int")},
			},
		},
		{
			name:         "multiple params no return",
			input:        "Method(a int, b bool)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []GoParameter{
				{Name: "a", Type: NamedRef("int")},
				{Name: "b", Type: NamedRef("bool")},
			},
		},
		{
			name:         "multiple params same type no return (backfilling)",
			input:        "Method(a, b int)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []GoParameter{
				{Name: "a", Type: NamedRef("int")},
				{Name: "b", Type: NamedRef("int")},
			},
		},
		{
			name:         "single param multiple returns",
			input:        "Method(a int) (int, error)",
			expectedName: "Method",
			expectedRets: []GoParameter{
				{Type: NamedRef("int")},
				{Type: NamedRef("error")},
			},
			expectedParams: []GoParameter{
				{Name: "a", Type: NamedRef("int")},
			},
		},
		{
			name:         "pointer param and slice of pointers return",
			input:        "Method(a *MyType) ([]*int, error)",
			expectedName: "Method",
			expectedRets: []GoParameter{
				{Type: SliceOf(PointerTo(NamedRef("int")))},
				{Type: NamedRef("error")},
			},
			expectedParams: []GoParameter{
				{Name: "a", Type: PointerTo(NamedRef("MyType"))},
			},
		},
		{
			name:         "mix of typed and untyped parameters",
			input:        "Method(a int, b, c bool)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []GoParameter{
				{Name: "a", Type: NamedRef("int")},
				{Name: "b", Type: NamedRef("bool")},
				{Name: "c", Type: NamedRef("bool")},
			},
		},
		{
			name:         "unnamed single parameter",
			input:        "Method(int)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []GoParameter{
				{Name: "int", Type: nil},
			},
		},
		{
			name:         "unnamed multiple parameters",
			input:        "Method(int, bool)",
			expectedName: "Method",
			expectedRets: nil,
			expectedParams: []GoParameter{
				{Name: "int", Type: nil},
				{Name: "bool", Type: nil},
			},
		},
		{
			name:         "multiple returns in parentheses",
			input:        "Method() (int, bool)",
			expectedName: "Method",
			expectedRets: []GoParameter{
				{Type: NamedRef("int")},
				{Type: NamedRef("bool")},
			},
			expectedParams: nil,
		},
		{
			name:         "multiple returns in parentheses (with named return)",
			input:        "Method() (result int, e error)",
			expectedName: "Method",
			expectedRets: []GoParameter{
				{Name: "result", Type: NamedRef("int")},
				{Name: "e", Type: NamedRef("error")},
			},
			expectedParams: nil,
		},
		{
			name:         "multiple returns in parentheses (with named return)",
			input:        "Method() (result *int, e error)",
			expectedName: "Method",
			expectedRets: []GoParameter{
				{Name: "result", Type: PointerTo(NamedRef("int"))},
				{Name: "e", Type: NamedRef("error")},
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
			name:        "at least one parameterized return is missing",
			input:       "Method() (res int, error, ok bool)",
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
			method, err := g.ParseMethod(toks, &MemberOptions{Visibility: ast.VisibilityUnknown})
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedName, method.MethodName())

			goMethod, ok := method.(*GoMethod)
			require.Truef(t, ok, "Returned method is not of type GoMethod")
			require.Equal(t, tt.expectedRets, goMethod.ReturnType)
			require.Equal(t, tt.expectedParams, goMethod.Parameters)
		})
	}
}
