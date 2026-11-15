package resolver

import (
	"slices"

	"yur4uwe/pac/pkg/parser/ast"
)

// Index declared entities separately so forward references can be
// disambiguated without changing the resolver's symbol creation order.
func indexDeclarations(tbl *SymbolTable, stmts []ast.Statement, path []string) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.Entity:
			if s.Identifier.EntityName() != "" {
				resolveEntity(tbl, s, path)
			}
		case *ast.Container:
			inner := path
			if s.Identifier.EntityName() != "" {
				inner = append(slices.Clone(path), s.Identifier.PackagePathStrings()...)
				inner = append(inner, s.Identifier.EntityName())
			}
			indexDeclarations(tbl, s.Statements, inner)
		}
	}
}

func (tbl *SymbolTable) resolveMemberRef(ref ast.TargetRef, path []string) ast.TargetRef {
	if tbl.packageSeparator != "::" || ref.Member != nil || len(ref.PackagePath) == 0 {
		return ref
	}
	// An actual entity with this qualified name takes precedence over a member.
	if tbl.LookupByRef(ref, path) != nil || tbl.declarations.LookupByRef(ref, path) != nil {
		return ref
	}

	n := len(ref.PackagePath) - 1
	owner := ast.TargetRef{Name: ref.PackagePath[n], PackagePath: ref.PackagePath[:n]}
	for _, scope := range []*SymbolTable{tbl, tbl.declarations} {
		if ent := scope.LookupByRef(owner, path); ent != nil && ent.AST != nil && hasMember(ent.AST.Members, ref.EntityName()) {
			owner.Member = &ast.MemberRef{Name: ref.Name}
			// A declaration found via the forward index may live outside
			// the current package. Preserve its actual scope when creating it.
			if scope == tbl.declarations && tbl.LookupByRef(owner, path) == nil {
				owner.Name = ent.AST.Identifier.Name
				owner.PackagePath = make([]ast.Ident, len(ent.PackagePath))
				for i, name := range ent.PackagePath {
					owner.PackagePath[i].Literal = name
				}
			}
			return owner
		}
	}
	return ref
}

func hasMember(members []ast.Member, name string) bool {
	return slices.ContainsFunc(members, func(member ast.Member) bool {
		switch m := member.(type) {
		case *ast.FieldDeclaration:
			return m.Field != nil && m.FieldName() == name
		case *ast.MethodDeclaration:
			return m.Method != nil && m.MethodName() == name
		default:
			return false
		}
	})
}
