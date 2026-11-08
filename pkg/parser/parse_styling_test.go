package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/tokenizer/keyword"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"
)

func newDirectParser(input string) (*Parser, tokenizer.Token) {
	p := &Parser{
		stream:  stream.NewTokenStream(input),
		Dialect: dialect.NewGoDialect(),
		ast:     &ast.Diagram{},
	}
	for {
		tok := p.stream.Emit()
		if tok.Type != tokenizer.NEWLINE {
			return p, tok
		}
	}
}

// -----------------------------------------------------------------------------
// Style Block Tests (Table-Driven for Similar Cases)
// -----------------------------------------------------------------------------

func TestParseStyleBlockSelectors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSelectors [][]string
	}{
		{
			name: "single root selector",
			input: `<style>
root {
    FontColor: #333;
}
</style>`,
			wantSelectors: [][]string{{"root"}},
		},
		{
			name: "dot stereotype and qualified class selectors",
			input: `<style>
.highlight {
    BackgroundColor: Yellow;
}
class.admin {
    FontColor: Red;
}
</style>`,
			wantSelectors: [][]string{{".highlight"}, {"class.admin"}},
		},
		{
			name: "comma separated selector lists",
			input: `<style>
class, interface {
    BackgroundColor: lightblue;
}
.primary, .secondary {
    LineThickness: 2;
}
</style>`,
			wantSelectors: [][]string{{"class", "interface"}, {".primary", ".secondary"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := newDirectParser(tc.input)
			sb := p.parseStyleBlock()
			require.NotNil(t, sb)
			require.False(t, p.ast.HasErrors())
			assertBlockClosed(t, sb)
			require.Len(t, sb.Statements, len(tc.wantSelectors))

			for i, wantSel := range tc.wantSelectors {
				rule, ok := sb.Statements[i].(*ast.StyleRule)
				require.True(t, ok)
				assertBlockClosed(t, rule)
				require.Equal(t, wantSel, rule.Selectors)
			}
		})
	}
}

func TestParseStyleBlock_Errors(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		bs             expectedBlockState
		ruleBs         expectedBlockState
		errContains    string
		wantStatements int
	}{
		{
			name:           "unterminated style block",
			input:          "<style>\nroot {\nFontColor: red;\n}\n",
			bs:             expectBlockUnclosed,
			ruleBs:         expectBlockClosed,
			errContains:    "unclosed <style>: expected </style>",
			wantStatements: 1,
		},
		{
			name:           "unterminated rule in style block",
			input:          "<style>\nroot {\nFontColor: red;\n</style>\n",
			bs:             expectBlockClosed,
			ruleBs:         expectBlockUnclosed,
			errContains:    "unclosed style rule: expected }",
			wantStatements: 1,
		},
		{
			name:           "empty selector list",
			input:          "<style>\n, {\nFontColor: red;\n}\n</style>\n",
			bs:             expectBlockClosed,
			ruleBs:         expectBlockClosed,
			errContains:    "empty selector in style rule",
			wantStatements: 1,
		},
		{
			name:           "missing opening brace after style rule",
			input:          "<style>\nroot\n</style>\n",
			bs:             expectBlockClosed,
			errContains:    "Expected opening brace after style rule",
			wantStatements: 1,
		},
		{
			name:           "unexpected token in style rule body",
			input:          "<style>\nroot {\n@invalid\nFontColor: red;\n}\n</style>\n",
			bs:             expectBlockClosed,
			ruleBs:         expectBlockClosed,
			errContains:    "Expected style property name or rule selector",
			wantStatements: 1,
		},
		{
			name:           "missing value after style declaration",
			input:          "<style>\nroot {\nFontColor:\n}\n</style>\n",
			bs:             expectBlockClosed,
			ruleBs:         expectBlockClosed,
			errContains:    "expected value after style declaration",
			wantStatements: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := newDirectParser(tc.input)
			sb := p.parseStyleBlock()
			require.NotNil(t, sb)

			switch tc.bs {
			case expectNoBlock:
				assertBlockBodyless(t, sb)
			case expectBlockClosed:
				assertBlockClosed(t, sb)
			case expectBlockUnclosed:
				assertBlockUnclosed(t, sb)
			}

			require.True(t, p.ast.HasErrors(), "expected parser to record diagnostics on bad syntax")
			require.NotEmpty(t, p.ast.Diagnostics)
			require.Contains(t, p.ast.Diagnostics[0].Message, tc.errContains)
			require.Len(t, sb.Statements, tc.wantStatements)

			if tc.ruleBs != expectNoBlock && len(sb.Statements) > 0 {
				if rule, ok := sb.Statements[0].(*ast.StyleRule); ok {
					switch tc.ruleBs {
					case expectBlockClosed:
						assertBlockClosed(t, rule)
					case expectBlockUnclosed:
						assertBlockUnclosed(t, rule)
					}
				}
			}
		})
	}
}

