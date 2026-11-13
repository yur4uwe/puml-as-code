package resolver

import (
	"slices"

	"yur4uwe/pac/pkg/parser/ast"
)

type EntitySymbol struct {
	FQN         string
	PackagePath []string
	AST         *ast.Entity
	Notes       []*ast.Note
	MemberNotes map[string][]*ast.Note
}

func newEntitySymbol(ent *ast.Entity, pkgPath []string) *EntitySymbol {
	combinedPkgPath := append(slices.Clone(pkgPath), ent.Identifier.PackagePathStrings()...)
	return &EntitySymbol{
		FQN:         FQN(ent.Identifier.EntityName(), combinedPkgPath),
		AST:         ent,
		PackagePath: combinedPkgPath,
	}
}

func createImplicitEntity(ref ast.TargetRef, pkgPath []string) *EntitySymbol {
	ent := &ast.Entity{
		Identifier: ref,
	}
	implicitEntity := newEntitySymbol(ent, pkgPath)
	return implicitEntity
}

func mergeEntities(target *ast.Entity, incoming *ast.Entity) *ast.Entity {
	merge := &ast.Entity{
		Kind:       target.Kind,
		Visibility: target.Visibility,
		Identifier: target.Identifier,
		Alias:      target.Alias,
		Stereotype: target.Stereotype,
		Generic:    target.Generic,
		Color:      target.Color,
		Tags:       target.Tags,
	}
	merge.Members = append(target.Members, incoming.Members...)

	if target.Kind != incoming.Kind {
		// Yeah ... plantuml just does this
		merge.Kind = incoming.Kind
	}

	// Properties backfilling
	// Unfortunately, puml overwrites the properties of the earlier declarations
	if incoming.Visibility != ast.VisibilityUnknown {
		merge.Visibility = incoming.Visibility
	}
	if incoming.Stereotype != nil {
		merge.Stereotype = incoming.Stereotype
	}
	if incoming.Generic != "" {
		merge.Generic = incoming.Generic
	}
	if incoming.Color != "" {
		merge.Color = incoming.Color
	}
	if !incoming.Alias.IsZero() {
		merge.Alias = incoming.Alias
	}
	if len(incoming.Tags) > 0 {
		merge.Tags = append(target.Tags, incoming.Tags...)
	}

	return merge
}
