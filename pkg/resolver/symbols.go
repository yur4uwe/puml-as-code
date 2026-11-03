package resolver

import (
	"fmt"
	"slices"

	"yur4uwe/pac/pkg/parser/ast"
)

func attachNote(tbl *SymbolTable, noteRef, noteTargetRef ast.TargetRef, note *ast.Note, noteLookup map[string]*ast.Note, pkgPath []string) {
	noteTarget := tbl.FindOrCreateByRef(noteTargetRef, pkgPath)
	if noteTargetRef.MemberName() != "" {
		if noteTarget.MemberNotes == nil {
			noteTarget.MemberNotes = make(map[string][]*ast.Note)
		}
		noteTarget.MemberNotes[noteTargetRef.MemberName()] = append(noteTarget.MemberNotes[noteTargetRef.MemberName()], note)
	} else {
		noteTarget.Notes = append(noteTarget.Notes, note)
	}
	delete(noteLookup, noteRef.EntityName())
}

func resolveEntity(tbl *SymbolTable, s *ast.Entity, pkgPath []string) {
	combinedPkgPath := append(slices.Clone(pkgPath), s.Identifier.PackagePathStrings()...)
	fqn := FQN(s.Identifier.EntityName(), combinedPkgPath)
	if ent, ok := tbl.lookup[fqn]; ok {
		if ent.AST == nil {
			ent.AST = s
			return
		}
		ent.AST = mergeEntities(ent.AST, s)
		return
	}
	ent := newEntitySymbol(s, pkgPath)
	tbl.Entities = append(tbl.Entities, ent)
	tbl.lookup[ent.FQN] = ent
	if !s.Alias.IsZero() {
		tbl.lookup[s.Alias.Literal] = ent
	}
}

func resolveInlineMember(tbl *SymbolTable, s *ast.InlineMember, pkgPath []string) {
	ent := tbl.FindOrCreateByRef(s.Target, pkgPath)
	if ent.AST == nil {
		ent.AST = &ast.Entity{
			Identifier: s.Target,
		}
	}
	ent.AST.Members = append(ent.AST.Members, s.Member)
}

func resolveContainer(tbl *SymbolTable, s *ast.Container, pkgPath []string, noteLookup map[string]*ast.Note, namedNotes *[]*ast.Note) error {
	containerPath := append(s.Identifier.PackagePathStrings(), s.Identifier.EntityName())
	childPkgPath := append(slices.Clone(pkgPath), containerPath...)
	return resolveStatements(tbl, s.Statements, childPkgPath, noteLookup, namedNotes)
}

func resolveNote(tbl *SymbolTable, s *ast.Note, prevRelationship *RelationshipSymbol, pkgPath []string, noteLookup map[string]*ast.Note, namedNotes *[]*ast.Note) error {
	// get rid of the link notes
	if s.Target != nil && s.Target.EntityName() == "link" {
		if prevRelationship == nil {
			return fmt.Errorf("no relationship found for link note %s", s.Text)
		}
		prevRelationship.Notes = append(prevRelationship.Notes, s)
		return nil
	}

	// get rid of the referenced notes
	if !s.Identifier.Name.IsZero() {
		noteLookup[s.Identifier.EntityName()] = s
		*namedNotes = append(*namedNotes, s)
		return nil
	}

	// get rid of targeted notes
	if s.Target != nil {
		target := tbl.LookupByRef(*s.Target, pkgPath)
		if target == nil {
			return fmt.Errorf("no entity found for targeted note %v", s)
		}
		if s.Target.MemberName() != "" {
			if target.MemberNotes == nil {
				target.MemberNotes = make(map[string][]*ast.Note)
			}
			target.MemberNotes[s.Target.MemberName()] = append(target.MemberNotes[s.Target.MemberName()], s)
			return nil
		}
		target.Notes = append(target.Notes, s)
		return nil
	}

	tbl.Notes = append(tbl.Notes, s)
	return nil
}

func resolveRelationship(tbl *SymbolTable, s *ast.Relationship, pkgPath []string, noteLookup map[string]*ast.Note) (*RelationshipSymbol, error) {
	rel, err := newRelationshipSymbol(tbl, s, pkgPath)
	if err != nil {
		return nil, err
	}

	if note, ok := noteLookup[rel.AST.LHS.EntityName()]; ok {
		attachNote(tbl, rel.AST.LHS, rel.AST.RHS, note, noteLookup, pkgPath)
		return rel, nil
	} else if note, ok := noteLookup[rel.AST.RHS.EntityName()]; ok {
		attachNote(tbl, rel.AST.RHS, rel.AST.LHS, note, noteLookup, pkgPath)
		return rel, nil
	}

	tbl.Relationships = append(tbl.Relationships, rel)
	return rel, nil
}

func resolveStatements(tbl *SymbolTable, stmts []ast.Statement, pkgPath []string, noteLookup map[string]*ast.Note, namedNotes *[]*ast.Note) error {
	var prevRelationship *RelationshipSymbol
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.Entity:
			resolveEntity(tbl, s, pkgPath)
		case *ast.InlineMember:
			resolveInlineMember(tbl, s, pkgPath)
		case *ast.Container:
			if err := resolveContainer(tbl, s, pkgPath, noteLookup, namedNotes); err != nil {
				return err
			}
		case *ast.Note:
			if err := resolveNote(tbl, s, prevRelationship, pkgPath, noteLookup, namedNotes); err != nil {
				return err
			}
		case *ast.Relationship:
			rel, err := resolveRelationship(tbl, s, pkgPath, noteLookup)
			if err != nil {
				return err
			}
			prevRelationship = rel
		}
	}
	return nil
}

// ResolveSymbols resolves all symbols in the AST and return flat symbol table.
func ResolveSymbols(diagram *ast.Diagram) (*SymbolTable, error) {
	tbl := &SymbolTable{
		lookup: map[string]*EntitySymbol{},
	}
	noteLookup := map[string]*ast.Note{}
	var namedNotes []*ast.Note
	err := resolveStatements(tbl, diagram.Statements, nil, noteLookup, &namedNotes)
	if err != nil {
		return nil, err
	}
	for _, n := range namedNotes {
		if _, ok := noteLookup[n.Identifier.EntityName()]; ok {
			tbl.Notes = append(tbl.Notes, n)
		}
	}
	return tbl, nil
}
