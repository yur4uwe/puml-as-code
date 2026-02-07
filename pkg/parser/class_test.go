package parser

import (
	"testing"

	"yur4uwe/pac/pkg/parser/ast"

	"github.com/stretchr/testify/require"
)

func TestEmptyClass(t *testing.T) {
	content := `
@startuml No relationships Diagram
class Alpha
class Beta
class Gamma
@enduml`

	classes, err := ParseClassDiagram(content)
	require.NoError(t, err)

	expected := ast.ClassDiagram{
		Classes: map[string]*ast.Class{
			"Alpha": {Name: "Alpha"},
			"Beta":  {Name: "Beta"},
			"Gamma": {Name: "Gamma"},
		},
		Relationships: make(ast.RelationshipsRepo),
	}
	require.Equal(t, expected, *classes)
}

func TestClassWithAttributesAndMethods(t *testing.T) {
	content := `
@startuml Single Inheritance Diagram
class Animal {
  +name: string
  +age: int
  +speak(): void
  -eat(food: string): void
}
@enduml`

	classes, err := ParseClassDiagram(content)
	require.NoError(t, err)

	expected := ast.ClassDiagram{
		Classes: map[string]*ast.Class{
			"Animal": {
				Name: "Animal",
				Attributes: []ast.Attribute{
					{Name: "name", Type: ast.ParseTypeRef("string"), Visibility: ast.Public},
					{Name: "age", Type: ast.ParseTypeRef("int"), Visibility: ast.Public},
				},
				Methods: []ast.Method{
					{Name: "speak", ReturnType: ast.ParseTypeRef("void"), Visibility: ast.Public, Parameters: []ast.Attribute{}},
					{Name: "eat", ReturnType: ast.ParseTypeRef("void"), Visibility: ast.Private, Parameters: []ast.Attribute{
						{Name: "food", Type: ast.ParseTypeRef("string"), Visibility: ast.Public},
					},
					},
				},
			},
		},
		Relationships: make(ast.RelationshipsRepo),
	}

	require.Equal(t, expected, *classes)
}

func TestCombinedRelationships(t *testing.T) {
	content := `
@startuml
class Vehicle
class Engine {
  +start(): bool
}
class Wheel {
  +size: int
}
class Car {
  -vin: string
  +Drive(destination: string): error
}
class Garage {
  +location: string
}
class Person {
  +name: string
  +DriveCar(c: Car): error
}

Vehicle <|-- Car
Car *-- Engine : has
Car o-- "4" Wheel : uses
Garage "1" --> "0..*" Car : stores
Person ..> Car : drives
@enduml
`
	cd, err := ParseClassDiagram(content)
	require.NoError(t, err)
	require.NotNil(t, cd)

	// classes existence
	for _, name := range []string{"Vehicle", "Car", "Engine", "Wheel", "Garage", "Person"} {
		_, ok := cd.Classes[name]
		require.True(t, ok, "expected class %s to exist", name)
	}

	// Car should participate in multiple relationships
	relsCar, ok := cd.Relationships.GetRelationships("Car")
	require.True(t, ok)
	require.GreaterOrEqual(t, len(relsCar), 3)

	// Check presence of composition Car <-> Engine with comment "has"
	expectedCarEngRel := &ast.Relationship{
		From:             cd.Classes["Engine"],
		To:               cd.Classes["Car"],
		Type:             ast.Composition,
		MultiplicityFrom: ast.UnknownMultiplicity,
		MultiplicityTo:   ast.UnknownMultiplicity,
		Comment:          "has",
	}
	require.Equal(t, relsCar[1], expectedCarEngRel, "expected a relationship between Car and Engine")

	expectedAgg := &ast.Relationship{
		From:             cd.Classes["Wheel"],
		To:               cd.Classes["Car"],
		Type:             ast.Aggregation,
		MultiplicityFrom: ast.Multiplicity{Raw: "4", Min: 4, Max: 4},
		MultiplicityTo:   ast.UnknownMultiplicity,
		Comment:          "uses",
	}
	require.Equal(t, relsCar[2], expectedAgg, "expected aggregation between Car and Wheel with multiplicity 4")

	// Check Garage -> Car multiplicities 1 and 0..*
	expectedStore1 := &ast.Relationship{
		From:             cd.Classes["Garage"],
		To:               cd.Classes["Car"],
		Type:             ast.Association,
		MultiplicityFrom: ast.Multiplicity{Raw: "1", Min: 1, Max: 1},
		MultiplicityTo:   ast.Multiplicity{Raw: "0..*", Min: 0, Max: -1},
		Comment:          "stores",
	}

	relsGarage, ok := cd.Relationships.GetRelationships("Garage")
	require.True(t, ok)
	require.Equal(t, relsGarage[0], expectedStore1, "expected Garage<->Car relationship with multiplicities 1 and 0..*")
}

func TestForwardReferenceCreatesPlaceholderClass(t *testing.T) {
	content := `
@startuml
class A
A --> B : references
@enduml
`
	cd, err := ParseClassDiagram(content)
	require.NoError(t, err)
	require.NotNil(t, cd)

	// B should be created as a placeholder even though it wasn't declared
	_, okA := cd.Classes["A"]
	_, okB := cd.Classes["B"]
	require.True(t, okA, "expected class A to exist")
	require.True(t, okB, "expected placeholder class B to be created")

	// relationship should reference both classes
	relsA, ok := cd.Relationships.GetRelationships("A")
	require.True(t, ok)
	expected := &ast.Relationship{
		From:             cd.Classes["A"],
		To:               cd.Classes["B"],
		Type:             ast.Association,
		Comment:          "references",
		MultiplicityFrom: ast.UnknownMultiplicity,
		MultiplicityTo:   ast.UnknownMultiplicity,
	}
	require.Contains(t, relsA, expected, "expected relationship between A and B")
}
