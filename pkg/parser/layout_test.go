package parser

import (
	"testing"

	"yur4uwe/pac/internal/testutil"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"

	"github.com/stretchr/testify/require"
)

func TestParseLayoutDirectives(t *testing.T) {
	newParser := func(input string) *Parser {
		return &Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
	}

	parseLayout := func(p *Parser) *ast.TextBlock {
		tok := p.stream.Emit()
		if keyword.Classify(tok.Literal) == keyword.Direction {
			actualKwTok := p.stream.Emit()
			return p.parseLayoutStatement(actualKwTok, &tok)
		}
		return p.parseLayoutStatement(tok, nil)
	}

	t.Run("header single line forms", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected ast.TextBlock
		}{
			{
				name:     "plain header",
				input:    "header My Header\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "My Header"},
			},
			{
				name:     "left header",
				input:    "left header My Left Header\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "My Left Header", HorizontalAlignment: "left"},
			},
			{
				name:     "center header",
				input:    "center header My Centered Header\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "My Centered Header", HorizontalAlignment: "center"},
			},
			{
				name:     "header right (trailing alignment)",
				input:    "header right My Right Header\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "My Right Header", HorizontalAlignment: "right"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				p := newParser(tc.input)
				tb := parseLayout(p)
				require.NotNil(t, tb)
				require.Equal(t, tc.expected.Kind, tb.Kind)
				require.Equal(t, tc.expected.Text, tb.Text)
				require.Equal(t, tc.expected.HorizontalAlignment, tb.HorizontalAlignment)
				require.Equal(t, tc.expected.VerticalAlignment, tb.VerticalAlignment)
			})
		}
	})

	t.Run("header multi-line block forms", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected ast.TextBlock
		}{
			{
				name:     "plain block with endheader",
				input:    "header\n  line 1\n  line 2\nendheader\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "  line 1\n  line 2"},
			},
			{
				name:     "block with end header space",
				input:    "header\n  line 1\nend header\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "  line 1"},
			},
			{
				name:     "preceding alignment block",
				input:    "left header\n  line 1\nendheader\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "  line 1", HorizontalAlignment: "left"},
			},
			{
				name:     "trailing alignment block",
				input:    "header center\n  line 1\nendheader\n",
				expected: ast.TextBlock{Kind: ast.BlockHeader, Text: "  line 1", HorizontalAlignment: "center"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				p := newParser(tc.input)
				tb := parseLayout(p)
				require.NotNil(t, tb)
				require.Equal(t, tc.expected.Kind, tb.Kind)
				require.Equal(t, tc.expected.Text, tb.Text)
				require.Equal(t, tc.expected.HorizontalAlignment, tb.HorizontalAlignment)
				require.Equal(t, tc.expected.VerticalAlignment, tb.VerticalAlignment)
			})
		}
	})

	t.Run("footer forms", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected ast.TextBlock
		}{
			{
				name:     "single line footer",
				input:    "footer Confidential\n",
				expected: ast.TextBlock{Kind: ast.BlockFooter, Text: "Confidential"},
			},
			{
				name:     "center footer single line",
				input:    "center footer Page 1\n",
				expected: ast.TextBlock{Kind: ast.BlockFooter, Text: "Page 1", HorizontalAlignment: "center"},
			},
			{
				name:     "footer right single line",
				input:    "footer right Page 1\n",
				expected: ast.TextBlock{Kind: ast.BlockFooter, Text: "Page 1", HorizontalAlignment: "right"},
			},
			{
				name:     "block footer endfooter",
				input:    "footer\n  Draft\n  v1.0\nendfooter\n",
				expected: ast.TextBlock{Kind: ast.BlockFooter, Text: "  Draft\n  v1.0"},
			},
			{
				name:     "block footer end footer",
				input:    "footer\n  Draft\nend footer\n",
				expected: ast.TextBlock{Kind: ast.BlockFooter, Text: "  Draft"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				p := newParser(tc.input)
				tb := parseLayout(p)
				require.NotNil(t, tb)
				require.Equal(t, tc.expected.Kind, tb.Kind)
				require.Equal(t, tc.expected.Text, tb.Text)
				require.Equal(t, tc.expected.HorizontalAlignment, tb.HorizontalAlignment)
				require.Equal(t, tc.expected.VerticalAlignment, tb.VerticalAlignment)
			})
		}
	})

	t.Run("legend forms", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected ast.TextBlock
		}{
			{
				name:     "legend block with trailing alignment",
				input:    "legend right\n  Short legend\n  multiple lines\nendlegend\n",
				expected: ast.TextBlock{Kind: ast.BlockLegend, Text: "  Short legend\n  multiple lines", HorizontalAlignment: "right"},
			},
			{
				name:     "legend block with end legend space",
				input:    "legend\n  Short legend\nend legend\n",
				expected: ast.TextBlock{Kind: ast.BlockLegend, Text: "  Short legend"},
			},
			{
				name:     "legend block with preceding alignment",
				input:    "right legend\n  Short legend\nendlegend\n",
				expected: ast.TextBlock{Kind: ast.BlockLegend, Text: "  Short legend", HorizontalAlignment: "right"},
			},
			{
				name:     "legend single line",
				input:    "legend Single line note\n",
				expected: ast.TextBlock{Kind: ast.BlockLegend, Text: "Single line note"},
			},
			{
				name:     "legend with vertical and horizontal alignment",
				input:    "legend top left\n  Top left legend\nendlegend\n",
				expected: ast.TextBlock{Kind: ast.BlockLegend, Text: "  Top left legend", VerticalAlignment: "top", HorizontalAlignment: "left"},
			},
			{
				name:     "legend with preceding vertical alignment",
				input:    "bottom legend\n  Bottom legend\nendlegend\n",
				expected: ast.TextBlock{Kind: ast.BlockLegend, Text: "  Bottom legend", VerticalAlignment: "bottom"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				p := newParser(tc.input)
				tb := parseLayout(p)
				require.NotNil(t, tb)
				require.Equal(t, tc.expected.Kind, tb.Kind)
				require.Equal(t, tc.expected.Text, tb.Text)
				require.Equal(t, tc.expected.HorizontalAlignment, tb.HorizontalAlignment)
				require.Equal(t, tc.expected.VerticalAlignment, tb.VerticalAlignment)
			})
		}
	})

	t.Run("caption single line", func(t *testing.T) {
		input := "caption Figure 1: Architecture Overview\n"
		p := newParser(input)
		tok := p.stream.Emit()
		stmt, err := p.parseUnhandled(tok)
		require.NoError(t, err)

		uh, ok := stmt.(*ast.UnhandledStatement)
		require.True(t, ok)
		require.Equal(t, "caption Figure 1: Architecture Overview", uh.Raw)
	})

	t.Run("title single line and block", func(t *testing.T) {
		t.Run("single line title", func(t *testing.T) {
			input := "title My Great Diagram\n"
			p := newParser(input)
			tb := parseLayout(p)
			require.NotNil(t, tb)
			require.Equal(t, "My Great Diagram", p.ast.Title)
			require.Equal(t, ast.BlockTitle, tb.Kind)
			require.Equal(t, "My Great Diagram", tb.Text)
		})

		t.Run("title rejects alignment", func(t *testing.T) {
			input := "center title Centered Title\n"
			p := newParser(input)
			prefixTok := p.stream.Emit()
			kwTok := p.stream.Emit()
			tb := p.parseLayoutStatement(kwTok, &prefixTok)
			require.NotNil(t, tb)
			require.True(t, p.ast.HasErrors())
			require.NotEmpty(t, p.ast.Diagnostics)
			require.Contains(t, p.ast.Diagnostics[0].Message, "Title alignment not supported")
		})

		t.Run("block title", func(t *testing.T) {
			input := "title\n  Line 1\n  Line 2\nend title\n"
			p := newParser(input)
			tb := parseLayout(p)
			require.NotNil(t, tb)
			require.Equal(t, "  Line 1\n  Line 2", p.ast.Title)
			require.Equal(t, ast.BlockTitle, tb.Kind)
			require.Equal(t, "  Line 1\n  Line 2", tb.Text)
		})
	})

	t.Run("diagram direction is not confused with layout alignment", func(t *testing.T) {
		input := "left to right direction\nclass A\n"
		p := NewParser(dialect.NewGoDialect())
		p.IsBoundless = true
		diag, err := p.Parse(input)
		require.NoError(t, err)

		dir, ok := diag.Statements[0].(*ast.DirectionCommand)
		require.True(t, ok)
		require.Equal(t, ast.LeftToRightDirection, dir.Direction)
	})

	t.Run("layout directive inside container", func(t *testing.T) {
		input := `package mypkg {
  legend
    Package legend
  endlegend
  class Foo
}
`
		p := newParser(input)
		tok := p.stream.Emit()
		cont, err := p.parseContainer(tok)
		require.NoError(t, err)
		require.NotNil(t, cont)
		require.Len(t, cont.Statements, 2)

		tb, ok := cont.Statements[0].(*ast.TextBlock)
		require.True(t, ok)
		require.Equal(t, ast.BlockLegend, tb.Kind)
		require.Equal(t, "    Package legend", tb.Text)

		ent, ok := cont.Statements[1].(*ast.Entity)
		require.True(t, ok)
		testutil.SemanticEqual(t, identRef("Foo"), ent.Identifier)
	})
}

