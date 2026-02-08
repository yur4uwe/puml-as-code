package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func tokensToStrings(lex *Lexer, max int) []string {
	out := []string{}
	for _ = range max {
		t := lex.NextToken()
		if t.Type == EOF {
			out = append(out, "EOF")
			break
		}
		out = append(out, t.Type.String()+":"+t.Literal)
	}
	return out
}

func TestSimpleClassDeclaration(t *testing.T) {
	src := `class Foo {
  +id : int
  -name : string
  foo() : void
}`
	lex := NewLexer(src)
	// We'll read tokens and ensure certain tokens appear in sequence
	expectedTokens := []string{
		"CLASS:class",
		"IDENTIFIER:Foo",
		"{:{",
		"NEWLINE:\n",
		"VISIBILITY:+",
		"IDENTIFIER:id",
		":::",
		"IDENTIFIER:int",
		"NEWLINE:\n",
		"VISIBILITY:-",
		"IDENTIFIER:name",
		":::",
		"IDENTIFIER:string",
		"NEWLINE:\n",
		"IDENTIFIER:foo",
		"(:(",
		"):)",
		":::",
		"IDENTIFIER:void",
		"NEWLINE:\n",
		"}:}",
		"EOF",
	}
	got := tokensToStrings(lex, 100)
	require.Equal(t, expectedTokens, got, "incorrect tokens")
}

func TestRelationshipLexing(t *testing.T) {
	src := `User "1" -- "0..*" Order : places`
	lex := NewLexer(src)
	// Expect: IDENT:User STRING:"1" RELATIONSHIP:-- STRING:"0..*" IDENT:Order COLON::
	expectedTypes := []string{
		"IDENTIFIER", "STRING", "RELATIONSHIP", "STRING", "IDENTIFIER", ":", "IDENTIFIER",
	}
	got := []string{}
	for _ = range 10 {
		tok := lex.NextToken()
		if tok.Type == EOF {
			break
		}
		got = append(got, tok.Type.String())
	}
	if len(got) < len(expectedTypes) {
		t.Fatalf("got too few tokens: %v", got)
	}
	require.Equal(t, expectedTypes, got, "unexpected token types %v", got)
}
