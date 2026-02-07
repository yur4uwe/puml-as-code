package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToRelationType(t *testing.T) {
	cases := []struct {
		in  string
		out RelationType
		sig int
	}{
		{"--", Association, 0},
		{"<--", Association, -1},
		{"-->", Association, 1},
		{"o--", Aggregation, -1},
		{"--o", Aggregation, 1},
		{"*--", Composition, -1},
		{"--*", Composition, 1},
		{"<|--", Inheritance, -1},
		{"--|>", Inheritance, 1},
		{"<..", Realization, -1},
		{"..>", Realization, 1},
		{"<..>", Dependency, 0},
		{"unknown", UnknownRelation, 0},
	}

	for _, c := range cases {
		rt, sig := ToRelationType(c.in)
		require.Equal(t, c.out, rt, "input=%q", c.in)
		require.Equal(t, c.sig, sig, "input=%q", c.in)
	}
}

func TestGetVisibility(t *testing.T) {
	require.Equal(t, UnknownVisibility, GetVisibility(""), "empty should be unknown")
	require.Equal(t, Public, GetVisibility("+name"))
	require.Equal(t, Private, GetVisibility("-f"))
	require.Equal(t, Protected, GetVisibility("#x"))
	require.Equal(t, Package, GetVisibility("~p"))
	require.Equal(t, UnknownVisibility, GetVisibility("x")) // not a visibility marker
}

func TestToValueType(t *testing.T) {
	require.Equal(t, Void, ToValueType("void"))
	require.Equal(t, Int, ToValueType("INT"))
	require.Equal(t, String, ToValueType(" String "))
	require.Equal(t, Float, ToValueType("float"))
	require.Equal(t, Bool, ToValueType("bool"))
	require.Equal(t, Custom, ToValueType("MyType"))
	require.Equal(t, UnknownType, ToValueType(""))
}

func TestParseMultiplicity(t *testing.T) {
	cases := []struct {
		in        string
		want      Multiplicity
		shouldErr bool
	}{
		{`"*"`, Multiplicity{Raw: "*", Min: 0, Max: -1}, false},
		{`*`, Multiplicity{Raw: "*", Min: 0, Max: -1}, false},
		{`"1"`, Multiplicity{Raw: "1", Min: 1, Max: 1}, false},
		{`1`, Multiplicity{Raw: "1", Min: 1, Max: 1}, false},
		{`"0..*"`, Multiplicity{Raw: "0..*", Min: 0, Max: -1}, false},
		{`0..*`, Multiplicity{Raw: "0..*", Min: 0, Max: -1}, false},
		{`"1..4"`, Multiplicity{Raw: "1..4", Min: 1, Max: 4}, false},
		{"", UnknownMultiplicity, false},
		{"abc", UnknownMultiplicity, true},
	}

	for _, c := range cases {
		m, err := ParseMultiplicity(c.in)
		if c.shouldErr {
			require.Error(t, err, "input=%q", c.in)
			continue
		}
		require.NoError(t, err, "input=%q", c.in)
		require.Equal(t, c.want.Raw, m.Raw, "raw mismatch for %q", c.in)
		require.Equal(t, c.want.Min, m.Min, "min mismatch for %q", c.in)
		require.Equal(t, c.want.Max, m.Max, "max mismatch for %q", c.in)
	}
}

func TestMultiplicityString(t *testing.T) {
	m, _ := ParseMultiplicity(`"*"`)
	require.Equal(t, "*", m.String())

	m, _ = ParseMultiplicity(`"1"`)
	require.Equal(t, "1", m.String())

	m, _ = ParseMultiplicity(`"1..*"`)
	require.Equal(t, "1..*", m.String())

	m, _ = ParseMultiplicity(`"1..4"`)
	require.Equal(t, "1..4", m.String())

	require.Equal(t, "", UnknownMultiplicity.String())
}
