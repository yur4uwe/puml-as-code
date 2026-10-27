package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

// assertASTEqual asserts that got and want are equal, ignoring any NodeSpan values.
func assertASTEqual(t *testing.T, want, got any) {
	t.Helper()
	w := stripSpans(want)
	g := stripSpans(got)
	if w != nil && g != nil && reflect.TypeOf(w) != reflect.TypeOf(g) {
		if reflect.TypeOf(w).Kind() == reflect.Pointer && reflect.TypeOf(w).Elem() == reflect.TypeOf(g) {
			w = reflect.ValueOf(w).Elem().Interface()
		} else if reflect.TypeOf(g).Kind() == reflect.Pointer && reflect.TypeOf(g).Elem() == reflect.TypeOf(w) {
			g = reflect.ValueOf(g).Elem().Interface()
		}
	}
	require.Equal(t, w, g)
}

// stripSpans returns a copy of v with all tokenizer.SourceSpan values zeroed out.
func stripSpans(v any) any {
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	return stripValue(val).Interface()
}

func stripValue(val reflect.Value) reflect.Value {
	if !val.IsValid() {
		return val
	}

	switch val.Kind() {
	case reflect.Pointer:
		if val.IsNil() {
			return val
		}
		elem := stripValue(val.Elem())
		newPtr := reflect.New(val.Type().Elem())
		newPtr.Elem().Set(elem)
		return newPtr

	case reflect.Interface:
		if val.IsNil() {
			return val
		}
		return stripValue(val.Elem())

	case reflect.Slice:
		if val.IsNil() {
			return val
		}
		newSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			newSlice.Index(i).Set(stripValue(val.Index(i)))
		}
		return newSlice

	case reflect.Map:
		if val.IsNil() {
			return val
		}
		newMap := reflect.MakeMap(val.Type())
		for _, key := range val.MapKeys() {
			newMap.SetMapIndex(key, stripValue(val.MapIndex(key)))
		}
		return newMap

	case reflect.Struct:
		if val.Type() == reflect.TypeFor[tokenizer.SourceSpan]() || val.Type() == reflect.TypeFor[tokenizer.Pos]() {
			return reflect.Zero(val.Type())
		}

		newStruct := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fieldVal := val.Field(i)

			if field.Name == "NodeSpan" {
				newStruct.Field(i).Set(reflect.Zero(field.Type))
				continue
			}

			if !field.IsExported() {
				continue
			}

			newStruct.Field(i).Set(stripValue(fieldVal))
		}
		return newStruct

	default:
		return val
	}
}

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
