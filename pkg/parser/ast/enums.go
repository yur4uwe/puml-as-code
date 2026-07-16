package ast

import (
	"fmt"
	"strconv"
	"strings"
)

type RelationType int

const (
	Association RelationType = iota
	Aggregation
	Composition
	Inheritance
	Dependency
	Realization
	UnknownRelation
)

func (rt RelationType) String() string {
	switch rt {
	case Association:
		return "Association"
	case Aggregation:
		return "Aggregation"
	case Composition:
		return "Composition"
	case Inheritance:
		return "Inheritance"
	case Dependency:
		return "Dependency"
	case Realization:
		return "Realization"
	default:
		return "Unknown Relation"
	}
}

func ToRelationType(relStr string) (RelationType, int) {
	relStr = strings.TrimSpace(relStr)
	switch relStr {
	case "--":
		return Association, 0
	case "<--":
		return Association, -1
	case "-->":
		return Association, 1
	case "o--":
		return Aggregation, -1
	case "--o":
		return Aggregation, 1
	case "*--":
		return Composition, -1
	case "--*":
		return Composition, 1
	case "<|--":
		return Inheritance, -1
	case "--|>":
		return Inheritance, 1
	case "<..":
		return Realization, -1
	case "..>":
		return Realization, 1
	case "<..>":
		return Dependency, 0
	default:
		return UnknownRelation, 0
	}
}

type VisibilityKind int

const (
	UnknownVisibility VisibilityKind = iota
	Public
	Private
	Protected
	Package
)

func (v VisibilityKind) String() string {
	switch v {
	case Public:
		return "public"
	case Private:
		return "private"
	case Protected:
		return "protected"
	case Package:
		return "package"
	default:
		return "unknown"
	}
}

type ValueType byte

const (
	Void ValueType = iota
	Int
	String
	Float
	Bool
	Custom
	UnknownType
)

func (vt ValueType) String() string {
	switch vt {
	case Void:
		return "void"
	case Int:
		return "int"
	case String:
		return "string"
	case Float:
		return "float"
	case Bool:
		return "bool"
	case Custom:
		return "custom"
	default:
		return "unknown"
	}
}

func ToValueType(typeStr string) ValueType {
	s := strings.TrimSpace(strings.ToLower(typeStr))
	switch s {
	case "void":
		return Void
	case "int":
		return Int
	case "string":
		return String
	case "float":
		return Float
	case "bool":
		return Bool
	default:
		if s == "" {
			return UnknownType
		}
		return Custom
	}
}

type Cardinality struct {
	Raw string
	Min int
	Max int // -1 represents '*' (unbounded)
}

// UnknownCardinality is an empty/unknown multiplicity.
var UnknownCardinality = Cardinality{Raw: "", Min: -2, Max: -2}

func ParseCardinality(s string) (Cardinality, error) {
	s = strings.TrimSpace(strings.Trim(s, `"`))
	if s == "" {
		return UnknownCardinality, nil
	}
	if s == "*" {
		return Cardinality{Raw: s, Min: 0, Max: -1}, nil
	}
	if strings.Contains(s, "..") {
		parts := strings.SplitN(s, "..", 2)
		if len(parts) != 2 {
			return UnknownCardinality, fmt.Errorf("invalid cardinality: %q", s)
		}
		min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return UnknownCardinality, fmt.Errorf("invalid min in cardinality %q: %w", s, err)
		}
		maxStr := strings.TrimSpace(parts[1])
		if maxStr == "*" {
			return Cardinality{Raw: s, Min: min, Max: -1}, nil
		}
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return UnknownCardinality, fmt.Errorf("invalid max in cardinality %q: %w", s, err)
		}
		return Cardinality{Raw: s, Min: min, Max: max}, nil
	}
	// single number
	n, err := strconv.Atoi(s)
	if err != nil {
		return UnknownCardinality, fmt.Errorf("invalid cardinality: %q", s)
	}
	return Cardinality{Raw: s, Min: n, Max: n}, nil
}

func (m Cardinality) String() string {
	if m == UnknownCardinality {
		return ""
	}
	if m.Max == -1 {
		if m.Min == 0 {
			return "*"
		}
		return fmt.Sprintf("%d..*", m.Min)
	}
	if m.Min == m.Max {
		return strconv.Itoa(m.Min)
	}
	return fmt.Sprintf("%d..%d", m.Min, m.Max)
}
