package resolver

import (
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

func resolveContainer(tbl *SymbolTable, s *ast.Container, pkgPath []string, noteLookup map[string]*ast.Note, namedNotes *[]*ast.Note) {
	containerPath := append(s.Identifier.PackagePathStrings(), s.Identifier.EntityName())
	childPkgPath := append(slices.Clone(pkgPath), containerPath...)
	resolveStatements(tbl, s.Statements, childPkgPath, noteLookup, namedNotes)
}

func resolveNote(tbl *SymbolTable, note *ast.Note, prevRelationship *RelationshipSymbol, pkgPath []string, noteLookup map[string]*ast.Note, namedNotes *[]*ast.Note) {
	// get rid of the link notes
	if note.Target != nil && note.Target.EntityName() == "link" {
		if prevRelationship == nil {
			err := NewResolutionErrorf(note, "no relationship found for link note %s", note.Text)
			tbl.ast.RecordError(err, note.Span(), note)
		} else {
			prevRelationship.Notes = append(prevRelationship.Notes, note)
		}
		return
	}

	// get rid of the referenced notes
	if !note.Identifier.Name.IsZero() {
		noteLookup[note.Identifier.EntityName()] = note
		*namedNotes = append(*namedNotes, note)
		return
	}

	// get rid of targeted notes
	if note.Target != nil {
		if note.Target.Name.IsZero() {
			err := NewResolutionErrorf(note, "empty target for note")
			tbl.ast.RecordError(err, note.Span(), note)
			tbl.Notes = append(tbl.Notes, note)
			return
		}
		target := tbl.FindOrCreateByRef(*note.Target, pkgPath)
		if target.AST == nil {
			target.AST = &ast.Entity{
				Identifier: *note.Target,
			}
		}
		if note.Target.MemberName() != "" {
			if target.MemberNotes == nil {
				target.MemberNotes = make(map[string][]*ast.Note)
			}
			target.MemberNotes[note.Target.MemberName()] = append(target.MemberNotes[note.Target.MemberName()], note)
			return
		}
		target.Notes = append(target.Notes, note)
		return
	}

	tbl.Notes = append(tbl.Notes, note)
}

func resolveRelationship(tbl *SymbolTable, s *ast.Relationship, pkgPath []string, noteLookup map[string]*ast.Note) *RelationshipSymbol {
	rel := newRelationshipSymbol(tbl, s, pkgPath)
	if rel == nil {
		return nil
	}

	if note, ok := noteLookup[rel.AST.LHS.EntityName()]; ok {
		attachNote(tbl, rel.AST.LHS, rel.AST.RHS, note, noteLookup, pkgPath)
		return rel
	} else if note, ok := noteLookup[rel.AST.RHS.EntityName()]; ok {
		attachNote(tbl, rel.AST.RHS, rel.AST.LHS, note, noteLookup, pkgPath)
		return rel
	}

	tbl.Relationships = append(tbl.Relationships, rel)
	return rel
}

func resolveStatements(tbl *SymbolTable, stmts []ast.Statement, pkgPath []string, noteLookup map[string]*ast.Note, namedNotes *[]*ast.Note) {
	var prevRelationship *RelationshipSymbol
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.Entity:
			resolveEntity(tbl, s, pkgPath)
		case *ast.InlineMember:
			resolveInlineMember(tbl, s, pkgPath)
		case *ast.Container:
			resolveContainer(tbl, s, pkgPath, noteLookup, namedNotes)
		case *ast.Note:
			resolveNote(tbl, s, prevRelationship, pkgPath, noteLookup, namedNotes)
		case *ast.Relationship:
			rel := resolveRelationship(tbl, s, pkgPath, noteLookup)
			if rel != nil {
				prevRelationship = rel
			}
		}
	}
}

// ResolveSymbols resolves all symbols in the AST and return flat symbol table.
func ResolveSymbols(diagram *ast.Diagram) *SymbolTable {
	tbl := &SymbolTable{
		ast:    diagram,
		lookup: map[string]*EntitySymbol{},
	}
	noteLookup := map[string]*ast.Note{}
	var namedNotes []*ast.Note
	resolveStatements(tbl, diagram.Statements, nil, noteLookup, &namedNotes)
	for _, n := range namedNotes {
		if _, ok := noteLookup[n.Identifier.EntityName()]; ok {
			tbl.Notes = append(tbl.Notes, n)
		}
	}
	return tbl
}
