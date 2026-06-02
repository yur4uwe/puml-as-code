package parser

import (
	"testing"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(p.ast.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(p.ast.Statements))
			}

			cmd, ok := p.ast.Statements[0].(ast.VisibilityCommand)
			if !ok {
				t.Fatalf("expected VisibilityCommand, got %T", p.ast.Statements[0])
			}

			if cmd.Kind != tt.expectedKind {
				t.Errorf("expected kind %v, got %v", tt.expectedKind, cmd.Kind)
			}
			if cmd.Target != tt.expectedTarget {
				t.Errorf("expected target %q, got %q", tt.expectedTarget, cmd.Target)
			}
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
