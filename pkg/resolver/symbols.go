package resolver

import (
	"fmt"
	"log"
	"slices"
	"strings"

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

type SymbolTable struct {
	Entities      []*EntitySymbol
	Relationships []*RelationshipSymbol
	Notes         []*ast.Note // Floating or unlinked notes

	// Lookup by identifier, alias, or qualified name
	lookup map[string]*EntitySymbol
}

func (tbl *SymbolTable) Lookup(name string) *EntitySymbol {
	return tbl.lookup[name]
}

func (tbl *SymbolTable) LookupByRef(ref ast.TargetRef) *EntitySymbol {
	return tbl.Lookup(FQN(ref.Entity, ref.PackagePath))
}

func (tbl *SymbolTable) FindOrCreateByRef(ref ast.TargetRef) *EntitySymbol {
	ent, ok := tbl.lookup[FQN(ref.Entity, ref.PackagePath)]
	if ok {
		return ent
	}
	newEnt := &EntitySymbol{
		FQN:         FQN(ref.Entity, ref.PackagePath),
		PackagePath: slices.Clone(ref.PackagePath),
	}
	tbl.Entities = append(tbl.Entities, newEnt)
	tbl.lookup[newEnt.FQN] = newEnt
	return newEnt
}

type EntitySymbol struct {
	FQN         string
	PackagePath []string
	AST         *ast.Entity
	Notes       []*ast.Note
}

func FQN(name string, pkgPath []string) string {
	if len(pkgPath) > 0 {
		return strings.Join(append(pkgPath, name), ".")
	}
	return name
}

func newEntitySymbol(ent *ast.Entity, pkgPath []string) *EntitySymbol {
	return &EntitySymbol{
		FQN:         FQN(ent.Identifier, pkgPath),
		AST:         ent,
		PackagePath: pkgPath,
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

func newRelationshipSymbol(tbl *SymbolTable, rel *ast.Relationship) (*RelationshipSymbol, error) {
	headL := classifyHead(rel.LArrow)
	headR := classifyHead(rel.RArrow)

	if headL >= HeadContainment && headR >= HeadContainment {
		if headL == HeadGeneralization && headR == HeadGeneralization {
			return nil, fmt.Errorf("cannot have bidirectional inheritance: %s and %s", rel.LHS.Entity, rel.RHS.Entity)
		}
		if headL == HeadContainment && headR == HeadContainment {
			return nil, fmt.Errorf("cannot have bidirectional composition/aggregation: %s and %s", rel.LHS.Entity, rel.RHS.Entity)
		}
		return nil, fmt.Errorf("conflicting relationship heads: cannot combine %s and %s on the same line", rel.TypeLHS, rel.TypeRHS)
	}

	symb := &RelationshipSymbol{
		AST: rel,
	}

	switch {
	case headL == HeadGeneralization:
		if headR == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on inheritance %s <|%c%c %s", rel.LHS.Entity, rel.Body, rel.Body, rel.RHS.Entity)
		}
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.RHS) // Subclass / Implementer
		symb.SourceMult = rel.MultRHS
		symb.Target = tbl.FindOrCreateByRef(rel.LHS) // Superclass / Interface
		symb.TargetMult = rel.MultLHS

	case headR == HeadGeneralization:
		if headL == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on inheritance %s %c%c|> %s", rel.LHS.Entity, rel.Body, rel.Body, rel.RHS.Entity)
		}
		symb.Type = rel.TypeRHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS) // Subclass / Implementer
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS) // Superclass / Interface
		symb.TargetMult = rel.MultRHS

	case headL == HeadContainment:
		if headR == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on containment %s %c-- %s", rel.LHS.Entity, rel.LArrow, rel.RHS.Entity)
		}
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS) // Owner
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS) // Contained
		symb.TargetMult = rel.MultRHS

	case headR == HeadContainment:
		if headL == HeadNavigation {
			log.Printf("warning: ignoring redundant association arrow on containment %s --%c %s", rel.LHS.Entity, rel.RArrow, rel.RHS.Entity)
		}
		symb.Type = rel.TypeRHS
		symb.Source = tbl.FindOrCreateByRef(rel.RHS) // Owner
		symb.SourceMult = rel.MultRHS
		symb.Target = tbl.FindOrCreateByRef(rel.LHS) // Contained
		symb.TargetMult = rel.MultLHS

	case headL == HeadNavigation && headR == HeadNavigation:
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS)
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS)
		symb.TargetMult = rel.MultRHS

	case headR == HeadNavigation: // A --> B
		symb.Type = rel.TypeRHS
		symb.Source = tbl.FindOrCreateByRef(rel.LHS)
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS)
		symb.TargetMult = rel.MultRHS

	case headL == HeadNavigation: // A <-- B
		symb.Type = rel.TypeLHS
		symb.Source = tbl.FindOrCreateByRef(rel.RHS)
		symb.SourceMult = rel.MultRHS
		symb.Target = tbl.FindOrCreateByRef(rel.LHS)
		symb.TargetMult = rel.MultLHS

	default: // Undirected A -- B
		symb.Source = tbl.FindOrCreateByRef(rel.LHS)
		symb.SourceMult = rel.MultLHS
		symb.Target = tbl.FindOrCreateByRef(rel.RHS)
		symb.TargetMult = rel.MultRHS
	}

	return symb, nil
}

