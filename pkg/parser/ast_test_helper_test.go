package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func identRef(name string, pkg ...string) ast.TargetRef {
	var pkgIdents []ast.Ident
	for _, p := range pkg {
		pkgIdents = append(pkgIdents, ast.Ident{Literal: p})
	}
	return ast.TargetRef{
		PackagePath: pkgIdents,
		Name:        ast.Ident{Literal: name},
	}
}

func identRefPtr(name string, pkg ...string) *ast.TargetRef {
	ref := identRef(name, pkg...)
	return &ref
}

func identRefMember(name string, member string, pkg ...string) ast.TargetRef {
	ref := identRef(name, pkg...)
	ref.Member = &ast.MemberRef{
		Name: ast.Ident{Literal: member},
	}
	return ref
}

func identRefMethod(name, method string, sig []tokenizer.Token, pkg ...string) ast.TargetRef {
	ref := identRef(name, pkg...)
	ref.Member = &ast.MemberRef{
		Name:      tokIdent(method),
		Signature: sig,
	}
	return ref
}

func identRefStringMember(name string, memberStr string, pkg ...string) ast.TargetRef {
	ref := identRef(name, pkg...)
	ref.Member = &ast.MemberRef{
		Name: tokString(memberStr),
	}
	return ref
}

func tokIdent(lit string) ast.Ident {
	return ast.Ident{Literal: lit}
}

func tokString(lit string) ast.Ident {
	if !strings.HasPrefix(lit, "\"") || !strings.HasSuffix(lit, "\"") {
		lit = "\"" + lit + "\""
	}
	return ast.Ident{Literal: lit}
}

func tokStereo(lit string) *ast.Stereotype {
	return &ast.Stereotype{
		Value: ast.Ident{Literal: lit},
	}
}

type expectedDiag struct {
	Severity ast.Severity
	Msg      string
	Span     tokenizer.SourceSpan
}

func assertDiagsEqual(t *testing.T, expected []expectedDiag, actual ast.DiagnosticList, brokenNode ast.Node) {
	t.Helper()
	require.Len(t, actual, len(expected))
	for i, wantD := range expected {
		gotD := actual[i]
		require.Equal(t, wantD.Severity, gotD.Severity)
		require.Contains(t, gotD.Message, wantD.Msg)
		if wantD.Span != (tokenizer.SourceSpan{}) {
			require.Equal(t, wantD.Span, gotD.Span)
		}
		require.Equal(t, brokenNode, gotD.Node)
	}
}

func assertBlockClosed(t *testing.T, node ast.Node) {
	t.Helper()
	b, ok := node.(ast.Block)
	require.True(t, ok, "expected node %T to implement ast.Block", node)
	require.True(t, b.HasBody(), "expected block to have an opened body")
	require.True(t, b.IsClosed(), "expected block to be closed")
}

// assertBlockUnclosed verifies that a node opened a body but reached EOF/boundary without closing.
func assertBlockUnclosed(t *testing.T, node ast.Node) {
	t.Helper()
	b, ok := node.(ast.Block)
	require.True(t, ok, "expected node %T to implement ast.Block", node)
	require.True(t, b.HasBody(), "expected block to have an opened body")
	require.False(t, b.IsClosed(), "expected block to remain unclosed")
}

// assertBlockBodyless verifies that a node declared no body.
func assertBlockBodyless(t *testing.T, node ast.Node) {
	t.Helper()
	b, ok := node.(ast.Block)
	require.True(t, ok, "expected node %T to implement ast.Block", node)
	require.False(t, b.HasBody(), "expected block to have no body")
	require.True(t, b.IsClosed(), "expected bodyless block to be considered closed")
}
