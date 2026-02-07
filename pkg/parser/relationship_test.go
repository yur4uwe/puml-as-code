package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"

	"github.com/stretchr/testify/require"
)

func TestParseRelationshipDirect(t *testing.T) {
	classes := map[string]*ast.Class{
		"User":  {Name: "User"},
		"Order": {Name: "Order"},
	}

	line := `User "1" --> "0..*" Order : places`
	rel, err := ParseRelationship(line, classes)
	require.NoError(t, err)
	require.NotNil(t, rel)
	require.Equal(t, "User", rel.From.Name)
	require.Equal(t, "Order", rel.To.Name)
}

func TestParseClassDiagramRelationships(t *testing.T) {
	content := `
@startuml
class Car
class Garage
Car *-- Engine : has
Garage "1" --> "0..*" Car : stores
@enduml
`
	cd, err := ParseClassDiagram(content)
	require.NoError(t, err)
	require.NotNil(t, cd)

	// relationship registered under both endpoints by the repo
	relsCar, ok := cd.Relationships.GetRelationships("Car")
	require.True(t, ok)
	require.GreaterOrEqual(t, len(relsCar), 1)

	// ensure one relation connects Car -> Engine (or vice-versa depending on ParseRelationship mapping)
	expectedCarEngRel := ast.Relationship{
		From:             cd.Classes["Engine"],
		To:               cd.Classes["Car"],
		Type:             ast.Composition,
		MultiplicityFrom: ast.UnknownMultiplicity,
		MultiplicityTo:   ast.UnknownMultiplicity,
		Comment:          "has",
	}
	require.Equal(t, expectedCarEngRel, *relsCar[0], "expected a relationship between Car and Engine")
	expectedCarGarageRel := ast.Relationship{

		From:             cd.Classes["Garage"],
		To:               cd.Classes["Car"],
		Type:             ast.Association,
		MultiplicityFrom: ast.Multiplicity{Raw: "1", Min: 1, Max: 1},
		MultiplicityTo:   ast.Multiplicity{Raw: "0..*", Min: 0, Max: -1},
		Comment:          "stores",
	}
	require.Equal(t, expectedCarGarageRel, *relsCar[1], "expected a relationship between Car and Garage")

	// relation should also be discoverable from Garage side
	relsGarage, ok := cd.Relationships.GetRelationships("Garage")
	require.True(t, ok)
	require.GreaterOrEqual(t, len(relsGarage), 1)

	expectedRel := ast.Relationship{
		From:             cd.Classes["Garage"],
		To:               cd.Classes["Car"],
		Type:             ast.Association,
		MultiplicityFrom: ast.Multiplicity{Raw: "1", Min: 1, Max: 1},
		MultiplicityTo:   ast.Multiplicity{Raw: "0..*", Min: 0, Max: -1},
		Comment:          "stores",
	}
	require.Equal(t, expectedRel, *(relsGarage[0]), "expected relationship from Garage to Car to match")
}
