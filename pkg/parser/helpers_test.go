package parser

import (
	"testing"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

func TestParseVisibilityCommand(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		tokenType      tokenizer.TokenType
		expectedKind   ast.VisibilityCommandKind
		expectedTarget string
	}{
		{
			name:           "hide empty members",
			input:          " empty members\n",
			tokenType:      tokenizer.HIDE,
			expectedKind:   ast.Hide,
			expectedTarget: "empty members",
		},
		{
			name:           "show class Name",
			input:          " class Name\n",
			tokenType:      tokenizer.SHOW,
			expectedKind:   ast.Show,
			expectedTarget: "class Name",
		},
		{
			name:           "remove circle",
			input:          " circle\n",
			tokenType:      tokenizer.REMOVE,
			expectedKind:   ast.Remove,
			expectedTarget: "circle",
		},
		{
			name:           "restore members",
			input:          " members\n",
			tokenType:      tokenizer.RESTORE,
			expectedKind:   ast.Restore,
			expectedTarget: "members",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				stream: tokenizer.NewTokenStream(tt.input),
				ast:    &ast.Diagram{},
			}
			tok := tokenizer.Token{
				Type: tt.tokenType,
			}
			err := p.parseVisibilityCommand(tok)
			require.NoError(t, err)
			require.Len(t, p.ast.Statements, 1, "expected 1 statement, got %d", len(p.ast.Statements))

			cmd, ok := p.ast.Statements[0].(ast.VisibilityCommand)
			require.True(t, ok, "expected VisibilityCommand, got %T", p.ast.Statements[0])
			require.Equal(t, tt.expectedKind, cmd.Kind, "expected kind %v, got %v", tt.expectedKind, cmd.Kind)
			require.Equal(t, tt.expectedTarget, cmd.Target, "expected target %q, got %q", tt.expectedTarget, cmd.Target)
		})
	}
}

func TestParseDiagDirection(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		initialToken      tokenizer.Token
		expectedDirection ast.DirectionCommandKind
		expectError       bool
	}{
		{
			name:  "left to right",
			input: "to right direction",
			initialToken: tokenizer.Token{
				Type:    tokenizer.DIRECTION,
				Literal: "left",
			},
			expectedDirection: ast.LeftToRightDirection,
			expectError:       false,
		},
		{
			name:  "top to bottom",
			input: "to bottom direction",
			initialToken: tokenizer.Token{
				Type:    tokenizer.DIRECTION,
				Literal: "top",
			},
			expectedDirection: ast.TopToBottomDirection,
			expectError:       false,
		},
		{
			name:  "invalid modifier",
			input: "to up direction",
			initialToken: tokenizer.Token{
				Type:    tokenizer.DIRECTION,
				Literal: "left",
			},
			expectError: true,
		},
		{
			name:  "missing 'to'",
			input: "right",
			initialToken: tokenizer.Token{
				Type:    tokenizer.DIRECTION,
				Literal: "left",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				stream: tokenizer.NewTokenStream(tt.input),
				ast:    &ast.Diagram{},
			}
			err := p.parseDiagDirection(tt.initialToken)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
			}

			if !tt.expectError {
				if len(p.ast.Statements) != 1 {
					t.Fatalf("expected 1 statement, got %d", len(p.ast.Statements))
				}

				cmd, ok := p.ast.Statements[0].(ast.DirectionCommand)
				if !ok {
					t.Fatalf("expected DirectionCommand, got %T", p.ast.Statements[0])
				}

				if cmd.Direction != tt.expectedDirection {
					t.Errorf("expected direction %v, got %v", tt.expectedDirection, cmd.Direction)
				}
			}
		})
	}
}

func TestParseScale(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    ast.ScaleCommand
		expectError bool
	}{
		{
			name:  "scale factor",
			input: "1.5",
			expected: ast.ScaleCommand{
				Scale: 1.5,
			},
		},
		{
			name:  "scale factor with slash",
			input: "2/3",
			expected: ast.ScaleCommand{
				Scale: 2.0 / 3.0,
			},
		},
		{
			name:  "scale width",
			input: "200 width",
			expected: ast.ScaleCommand{
				Width: 200,
			},
		},
		{
			name:  "scale height",
			input: "300 height",
			expected: ast.ScaleCommand{
				Height: 300,
			},
		},
		{
			name:  "scale max width",
			input: "max 1024 width",
			expected: ast.ScaleCommand{
				IsMax: true,
				Width: 1024,
			},
		},
		{
			name:  "scale asterisk",
			input: "200*300",
			expected: ast.ScaleCommand{
				Width:  200,
				Height: 300,
			},
		},
		{
			name:  "scale asterisk spaces",
			input: "200 * 300",
			expected: ast.ScaleCommand{
				Width:  200,
				Height: 300,
			},
		},
		{
			name:  "scale x",
			input: "200x300",
			expected: ast.ScaleCommand{
				Width:  200,
				Height: 300,
			},
		},
		{
			name:  "scale max 200x300",
			input: "max 200x300",
			expected: ast.ScaleCommand{
				IsMax:  true,
				Width:  200,
				Height: 300,
			},
		},
		{
			name:  "scale 200 x 300",
			input: "200 x 300",
			expected: ast.ScaleCommand{
				Width:  200,
				Height: 300,
			},
		},
		// Incorrect input to test error handling
		{
			name:        "height and width must be positive integers",
			input:       "0.5 height",
			expectError: true,
		},
		{
			name:        "malformed 'x' scale",
			input:       "200x",
			expectError: true,
		},
		{
			name:        "malformed '/' scale",
			input:       "2/0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				stream: tokenizer.NewTokenStream(tt.input),
				ast:    &ast.Diagram{},
			}
			err := p.parseScale()
			if tt.expectError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			require.Len(t, p.ast.Statements, 1, "expected 1 statement, got %d", len(p.ast.Statements))

			cmd, ok := p.ast.Statements[0].(ast.ScaleCommand)
			require.True(t, ok, "expected ScaleCommand, got %T", p.ast.Statements[0])
			require.Equal(t, tt.expected, cmd)
		})
	}
}

func TestParseCSSLikeStyles(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     string
		expectsError bool
	}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				stream: tokenizer.NewTokenStream(tt.input),
				ast:    &ast.Diagram{},
				styles: make([]string, 0),
			}
			err := p.parseStyles(tokenizer.Token{})
			if tt.expectsError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected, p.styles[0])
		})
	}
}

func TestParseSkinparamStyles(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     ast.Skinparam
		expectsError bool
	}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				stream: tokenizer.NewTokenStream(tt.input),
				ast:    &ast.Diagram{},
				styles: make([]string, 0),
			}
			err := p.parseStyles(tokenizer.Token{})
			if tt.expectsError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected, p.skinparam)
		})
	}
}
