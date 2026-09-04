package gogenerator

import (
	"slices"
	"strings"

	"yur4uwe/pac/pkg/generator/go/stdlib"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"
)

func toStructView(tbl *resolver.SymbolTable, ent *resolver.EntitySymbol, fileView *FileView) (StructView, error) {
	view := StructView{
		Name:      stdlib.SimpleName(ent.FQN),
		NotesView: toNotesView(ent.Notes),
	}
	if ent.AST != nil {
		view.TriviaView = toTriviaView(ent.AST.Trivia)
	}

	// process generics
	if ent.AST != nil && len(ent.AST.Generic) != 0 {
		generics, err := parseGeneric(ent.AST.Generic)
		if err != nil {
			return view, err
		}
		view.Generics = generics
	}

	for _, rel := range tbl.Relationships {
		if rel.Source == ent {
			fillSourceStructByRel(&view, rel, fileView)
		}
		// // For now doesn't handle any cases
		// } else if rel.Target == ent {
		// 	renderTargetStructRel(&view, rel)
		// }
	}

	if ent.AST == nil {
		return view, nil
	}

	for _, member := range ent.AST.Members {
		switch member := member.(type) {
		case *dialect.GoField:
			view.Fields = append(view.Fields, toFieldView(ent.AST, member, fileView))
		case *dialect.GoMethod:
			view.Methods = append(view.Methods, toMethodView(ent.AST, member, fileView))
		case ast.ClassSeparator:
		}
	}

	return view, nil
}

func toInterfaceView(tbl *resolver.SymbolTable, ent *resolver.EntitySymbol, fileView *FileView) (InterfaceView, error) {
	view := InterfaceView{
		Name: stdlib.SimpleName(ent.FQN),
	}

	if ent.AST != nil {
		if len(ent.AST.Generic) != 0 {
			generics, err := parseGeneric(ent.AST.Generic)
			if err != nil {
				return view, err
			}
			view.Generics = generics
		}
		view.TriviaView = toTriviaView(ent.AST.Trivia)
	}

	for _, rel := range tbl.Relationships {
		if rel.Source != ent {
			continue
		}

		switch rel.Type {
		case ast.RelationInheritance, ast.RelationRealization:
			if impPath, ok := stdlib.LookupImportPath(rel.Target.PackagePath); ok {
				fileView.AddImport(impPath)
			} else if len(rel.Target.PackagePath) > 0 && !slices.Equal(rel.Source.PackagePath, rel.Target.PackagePath) {
				fileView.AddImport(strings.Join(rel.Target.PackagePath, "/"))
			}

			embedding := targetTypeName(rel.Source.PackagePath, rel.Target)
			view.Embeds = append(view.Embeds, embedding)
		}
	}

	for _, member := range ent.AST.Members {
		view.Methods = append(
			view.Methods,
			toMethodView(ent.AST, member.(*dialect.GoMethod), fileView),
		)
	}
	return view, nil
}

func toEnumView(ent *resolver.EntitySymbol) EnumView {
	if ent == nil {
		panic("enums should always be explicitly defined")
	}
	cases := make([]string, len(ent.AST.Members))
	for i, member := range ent.AST.Members {
		cases[i] = member.(*dialect.GoField).Name
	}
	view := EnumView{
		Name:   ent.AST.Identifier,
		Values: cases,
	}
	return view
}

func collectImports(typeRef *dialect.GoTypeRef, fileView *FileView) {
	if typeRef == nil || fileView == nil {
		return
	}

	// Traverse modifiers (*, [], [N]) to find the root named node
	curr := typeRef
	for curr != nil && curr.Typ != dialect.KindNamed {
		curr = curr.Base
	}

	// If curr has a Base, it is a qualified package reference (e.g. "time" in time.
	// Time)
	if curr == nil || curr.Base == nil {
		return
	}

	pkgName := curr.Name
	if imp, ok := stdlib.LookupImportPath([]string{pkgName}); ok {
		fileView.AddImport(imp)
	} else {
		fileView.AddImport(pkgName)
	}
}

func toFieldView(owner *ast.Entity, field *dialect.GoField, fileView *FileView) FieldView {
	collectImports(field.Type, fileView)
	return FieldView{
		Name:       ensureCorrectCase(owner.Identifier, field.Name, field.Visibility),
		Type:       field.Type.String(),
		TriviaView: toTriviaView(field.Trivia),
		// NotesView:  toNotesView(field.Notes),
	}
}

func toMethodView(owner *ast.Entity, method *dialect.GoMethod, fileView *FileView) MethodView {
	for _, param := range method.Parameters {
		collectImports(param.Type, fileView)
	}
	for _, ret := range method.ReturnType {
		collectImports(ret.Type, fileView)
	}
	return MethodView{
		Name:       ensureCorrectCase(owner.Identifier, method.Name, method.Visibility),
		Signature:  method.Signature(),
		TriviaView: toTriviaView(method.Trivia),
		// NotesView:  toNotesView(method.Notes),
	}
}

func toNotesView(note []*ast.Note) NotesView {
	var notes []string
	for _, n := range note {
		notes = append(notes, n.Text)
	}
	return NotesView{
		Notes: notes,
	}
}

func toTriviaView(t ast.Trivia) TriviaView {
	var leadingTrivia []string
	for _, tok := range t.GetLeadingTrivia() {
		leadingTrivia = append(leadingTrivia, tok.Literal)
	}
	var trailingTrivia []string
	for _, tok := range t.GetTrailingTrivia() {
		trailingTrivia = append(trailingTrivia, tok.Literal)
	}
	return TriviaView{
		LeadingTrivia:  leadingTrivia,
		TrailingTrivia: trailingTrivia,
	}
}