func TestParseStyleBlock_Recovery(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		bs             expectedBlockState
		wantStatements int
	}{
		{
			name:           "restore to next rule on empty selector",
			input:          "<style>\n, {\nFontColor: red;\n}\nclass {\nBackgroundColor: blue;\n}\n</style>\nclass NextClass",
			bs:             expectBlockClosed,
			wantStatements: 2,
		},
		{
			name:           "restore to next rule on missing opening brace",
			input:          "<style>\nroot\nclass {\nBackgroundColor: blue;\n}\n</style>\nclass NextClass",
			bs:             expectBlockClosed,
			wantStatements: 2,
		},
		{
			name:           "restore past bad token inside rule body",
			input:          "<style>\nroot {\n@invalid\nFontColor: red;\n}\n</style>\nclass NextClass",
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore past missing value in declaration",
			input:          "<style>\nroot {\nFontColor:\nMargin: 10;\n}\n</style>\nclass NextClass",
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore on unclosed rule before style block closer",
			input:          "<style>\nroot {\nFontColor: red;\n</style>\nclass NextClass",
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to EOF on unclosed style block",
			input:          "<style>\nroot {\nFontColor: red;\n}\n",
			bs:             expectBlockUnclosed,
			wantStatements: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := newDirectParser(tc.input)
			sb := p.parseStyleBlock()
			require.NotNil(t, sb)
			require.True(t, p.ast.HasErrors(), "expected parser to record diagnostics on bad syntax")

			switch tc.bs {
			case expectNoBlock:
				assertBlockBodyless(t, sb)
			case expectBlockClosed:
				assertBlockClosed(t, sb)
			case expectBlockUnclosed:
				assertBlockUnclosed(t, sb)
			}

			require.Len(t, sb.Statements, tc.wantStatements)

			if tc.bs == expectBlockUnclosed {
				require.True(t, p.stream.AssertType(tokenizer.EOF), "expected stream to reach EOF on unclosed block")
			} else {
				// Next token must be the separating newline
				require.Equal(t, tokenizer.NEWLINE, p.stream.PeekTokenAt(0).Type, "expected next token to be NEWLINE after recovery")

				// Canary assertion: consume newline and parse subsequent statement
				p.stream.Emit()
				nextTok := p.stream.Emit()
				require.Equal(t, keyword.Class, keyword.Classify(nextTok.Literal))
				nextEnt := p.parseEntity(nextTok)
				require.NotNil(t, nextEnt)
				require.Equal(t, "NextClass", nextEnt.Identifier.EntityName())
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Style Block Structure & Hierarchy Tests (Individual Specialized Tests)
// -----------------------------------------------------------------------------

func TestParseStyleBlockRootDeclarations(t *testing.T) {
	input := `<style>
root {
    FontColor: #333333;
    FontName: Arial;
    Margin: 10;
}
</style>`
	p, _ := newDirectParser(input)
	sb := p.parseStyleBlock()
	require.NotNil(t, sb)
	require.False(t, p.ast.HasErrors())
	assertBlockClosed(t, sb)

	require.Len(t, sb.Statements, 1, "expected single statement in style block")

	rule, ok := sb.Statements[0].(*ast.StyleRule)
	require.True(t, ok)
	assertBlockClosed(t, rule)
	require.Equal(t, []string{"root"}, rule.Selectors)
	require.Len(t, rule.Statements, 3)

	d1 := rule.Statements[0].(*ast.StyleDeclaration)
	require.Equal(t, "FontColor", d1.Property)
	require.Equal(t, "#333333", d1.Value)

	d2 := rule.Statements[1].(*ast.StyleDeclaration)
	require.Equal(t, "FontName", d2.Property)
	require.Equal(t, "Arial", d2.Value)

	d3 := rule.Statements[2].(*ast.StyleDeclaration)
	require.Equal(t, "Margin", d3.Property)
	require.Equal(t, "10", d3.Value)
}

func TestParseStyleDeclarationDelimiters(t *testing.T) {
	input := `<style>
class {
    FontColor: #333333
    FontSize 12;
    BackgroundColor #FF0000
    LineThickness: 2;
}
</style>`
	p, _ := newDirectParser(input)
	sb := p.parseStyleBlock()
	require.NotNil(t, sb)
	require.False(t, p.ast.HasErrors())
	assertBlockClosed(t, sb)

	require.Len(t, sb.Statements, 1)

	rule, ok := sb.Statements[0].(*ast.StyleRule)
	require.True(t, ok)
	assertBlockClosed(t, rule)
	require.Len(t, rule.Statements, 4)

	// Case 1: colon, no semicolon
	d1 := rule.Statements[0].(*ast.StyleDeclaration)
	require.Equal(t, "FontColor", d1.Property)
	require.Equal(t, "#333333", d1.Value)

	// Case 2: no colon, with semicolon
	d2 := rule.Statements[1].(*ast.StyleDeclaration)
	require.Equal(t, "FontSize", d2.Property)
	require.Equal(t, "12", d2.Value)

	// Case 3: no colon, no semicolon
	d3 := rule.Statements[2].(*ast.StyleDeclaration)
	require.Equal(t, "BackgroundColor", d3.Property)
	require.Equal(t, "#FF0000", d3.Value)

	// Case 4: colon and semicolon
	d4 := rule.Statements[3].(*ast.StyleDeclaration)
	require.Equal(t, "LineThickness", d4.Property)
	require.Equal(t, "2", d4.Value)
}

func TestParseStyleBlockNestedRules(t *testing.T) {
	input := `<style>
classDiagram {
    class {
        BackGroundColor: PaleGreen;
        .highlight {
            BackGroundColor: Yellow;
        }
    }
}
</style>`
	p, _ := newDirectParser(input)
	sb := p.parseStyleBlock()
	require.NotNil(t, sb)
	require.False(t, p.ast.HasErrors())
	assertBlockClosed(t, sb)

	require.Len(t, sb.Statements, 1)

	outer := sb.Statements[0].(*ast.StyleRule)
	assertBlockClosed(t, outer)
	require.Equal(t, []string{"classDiagram"}, outer.Selectors)
	require.Len(t, outer.Statements, 1)

	classRule := outer.Statements[0].(*ast.StyleRule)
	assertBlockClosed(t, classRule)
	require.Equal(t, []string{"class"}, classRule.Selectors)
	require.Len(t, classRule.Statements, 2)

	decl := classRule.Statements[0].(*ast.StyleDeclaration)
	require.Equal(t, "BackGroundColor", decl.Property)
	require.Equal(t, "PaleGreen", decl.Value)

	highlightRule := classRule.Statements[1].(*ast.StyleRule)
	assertBlockClosed(t, highlightRule)
	require.Equal(t, []string{".highlight"}, highlightRule.Selectors)
	require.Len(t, highlightRule.Statements, 1)
	hDecl := highlightRule.Statements[0].(*ast.StyleDeclaration)
	require.Equal(t, "BackGroundColor", hDecl.Property)
	require.Equal(t, "Yellow", hDecl.Value)
}

func TestParseStyleBlockDirectivesAndTrivia(t *testing.T) {
	input := `<style>
!include_once common.puml
root {
    ' comment before prop
    FontName: Helvetica; ' inline comment
}
</style>`
	p, _ := newDirectParser(input)
	sb := p.parseStyleBlock()
	require.NotNil(t, sb)
	require.False(t, p.ast.HasErrors())
	assertBlockClosed(t, sb)
	require.Len(t, sb.Statements, 2)

	// Directive
	_, ok := sb.Statements[0].(*ast.IncludeDirective)
	require.True(t, ok)

	// Rule & Declaration with trivia
	rule := sb.Statements[1].(*ast.StyleRule)
	assertBlockClosed(t, rule)
	decl := rule.Statements[0].(*ast.StyleDeclaration)
	require.Equal(t, "FontName", decl.Property)
	require.Equal(t, "Helvetica", decl.Value)
	require.NotEmpty(t, decl.GetLeadingTrivia())
	require.Contains(t, decl.GetLeadingTrivia()[0].Text, "comment before prop")
	require.NotEmpty(t, decl.GetTrailingTrivia())
	require.Contains(t, decl.GetTrailingTrivia()[0].Text, "inline comment")
}