func createImplicitEntity(ref ast.TargetRef, pkgPath []string) *EntitySymbol {
	ent := &ast.Entity{
		Identifier: ref.Entity,
	}
	implicitEntity := newEntitySymbol(ent, pkgPath)
	return implicitEntity
}

func attachNote(tbl *SymbolTable, noteRef, noteTargetRef ast.TargetRef, note *ast.Note, noteLookup map[string]*ast.Note) {
	noteTarget := tbl.FindOrCreateByRef(noteTargetRef)
	noteTarget.Notes = append(noteTarget.Notes, note)
	delete(noteLookup, noteRef.Entity)
}

func mergeEntity(target *ast.Entity, incoming *ast.Entity) error {
	target.Members = append(target.Members, incoming.Members...)

	if target.Kind != incoming.Kind {
		// Yeah ... plantuml just does this
		target.Kind = incoming.Kind
	}

	// Properties backfilling
	// Unfortunately, puml overwrites the properties of the earlier declarations
	if incoming.Stereotype != "" {
		target.Stereotype = incoming.Stereotype
	}
	if incoming.Generic != "" {
		target.Generic = incoming.Generic
	}
	if incoming.Color != "" {
		target.Color = incoming.Color
	}
	if incoming.Alias != "" {
		target.Alias = incoming.Alias
	}
	if len(incoming.Tags) > 0 {
		target.Tags = append(target.Tags, incoming.Tags...)
	}

	return nil
}

func resolveStatements(tbl *SymbolTable, stmts []ast.Statement, pkgPath []string, noteLookup map[string]*ast.Note) error {
	var prevRelationship *RelationshipSymbol
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case ast.Entity:
			if ent, ok := tbl.lookup[FQN(s.Identifier, pkgPath)]; ok {
				if ent.AST == nil {
					ent.AST = &s
					continue
				}
				if err := mergeEntity(ent.AST, &s); err != nil {
					return err
				}
				continue
			}
			ent := newEntitySymbol(&s, pkgPath)
			tbl.Entities = append(tbl.Entities, ent)
			tbl.lookup[ent.FQN] = ent
		case ast.Container:
			childPkgPath := append(pkgPath, s.Identifier)
			resolveStatements(tbl, s.Statements, childPkgPath, noteLookup)
		case ast.Note:
			// get rid of the link notes
			if s.Target.Entity == "link" {
				if prevRelationship == nil {
					return fmt.Errorf("no relationship found for link note %s", s.Text)
				}
				prevRelationship.Notes = append(prevRelationship.Notes, &s)
				continue
			}

			// get rid of the referenced notes
			if s.Identifier != "" {
				noteLookup[s.Identifier] = &s
				continue
			}

			// get rid of targeted notes
			if s.Target != nil {
				target := tbl.LookupByRef(*s.Target)
				if target == nil {
					return fmt.Errorf("no entity found for targeted note %v", s)
				}
				target.Notes = append(target.Notes, &s)
				continue
			}

			tbl.Notes = append(tbl.Notes, &s)
		case ast.Relationship:
			rel, err := newRelationshipSymbol(tbl, &s)
			if err != nil {
				return err
			}
			prevRelationship = rel

			if note, ok := noteLookup[rel.AST.LHS.Entity]; ok {
				attachNote(tbl, rel.AST.LHS, rel.AST.RHS, note, noteLookup)
				continue
			} else if note, ok := noteLookup[rel.AST.RHS.Entity]; ok {
				attachNote(tbl, rel.AST.RHS, rel.AST.LHS, note, noteLookup)
				continue
			}

			tbl.Relationships = append(tbl.Relationships, rel)
		}
	}
	return nil
}

// ResolveSymbols resolves all symbols in the AST and return flat symbol table.
func ResolveSymbols(diagram *ast.Diagram) (*SymbolTable, error) {
	tbl := &SymbolTable{
		lookup: map[string]*EntitySymbol{},
	}
	err := resolveStatements(tbl, diagram.Statements, nil, map[string]*ast.Note{})
	if err != nil {
		return nil, err
	}
	return tbl, nil
}
