package resolver

import (
	"slices"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

type SymbolTable struct {
	ast           *ast.Diagram
	Entities      []*EntitySymbol
	Relationships []*RelationshipSymbol
	Notes         []*ast.Note // Floating or unlinked notes

	// Lookup by identifier, alias, or qualified name
	lookup           map[string]*EntitySymbol
	packageSeparator string
	declarations     *SymbolTable // Temporary index for references to later declarations.
}

func (tbl *SymbolTable) Lookup(name string) *EntitySymbol {
	return tbl.lookup[name]
}

func (tbl *SymbolTable) LookupByRef(ref ast.TargetRef, pkgPath []string) *EntitySymbol {
	pkgPathStrings := ref.PackagePathStrings()
	entityName := ref.EntityName()

	if len(pkgPathStrings) > 0 {
		if ent, ok := tbl.lookup[FQN(entityName, pkgPathStrings)]; ok {
			return ent
		}
		if len(pkgPath) > 0 {
			relPkg := append(slices.Clone(pkgPath), pkgPathStrings...)
			if ent, ok := tbl.lookup[FQN(entityName, relPkg)]; ok {
				return ent
			}
		}
		return nil
	}

	// 1. Current package scope
	if len(pkgPath) > 0 {
		if ent, ok := tbl.lookup[FQN(entityName, pkgPath)]; ok {
			return ent
		}
	}

	// 2. Parent package scopes up to root
	for i := len(pkgPath) - 1; i >= 1; i-- {
		if ent, ok := tbl.lookup[FQN(entityName, pkgPath[:i])]; ok {
			return ent
		}
	}

	// 3. Root scope
	if ent, ok := tbl.lookup[entityName]; ok {
		return ent
	}

	// 4. Global search across all known entities for unique simple name match
	var candidate *EntitySymbol
	for _, ent := range tbl.Entities {
		simpleName := ent.FQN
		if idx := strings.LastIndex(ent.FQN, "."); idx != -1 {
			simpleName = ent.FQN[idx+1:]
		}
		if simpleName == entityName || (ent.AST != nil && ent.AST.Identifier.EntityName() == entityName) {
			if candidate == nil {
				candidate = ent
			} else if candidate.FQN != ent.FQN {
				// Ambiguous match across multiple packages
				return nil
			}
		}
	}

	return candidate
}

func (tbl *SymbolTable) FindOrCreateByRef(ref ast.TargetRef, pkgPath []string) *EntitySymbol {
	if ent := tbl.LookupByRef(ref, pkgPath); ent != nil {
		return ent
	}

	targetPkg := ref.PackagePathStrings()
	if len(targetPkg) == 0 && len(pkgPath) > 0 {
		targetPkg = slices.Clone(pkgPath)
	}

	newEnt := &EntitySymbol{
		FQN:         FQN(ref.EntityName(), targetPkg),
		PackagePath: slices.Clone(targetPkg),
	}
	tbl.Entities = append(tbl.Entities, newEnt)
	tbl.lookup[newEnt.FQN] = newEnt
	return newEnt
}

func FQN(name string, pkgPath []string) string {
	if len(pkgPath) > 0 {
		return strings.Join(append(pkgPath, name), ".")
	}
	return name
}
