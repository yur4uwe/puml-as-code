package stdlib

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/resolver"

	"github.com/stretchr/testify/assert"
)

func TestBuiltinTypes(t *testing.T) {
	assert.True(t, IsBuiltin("error"))
	assert.True(t, IsBuiltin("any"))
	assert.True(t, IsBuiltin("string"))
	assert.True(t, IsBuiltin("int64"))
	assert.False(t, IsBuiltin("MyCustomType"))
	assert.False(t, IsBuiltin("Time"))
}

func TestLookupImportPath(t *testing.T) {
	tests := []struct {
		input        string
		expectedPath string
		expectedOk   bool
	}{
		{"time", "time", true},
		{"context", "context", true},
		{"io", "io", true},
		{"http", "net/http", true},
		{"net/http", "net/http", true},
		{"json", "encoding/json", true},
		{"sql", "database/sql", true},
		{"rand", "crypto/rand", true},
		{"unknown_pkg", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			path, ok := LookupImportPath(tt.input)
			assert.Equal(t, tt.expectedOk, ok)
			if ok {
				assert.Equal(t, tt.expectedPath, path)
			}
		})
	}
}

func TestIsStdlibEntity(t *testing.T) {
	tests := []struct {
		name     string
		ent      *resolver.EntitySymbol
		expected bool
	}{
		{
			name: "builtin error",
			ent: &resolver.EntitySymbol{
				FQN: "error",
				AST: &ast.Entity{Identifier: "error"},
			},
			expected: true,
		},
		{
			name: "time.Time with PackagePath",
			ent: &resolver.EntitySymbol{
				FQN:         "time.Time",
				PackagePath: []string{"time"},
				AST:         &ast.Entity{Identifier: "Time"},
			},
			expected: true,
		},
		{
			name: "net.http.Request with PackagePath",
			ent: &resolver.EntitySymbol{
				FQN:         "net.http.Request",
				PackagePath: []string{"net", "http"},
				AST:         &ast.Entity{Identifier: "Request"},
			},
			expected: true,
		},
		{
			name: "time.Time without PackagePath (dot in FQN)",
			ent: &resolver.EntitySymbol{
				FQN: "time.Time",
				AST: &ast.Entity{Identifier: "Time"},
			},
			expected: true,
		},
		{
			name: "user custom struct",
			ent: &resolver.EntitySymbol{
				FQN:         "models.User",
				PackagePath: []string{"models"},
				AST:         &ast.Entity{Identifier: "User"},
			},
			expected: false,
		},
		{
			name: "user custom root struct",
			ent: &resolver.EntitySymbol{
				FQN: "Order",
				AST: &ast.Entity{Identifier: "Order"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsStdlibEntity(tt.ent))
		})
	}
}
