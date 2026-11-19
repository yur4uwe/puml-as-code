package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"
)

func TestParseUnhandled(t *testing.T) {
	newParser := func() *Parser {
		return &Parser{Dialect: dialect.NewGoDialect()}
	}

	t.Run("single-line keywords", func(t *testing.T) {
		t.Run("caption with text", func(t *testing.T) {
			input := "caption Figure 1: Architecture Overview"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			stmt, err := p.parseUnhandled(tok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "caption Figure 1: Architecture Overview", uh.Raw)
		})

		t.Run("bare caption without trailing tokens", func(t *testing.T) {
			input := "caption"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			stmt, err := p.parseUnhandled(tok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "caption", uh.Raw)
		})

		t.Run("sprite single-line", func(t *testing.T) {
			input := "sprite $my_icon my_icon.png"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			stmt, err := p.parseUnhandled(tok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "sprite $my_icon my_icon.png", uh.Raw)
		})
	})

	t.Run("fallback single-line directives", func(t *testing.T) {
		t.Run("!define directive", func(t *testing.T) {
			input := "!define FOO(x) class x"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "!define FOO(x) class x", uh.Raw)
		})

		t.Run("!global directive", func(t *testing.T) {
			input := "!global $VAR = 42"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "!global $VAR = 42", uh.Raw)
		})
	})

	t.Run("blocked directives", func(t *testing.T) {
		t.Run("!function block", func(t *testing.T) {
			input := "!function $my_func($a)\n  !return $a + 1\n!endfunction"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "!function $my_func($a)\n  !return $a + 1\n!endfunction", uh.Raw)
		})

		t.Run("!procedure block", func(t *testing.T) {
			input := "!procedure $my_proc()\n  class GeneratedClass\n!endprocedure"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "!procedure $my_proc()\n  class GeneratedClass\n!endprocedure", uh.Raw)
		})

		t.Run("!definelong block", func(t *testing.T) {
			input := "!definelong MACRO_NAME\n  class InsideMacro\n!enddefinelong"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "!definelong MACRO_NAME\n  class InsideMacro\n!enddefinelong", uh.Raw)
		})

		t.Run("!while block", func(t *testing.T) {
			input := "!while ($val > 0)\n  class C\n!endwhile"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "!while ($val > 0)\n  class C\n!endwhile", uh.Raw)
		})

		t.Run("!foreach block", func(t *testing.T) {
			input := "!foreach $item in $list\n  class Item\n!endfor"
			p := Parser{
				stream:  stream.NewTokenStream(input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.NoError(t, err)

			uh, ok := stmt.(*ast.UnhandledStatement)
			require.True(t, ok)
			require.Equal(t, "!foreach $item in $list\n  class Item\n!endfor", uh.Raw)
		})
	})

	t.Run("nested !if block", func(t *testing.T) {
		input := `!if (A)
  class ClassA
  !if (B)
    class ClassB
  !else
    class ClassElse
  !endif
  class ClassAfterInner
!endif`
		p := Parser{
			stream:  stream.NewTokenStream(input),
			Dialect: dialect.NewGoDialect(),
			ast:     &ast.Diagram{},
		}
		exclTok := p.stream.Emit()
		dirTok := p.stream.Emit()
		stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
		require.NoError(t, err)

		uh, ok := stmt.(*ast.UnhandledStatement)
		require.True(t, ok)
		expected := `!if (A)
  class ClassA
  !if (B)
    class ClassB
  !else
    class ClassElse
  !endif
  class ClassAfterInner
!endif`
		require.Equal(t, expected, uh.Raw)
	})

	t.Run("positive whitelist matching: non-class-diagram syntax is rejected", func(t *testing.T) {
		input := "@startuml\nparticipant Alice\n@enduml"
		p := newParser()
		_, err := p.Parse(input)
		require.Error(t, err)
	})

	t.Run("unhandled statements inside container", func(t *testing.T) {
		input := "@startuml\npackage mypkg {\n  caption Figure in Package\n  sprite $icon icon.png\n}\n@enduml"
		p := newParser()
		diag, err := p.Parse(input)
		require.NoError(t, err)
		require.Len(t, diag.Statements, 3)

		cont, ok := diag.Statements[1].(*ast.Container)
		require.True(t, ok)
		require.Len(t, cont.Statements, 2)

		uh1, ok := cont.Statements[0].(*ast.UnhandledStatement)
		require.True(t, ok)
		require.Equal(t, "caption Figure in Package", uh1.Raw)

		uh2, ok := cont.Statements[1].(*ast.UnhandledStatement)
		require.True(t, ok)
		require.Equal(t, "sprite $icon icon.png", uh2.Raw)
	})

	t.Run("generic opener and closer pattern matching (non-directive and directive)", func(t *testing.T) {
		tok := func(lit string, tt tokenizer.TokenType) tokenizer.Token {
			return tokenizer.Token{Literal: lit, Type: tt}
		}

		// Directives with single and multi-token
		require.True(t, isMatchingOpener([]tokenizer.Token{tok("!", tokenizer.EXCLAMATION), tok("if", tokenizer.IDENTIFIER)}, "!if"))
		require.True(t, isMatchingOpener([]tokenizer.Token{tok("!", tokenizer.EXCLAMATION), tok("while", tokenizer.IDENTIFIER)}, "!while"))
		require.False(t, isMatchingOpener([]tokenizer.Token{tok("!", tokenizer.EXCLAMATION), tok("ifdef", tokenizer.IDENTIFIER)}, "!if"))
		require.False(t, isMatchingOpener([]tokenizer.Token{tok("!", tokenizer.EXCLAMATION), tok("elseif", tokenizer.IDENTIFIER)}, "!if"))

		require.True(t, isMatchingCloser([]tokenizer.Token{tok("!", tokenizer.EXCLAMATION), tok("endif", tokenizer.IDENTIFIER)}, "!endif"))
		require.True(t, isMatchingCloser([]tokenizer.Token{tok("!", tokenizer.EXCLAMATION), tok("end", tokenizer.IDENTIFIER), tok("function", tokenizer.IDENTIFIER)}, "!endfunction"))
		require.False(t, isMatchingCloser([]tokenizer.Token{tok("!", tokenizer.EXCLAMATION), tok("elseif", tokenizer.IDENTIFIER)}, "!endif"))

		// Non-directive openers (sequence/activity/state diagram syntax)
		require.True(t, isMatchingOpener([]tokenizer.Token{tok("alt", tokenizer.IDENTIFIER), tok("x > 0", tokenizer.IDENTIFIER)}, "alt"))
		require.True(t, isMatchingOpener([]tokenizer.Token{tok("loop", tokenizer.IDENTIFIER), tok("1000", tokenizer.NUMBER)}, "loop"))
		require.False(t, isMatchingOpener([]tokenizer.Token{tok("alternative", tokenizer.IDENTIFIER)}, "alt"))

		// Non-directive closers (single word, two words, compound)
		require.True(t, isMatchingCloser([]tokenizer.Token{tok("end", tokenizer.IDENTIFIER)}, "end"))
		require.True(t, isMatchingCloser([]tokenizer.Token{tok("end", tokenizer.IDENTIFIER), tok("note", tokenizer.IDENTIFIER)}, "end note"))
		require.True(t, isMatchingCloser([]tokenizer.Token{tok("endnote", tokenizer.IDENTIFIER)}, "end note"))
		require.True(t, isMatchingCloser([]tokenizer.Token{tok("end", tokenizer.IDENTIFIER), tok("legend", tokenizer.IDENTIFIER)}, "endlegend"))
		require.True(t, isMatchingCloser([]tokenizer.Token{tok("endlegend", tokenizer.IDENTIFIER)}, "endlegend"))

		// Distinct closers: "end" must NOT match compound closers like "end note"
		require.False(t, isMatchingCloser([]tokenizer.Token{tok("end", tokenizer.IDENTIFIER), tok("note", tokenizer.IDENTIFIER)}, "end"))
		require.False(t, isMatchingCloser([]tokenizer.Token{tok("end", tokenizer.IDENTIFIER), tok("legend", tokenizer.IDENTIFIER)}, "end"))
	})
}

func TestParseUnhandled_Recovery(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "unterminated at EOF",
			input:       "!function $f()\n  line 1\n",
			errContains: "unterminated block statement for !function",
		},
		{
			name:        "unterminated hitting @enduml",
			input:       "!function $f()\n  line 1\n@enduml",
			errContains: "unterminated block statement for !function",
		},
		{
			name:        "unterminated inside container hitting scope delimiter",
			input:       "!if (flag)\n    class Inside\n}",
			errContains: "unterminated block statement for !if (hit enclosing scope delimiter)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  stream.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			exclTok := p.stream.Emit()
			dirTok := p.stream.Emit()
			stmt, err := p.parseUnhandledDirective(exclTok, dirTok)
			require.Nil(t, stmt)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains)
		})
	}
}