func TestParseLayoutDirectives_Errors(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedKind ast.TextBlockKind
		errContains  string
	}{
		{
			name:         "title rejects prefix alignment",
			input:        "center title My Title\n",
			expectedKind: ast.BlockTitle,
			errContains:  "Title alignment not supported",
		},
		{
			name:         "vertical alignment not supported for header",
			input:        "top header My Header\n",
			expectedKind: ast.BlockHeader,
			errContains:  "Vertical alignment only supported for legend",
		},
		{
			name:         "vertical alignment not supported for footer",
			input:        "bottom footer My Footer\n",
			expectedKind: ast.BlockFooter,
			errContains:  "Vertical alignment only supported for legend",
		},
		{
			name:         "duplicate horizontal alignment on header",
			input:        "header left right My Header\n",
			expectedKind: ast.BlockHeader,
			errContains:  "Horizontal alignment already set",
		},
		{
			name:         "duplicate horizontal alignment on footer",
			input:        "footer center left My Footer\n",
			expectedKind: ast.BlockFooter,
			errContains:  "Horizontal alignment already set",
		},
		{
			name:         "duplicate horizontal alignment on legend",
			input:        "legend left right My Legend\n",
			expectedKind: ast.BlockLegend,
			errContains:  "Horizontal alignment already set",
		},
		{
			name:         "duplicate vertical alignment on legend",
			input:        "legend top bottom My Legend\n",
			expectedKind: ast.BlockLegend,
			errContains:  "Vertical alignment already set",
		},
		{
			name:         "unclosed multi-line header block",
			input:        "header\n  line 1\n  line 2\n",
			expectedKind: ast.BlockHeader,
			errContains:  "unclosed header: expected end header",
		},
		{
			name:         "unclosed multi-line footer block",
			input:        "footer\n  line 1\n",
			expectedKind: ast.BlockFooter,
			errContains:  "unclosed footer: expected end footer",
		},
		{
			name:         "unclosed multi-line legend block",
			input:        "legend\n  line 1\n",
			expectedKind: ast.BlockLegend,
			errContains:  "unclosed legend: expected end legend",
		},
		{
			name:         "unclosed multi-line title block",
			input:        "title\n  line 1\n",
			expectedKind: ast.BlockTitle,
			errContains:  "unclosed title: expected end title",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  stream.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			var tb *ast.TextBlock
			tok := p.stream.Emit()
			if keyword.Classify(tok.Literal) == keyword.Direction {
				actualKwTok := p.stream.Emit()
				tb = p.parseLayoutStatement(actualKwTok, &tok)
			} else {
				tb = p.parseLayoutStatement(tok, nil)
			}
			require.NotNil(t, tb)
			require.Equal(t, tc.expectedKind, tb.Kind)
			require.True(t, p.ast.HasErrors())
			require.NotEmpty(t, p.ast.Diagnostics)
			require.Contains(t, p.ast.Diagnostics[0].Message, tc.errContains)
		})
	}
}

