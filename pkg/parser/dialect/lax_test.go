package dialect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"yur4uwe/pac/pkg/tokenizer"
)

func tokenizeSlice(input string) []tokenizer.Token {
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

func TestStringifyTokenSlice_GoSyntax(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "primitive field",
			input:    "id string",
			expected: "id string",
		},
		{
			name:     "pointer field",
			input:    "user *User",
			expected: "user *User",
		},
		{
			name:     "slice field",
			input:    "tags []string",
			expected: "tags []string",
		},
		{
			name:     "pointer to slice",
			input:    "items *[]string",
			expected: "items *[]string",
		},
		{
			name:     "slice of pointers",
			input:    "users []*User",
			expected: "users []*User",
		},
		{
			name:     "fixed array field",
			input:    "matrix [4][4]float64",
			expected: "matrix [4][4]float64",
		},
		{
			name:     "qualified type field",
			input:    "createdAt time.Time",
			expected: "createdAt time.Time",
		},
		{
			name:     "map field",
			input:    "metadata map[string]any",
			expected: "metadata map[string]any",
		},
		{
			name:     "method without params or return",
			input:    "Close()",
			expected: "Close()",
		},
		{
			name:     "method with single param",
			input:    "FindByID(id string)",
			expected: "FindByID(id string)",
		},
		{
			name:     "method with multiple params",
			input:    "Update(id string, active bool)",
			expected: "Update(id string, active bool)",
		},
		{
			name:     "method with single return type",
			input:    "GetName() string",
			expected: "GetName() string",
		},
		{
			name:     "method with pointer return type",
			input:    "GetUser(id string) *User",
			expected: "GetUser(id string) *User",
		},
		{
			name:     "method with multiple returns",
			input:    "Find(id string) (*User, error)",
			expected: "Find(id string) (*User, error)",
		},
		{
			name:     "method with named returns",
			input:    "Lookup(key string) (val string, ok bool)",
			expected: "Lookup(key string) (val string, ok bool)",
		},
		{
			name:     "method with complex params and returns",
			input:    "Process(ctx context.Context, items []*Item) ([]byte, error)",
			expected: "Process(ctx context.Context, items []*Item) ([]byte, error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks := tokenizeSlice(tt.input)
			actual := StringifyTokenSlice(toks)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringifyTokenSlice_UMLAndTypeScriptSyntax(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "UML colon field",
			input:    "name: string",
			expected: "name: string",
		},
		{
			name:     "UML colon field with spaces around colon in input",
			input:    "name  :   string",
			expected: "name: string",
		},
		{
			name:     "UML field with default number value",
			input:    "count: number = 0",
			expected: "count: number = 0",
		},
		{
			name:     "UML field with default string value",
			input:    `status: string = "active"`,
			expected: `status: string = "active"`,
		},
		{
			name:     "TypeScript array field",
			input:    "items: string[]",
			expected: "items: string[]",
		},
		{
			name:     "UML method with return type",
			input:    "getName(): string",
			expected: "getName(): string",
		},
		{
			name:     "UML method with multiple typed params",
			input:    "calculate(x: int, y: int): int",
			expected: "calculate(x: int, y: int): int",
		},
		{
			name:     "UML generic return type",
			input:    "fetchUser(id: string): Promise<User>",
			expected: "fetchUser(id: string): Promise<User>",
		},
		{
			name:     "generic with multiple type parameters",
			input:    "cache: Map<string, User>",
			expected: "cache: Map<string, User>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks := tokenizeSlice(tt.input)
			actual := StringifyTokenSlice(toks)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringifyTokenSlice_JavaCSharpSyntax(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Java type first field",
			input:    "String name",
			expected: "String name",
		},
		{
			name:     "Java field with initialized value",
			input:    "int maxRetries = 3",
			expected: "int maxRetries = 3",
		},
		{
			name:     "Java generic field",
			input:    "List<String> items",
			expected: "List<String> items",
		},
		{
			name:     "Java array field",
			input:    "byte[] buffer",
			expected: "byte[] buffer",
		},
		{
			name:     "Java method without params",
			input:    "void process()",
			expected: "void process()",
		},
		{
			name:     "Java method with typed params",
			input:    "String formatUser(User user, boolean verbose)",
			expected: "String formatUser(User user, boolean verbose)",
		},
		{
			name:     "Java method returning generic",
			input:    "Optional<User> findById(Long id)",
			expected: "Optional<User> findById(Long id)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks := tokenizeSlice(tt.input)
			actual := StringifyTokenSlice(toks)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringifyTokenSlice_SketchAndEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty slice",
			input:    "",
			expected: "",
		},
		{
			name:     "single identifier field",
			input:    "id",
			expected: "id",
		},
		{
			name:     "untyped method",
			input:    "draw()",
			expected: "draw()",
		},
		{
			name:     "untyped method with untyped params",
			input:    "render(x, y, z)",
			expected: "render(x, y, z)",
		},
		{
			name:     "semicolon terminated member",
			input:    "id: string;",
			expected: "id: string;",
		},
		{
			name:     "multiple assignments or symbols",
			input:    "a = 1",
			expected: "a = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks := tokenizeSlice(tt.input)
			actual := StringifyTokenSlice(toks)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
