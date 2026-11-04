package resolver

import (
	"log"

	"yur4uwe/pac/pkg/parser/ast"
)

type HeadKind int

// ORDER MATTERS
const (
	HeadNone           HeadKind = iota
	HeadNavigation              // Arrow: '<' or '>' (Association / Dependency)
	HeadContainment             // Diamond: '*' or 'o' (Composition / Aggregation)
	HeadGeneralization          // Triangle: '|' (Inheritance / Realization)
)

func classifyHead(arrow rune) HeadKind {
	switch arrow {
	case '|':
		return HeadGeneralization
	case '*', 'o':
		return HeadContainment
	case '<', '>':
		return HeadNavigation
	default:
		return HeadNone
	}
}

type RelationshipSymbol struct {
	AST        *ast.Relationship
	Source     *EntitySymbol // Subclass / Implementer / Owner
	Target     *EntitySymbol // Superclass / Interface / Contained
	SourceMult ast.Cardinality
	TargetMult ast.Cardinality
	Type       ast.RelationType
	Notes      []*ast.Note // Notes on the relationship/link
}

func newRelationshipSymbol(tbl *SymbolTable, rel *ast.Relationship, pkgPath []string) *RelationshipSymbol {
	headL := classifyHead(rel.LArrow)
	headR := classifyHead(rel.RArrow)

	if headL >= HeadContainment && headR >= HeadContainment {
		if headL == HeadGeneralization && headR == HeadGeneralization {
			err := NewResolutionErrorf(rel, "cannot have bidirectional inheritance: %s and %s", rel.LHS.EntityName(), rel.RHS.EntityName())
			tbl.ast.RecordError(err, rel.Span(), rel)
		} else if headL == HeadContainment && headR == HeadContainment {
			err := NewResolutionErrorf(rel, "cannot have bidirectional composition/aggregation: %s and %s", rel.LHS.EntityName(), rel.RHS.EntityName())
			tbl.ast.RecordError(err, rel.Span(), rel)
		} else {
			err := NewResolutionErrorf(rel, "conflicting relationship heads: cannot combine %s and %s on the same line", rel.TypeLHS, rel.TypeRHS)
			tbl.ast.RecordError(err, rel.Span(), rel)
		}
		return nil
	}

	symb := &RelationshipSymbol{
		AST: rel,
	}

	switch {
	case headL == HeadGeneralization:
		if headR == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on inheritance %s <|%c%c %s", rel.LHS.EntityName(), rel.Body, rel.Body, rel.RHS.EntityName())
		}
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.RHS, pkgPath) // Subclass / Implementer
		symb.SourceMult = rel.MultRHS
		symb.Target = tbl.FindOrCreateByRef(rel.LHS, pkgPath) // Superclass / Interface
		symb.TargetMult = rel.MultLHS

	case headR == HeadGeneralization:
		if headL == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on inheritance %s %c%c|> %s", rel.LHS.EntityName(), rel.Body, rel.Body, rel.RHS.EntityName())
		}
		symb.Type = rel.TypeRHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS, pkgPath) // Subclass / Implementer
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS, pkgPath) // Superclass / Interface
		symb.TargetMult = rel.MultRHS

	case headL == HeadContainment:
		if headR == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on containment %s %c-- %s", rel.LHS.EntityName(), rel.LArrow, rel.RHS.EntityName())
		}
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS, pkgPath) // Owner
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS, pkgPath) // Contained
		symb.TargetMult = rel.MultRHS

	case headR == HeadContainment:
		if headL == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on containment %s --%c %s", rel.LHS.EntityName(), rel.RArrow, rel.RHS.EntityName())
		}
		symb.Type = rel.TypeRHS
		symb.Source = tbl.FindOrCreateByRef(rel.RHS, pkgPath) // Owner
		symb.SourceMult = rel.MultRHS
		symb.Target = tbl.FindOrCreateByRef(rel.LHS, pkgPath) // Contained
		symb.TargetMult = rel.MultLHS

	case headL == HeadNavigation && headR == HeadNavigation:
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS, pkgPath)
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS, pkgPath)
		symb.TargetMult = rel.MultRHS

	case headR == HeadNavigation: // A --> B
		symb.Type = rel.TypeRHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS, pkgPath)
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS, pkgPath)
		symb.TargetMult = rel.MultRHS

	case headL == HeadNavigation: // A <-- B
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.RHS, pkgPath)
		symb.SourceMult = rel.MultRHS
		symb.Target = tbl.FindOrCreateByRef(rel.LHS, pkgPath)
		symb.TargetMult = rel.MultLHS

	default: // Undirected A -- B
		symb.Source = tbl.FindOrCreateByRef(rel.LHS, pkgPath)
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS, pkgPath)
		symb.TargetMult = rel.MultRHS
	}

	return symb
}
