package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Skinparam Single-Line Tests (Table-Driven)
// -----------------------------------------------------------------------------

func TestParseSkinparamSingleLine(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantStereo string
		wantValue  string
	}{
		{
			name:      "gradient with backslash preserved",
			input:     "skinparam backgroundColor #red\\blue\n",
			wantName:  "backgroundColor",
			wantValue: "#red\\blue",
		},
		{
			name:      "quotes preserved in string value",
			input:     "skinparam defaultFontName \"Courier New\"\n",
			wantName:  "defaultFontName",
			wantValue: `"Courier New"`,
		},
		{
			name:      "numeric value",
			input:     "skinparam classAttributeIconSize 0\n",
			wantName:  "classAttributeIconSize",
			wantValue: "0",
		},
		{
			name:      "boolean value",
			input:     "skinparam shadowing false\n",
			wantName:  "shadowing",
			wantValue: "false",
		},
		{
			name:      "valueless skinparam",
			input:     "skinparam handwritten\n",
			wantName:  "handwritten",
			wantValue: "",
		},
		{
			name:      "semicolon terminated setting",
			input:     "skinparam backgroundColor red;\n",
			wantName:  "backgroundColor",
			wantValue: "red",
		},
		{
			name:      "dotted name",
			input:     "skinparam sequence.ArrowColor red\n",
			wantName:  "sequence.ArrowColor",
			wantValue: "red",
		},
		{
			name:       "middle stereotype",
			input:      "skinparam class<<Foo>>backgroundColor Wheat\n",
			wantName:   "classbackgroundColor",
			wantStereo: "Foo",
			wantValue:  "Wheat",
		},
		{
			name:       "end stereotype",
			input:      "skinparam class.BackgroundColor<< Bar >> DimGray\n",
			wantName:   "class.BackgroundColor",
			wantStereo: "Bar",
			wantValue:  "DimGray",
		},
		{
			name:       "end stereotype with space",
			input:      "skinparam class.BackgroundColor << Bar >> DimGray\n",
			wantName:   "class.BackgroundColor",
			wantStereo: "Bar",
			wantValue:  "DimGray",
		},
		{
			name:       "stereotype with spaces between words",
			input:      "skinparam class<<foo bar>>backgroundColor Wheat\n",
			wantName:   "classbackgroundColor",
			wantStereo: "foo bar",
			wantValue:  "Wheat",
		},
		{
			name:       "dotted target with middle stereotype",
			input:      "skinparam class.sub<<Foo>>param red\n",
			wantName:   "class.subparam",
			wantStereo: "Foo",
			wantValue:  "red",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, tok := newDirectParser(tc.input)
			stmt, err := p.parseSkinparamRoot(tok)
			require.NoError(t, err)

			setting, ok := stmt.(*ast.SkinparamSetting)
			require.True(t, ok)
			require.Equal(t, "skinparam", setting.Keyword)
			require.Equal(t, tc.wantName, setting.Name.Literal)
			var gotStereo string
			if setting.Stereotype != nil {
				gotStereo = setting.Stereotype.Value.Literal
			}
			require.Equal(t, tc.wantStereo, gotStereo)
			require.Equal(t, tc.wantValue, setting.Value)
		})
	}
}

func TestParseSkinparamErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing target after skinparam",
			input: "skinparam\n",
		},
		{
			name:  "unterminated block error",
			input: "skinparam class {\n BackgroundColor red\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, tok := newDirectParser(tc.input)
			_, err := p.parseSkinparamRoot(tok)
			require.Error(t, err)
		})
	}
}

// -----------------------------------------------------------------------------
// Skinparam Block Tests (Individual Specialized Tests)
// -----------------------------------------------------------------------------

func TestParseSkinparamAnonymousBlock(t *testing.T) {
	input := "skinparam {\n BackgroundColor PaleGreen\n}\n"
	p, tok := newDirectParser(input)
	stmt, err := p.parseSkinparamRoot(tok)
	require.NoError(t, err)

	block, ok := stmt.(*ast.SkinparamBlock)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "skinparam", block.Keyword)
	require.Equal(t, "", block.Name.Literal)
	require.Len(t, block.Statements, 1)

	entry, ok := block.Statements[0].(*ast.SkinparamSetting)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "BackgroundColor", entry.Name.Literal)
	require.Equal(t, "PaleGreen", entry.Value)
}