func TestParseLayoutDirectives_Recovery(t *testing.T) {
	t.Run("single line header with duplicate alignment preserves text block and parses subsequent entity", func(t *testing.T) {
		input := "header left right My Header\nclass User\n"

		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		tok := p.stream.Emit()
		tb := p.parseLayoutStatement(tok, nil)

		require.NotNil(t, tb)
		require.Equal(t, ast.BlockHeader, tb.Kind)
		require.Equal(t, "My Header", tb.Text)
		require.Equal(t, "left", tb.HorizontalAlignment)

		require.True(t, p.ast.HasErrors())
		require.Len(t, p.ast.Diagnostics, 1)
		require.Contains(t, p.ast.Diagnostics[0].Message, "Horizontal alignment already set")

		require.Equal(t, tokenizer.NEWLINE, p.stream.PeekTokenAt(0).Type, "expected next token to be NEWLINE after recovery")
		p.stream.Emit()
		nextTok := p.stream.Emit()
		require.Equal(t, keyword.Class, keyword.Classify(nextTok.Literal))
		ent, err := p.parseEntity(nextTok)
		require.NoError(t, err)
		require.NotNil(t, ent)
		require.Equal(t, "User", ent.Identifier.EntityName())
	})

	t.Run("title with rejected alignment sets diagram title and parses subsequent entity", func(t *testing.T) {
		input := "center title App Architecture\nclass Service\n"

		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		prefixTok := p.stream.Emit()
		kwTok := p.stream.Emit()
		tb := p.parseLayoutStatement(kwTok, &prefixTok)

		require.NotNil(t, tb)
		require.Equal(t, ast.BlockTitle, tb.Kind)
		require.Equal(t, "App Architecture", tb.Text)
		require.Equal(t, "App Architecture", p.ast.Title)

		require.True(t, p.ast.HasErrors())
		require.Len(t, p.ast.Diagnostics, 1)
		require.Contains(t, p.ast.Diagnostics[0].Message, "Title alignment not supported")

		require.Equal(t, tokenizer.NEWLINE, p.stream.PeekTokenAt(0).Type, "expected next token to be NEWLINE after recovery")
		p.stream.Emit()
		nextTok := p.stream.Emit()
		require.Equal(t, keyword.Class, keyword.Classify(nextTok.Literal))
		ent, err := p.parseEntity(nextTok)
		require.NoError(t, err)
		require.NotNil(t, ent)
		require.Equal(t, "Service", ent.Identifier.EntityName())
	})

	t.Run("footer with invalid vertical alignment preserves footer and parses subsequent entity", func(t *testing.T) {
		input := "top footer Confidential\nclass Data\n"

		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		prefixTok := p.stream.Emit()
		kwTok := p.stream.Emit()
		tb := p.parseLayoutStatement(kwTok, &prefixTok)

		require.NotNil(t, tb)
		require.Equal(t, ast.BlockFooter, tb.Kind)
		require.Equal(t, "Confidential", tb.Text)

		require.True(t, p.ast.HasErrors())
		require.Len(t, p.ast.Diagnostics, 1)
		require.Contains(t, p.ast.Diagnostics[0].Message, "Vertical alignment only supported for legend")

		require.Equal(t, tokenizer.NEWLINE, p.stream.PeekTokenAt(0).Type, "expected next token to be NEWLINE after recovery")
		p.stream.Emit()
		nextTok := p.stream.Emit()
		require.Equal(t, keyword.Class, keyword.Classify(nextTok.Literal))
		ent, err := p.parseEntity(nextTok)
		require.NoError(t, err)
		require.NotNil(t, ent)
		require.Equal(t, "Data", ent.Identifier.EntityName())
	})

	t.Run("unclosed multi-line block preserves text block node without panicking", func(t *testing.T) {
		input := "header\n  Header Line 1\n  Header Line 2\n"

		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		kwTok := p.stream.Emit()
		tb := p.parseLayoutStatement(kwTok, nil)

		require.NotNil(t, tb)
		require.Equal(t, ast.BlockHeader, tb.Kind)

		require.True(t, p.ast.HasErrors())
		require.Len(t, p.ast.Diagnostics, 1)
		require.Contains(t, p.ast.Diagnostics[0].Message, "unclosed header: expected end header")
		require.True(t, p.stream.AssertType(tokenizer.EOF), "expected stream to reach EOF on unclosed block")
	})

	t.Run("layout statement with alignment error inside container preserves text block and container children", func(t *testing.T) {
		input := `package mypkg {
  header left right InPackageHeader
  class InnerService
}
`

		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		tok := p.stream.Emit()
		cont, err := p.parseContainer(tok)
		require.NoError(t, err)

		require.NotNil(t, cont)
		require.True(t, p.ast.HasErrors())
		require.Len(t, p.ast.Diagnostics, 1)
		require.Contains(t, p.ast.Diagnostics[0].Message, "Horizontal alignment already set")

		require.Len(t, cont.Statements, 2)

		tb, ok := cont.Statements[0].(*ast.TextBlock)
		require.True(t, ok)
		require.Equal(t, ast.BlockHeader, tb.Kind)
		require.Equal(t, "InPackageHeader", tb.Text)

		ent, ok := cont.Statements[1].(*ast.Entity)
		require.True(t, ok)
		require.Equal(t, "InnerService", ent.Identifier.EntityName())
	})

	t.Run("multiple layout errors in same diagram collects all diagnostics and retains all statements", func(t *testing.T) {
		input := `header left right BadHeader
top footer BadFooter
center title BadTitle
class FinalEntity
`

		p := NewParser(dialect.NewGoDialect())
		p.IsBoundless = true
		diag, err := p.Parse(input)

		require.Error(t, err)
		require.NotNil(t, diag)
		require.True(t, diag.HasErrors())
		require.Len(t, diag.Diagnostics, 3)

		require.Len(t, diag.Statements, 4)

		h, ok := diag.Statements[0].(*ast.TextBlock)
		require.True(t, ok)
		require.Equal(t, ast.BlockHeader, h.Kind)
		require.Equal(t, "BadHeader", h.Text)

		f, ok := diag.Statements[1].(*ast.TextBlock)
		require.True(t, ok)
		require.Equal(t, ast.BlockFooter, f.Kind)
		require.Equal(t, "BadFooter", f.Text)

		ti, ok := diag.Statements[2].(*ast.TextBlock)
		require.True(t, ok)
		require.Equal(t, ast.BlockTitle, ti.Kind)
		require.Equal(t, "BadTitle", ti.Text)

		ent, ok := diag.Statements[3].(*ast.Entity)
		require.True(t, ok)
		require.Equal(t, "FinalEntity", ent.Identifier.EntityName())
	})
}