func TestParseSkinparamBlockWithStereotype(t *testing.T) {
	input := "skinparam class<<Special>> {\n BackgroundColor Gold\n}\n"
	p, tok := newDirectParser(input)
	stmt, err := p.parseSkinparamRoot(tok)
	require.NoError(t, err)

	block, ok := stmt.(*ast.SkinparamBlock)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "class", block.Name.Literal)
	var gotStereo string
	if block.Stereotype != nil {
		gotStereo = block.Stereotype.Value.Literal
	}
	require.Equal(t, "Special", gotStereo)
	require.Len(t, block.Statements, 1)

	entry, ok := block.Statements[0].(*ast.SkinparamSetting)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "BackgroundColor", entry.Name.Literal)
	require.Equal(t, "Gold", entry.Value)
}

func TestParseSkinparamBlockWithComments(t *testing.T) {
	input := `' leading comment
skinparam class { ' opener comment
 ' entry leading comment
 BackgroundColor PaleGreen ' same line comment
 ArrowColor SeaGreen
} ' after close
`
	p, tok := newDirectParser(input)
	stmt, err := p.parseSkinparamRoot(tok)
	require.NoError(t, err)

	block, ok := stmt.(*ast.SkinparamBlock)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "class", block.Name.Literal)
	require.NotEmpty(t, block.GetLeadingTrivia())
	require.Contains(t, block.GetLeadingTrivia()[0].Text, "leading comment")
	require.Len(t, block.Statements, 2)

	// Entry 1
	e1, ok := block.Statements[0].(*ast.SkinparamSetting)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "BackgroundColor", e1.Name.Literal)
	require.Equal(t, "PaleGreen", e1.Value)
	require.NotEmpty(t, e1.GetLeadingTrivia())
	require.Contains(t, e1.GetLeadingTrivia()[0].Text, "entry leading comment")
	require.NotEmpty(t, e1.GetTrailingTrivia())
	require.Contains(t, e1.GetTrailingTrivia()[0].Text, "same line comment")

	// Entry 2
	e2, ok := block.Statements[1].(*ast.SkinparamSetting)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "ArrowColor", e2.Name.Literal)
	require.Equal(t, "SeaGreen", e2.Value)

	// Closing comments
	require.NotEmpty(t, block.GetTrailingTrivia())
}

func TestParseSkinparamNestedAndFlattening(t *testing.T) {
	input := `skinparam class {
 BackgroundColor PaleGreen
 !include_once common.puml
 header {
     FontSize 12
     FontColor DarkGreen
 }
 ArrowColor SeaGreen
}
`
	p, tok := newDirectParser(input)
	stmt, err := p.parseSkinparamRoot(tok)
	require.NoError(t, err)

	block, ok := stmt.(*ast.SkinparamBlock)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "class", block.Name.Literal)
	require.Len(t, block.Statements, 4)

	// BackgroundColor
	s1, ok := block.Statements[0].(*ast.SkinparamSetting)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "BackgroundColor", s1.Name.Literal)

	// Directive
	_, ok = block.Statements[1].(*ast.IncludeDirective)
	require.True(t, ok, "expected skinparam block but got %T", stmt)

	// Nested header block
	nested, ok := block.Statements[2].(*ast.SkinparamBlock)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "header", nested.Name.Literal)
	require.Len(t, nested.Statements, 2)

	// ArrowColor
	s4, ok := block.Statements[3].(*ast.SkinparamSetting)
	require.True(t, ok, "expected skinparam block but got %T", stmt)
	require.Equal(t, "ArrowColor", s4.Name.Literal)

	// Test .Settings() flattening
	settings := block.Settings()
	require.Len(t, settings, 4)
	require.Equal(t, "class.BackgroundColor", settings[0].Key)
	require.Equal(t, "PaleGreen", settings[0].Value)
	require.Equal(t, "class.header.FontSize", settings[1].Key)
	require.Equal(t, "12", settings[1].Value)
	require.Equal(t, "class.header.FontColor", settings[2].Key)
	require.Equal(t, "DarkGreen", settings[2].Value)
	require.Equal(t, "class.ArrowColor", settings[3].Key)
	require.Equal(t, "SeaGreen", settings[3].Value)
}
