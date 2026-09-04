package gogenerator

import (
	"fmt"
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
			view.Fields = append(view.Fields, toFieldView(ent, member, fileView))
		case *dialect.GoMethod:
			view.Methods = append(view.Methods, toMethodView(ent, member, fileView))
		case ast.ClassSeparator:
		}
	}

	return view, nil
}

func toInterfaceView(tbl *resolver.SymbolTable, ent *resolver.EntitySymbol, fileView *FileView) (InterfaceView, error) {
	view := InterfaceView{
		Name:      stdlib.SimpleName(ent.FQN),
		NotesView: toNotesView(ent.Notes),
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
			toMethodView(ent, member.(*dialect.GoMethod), fileView),
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
		Name:       ent.AST.Identifier,
		Values:     cases,
		NotesView:  toNotesView(ent.Notes),
		TriviaView: toTriviaView(ent.AST.Trivia),
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

func visibilityComment(vis ast.VisibilityKind) string {
	switch vis {
	case ast.VisibilityPrivate:
		return "private"
	case ast.VisibilityProtected:
		return "protected"
	default:
		return ""
	}
}

func attachVisibilityComment(trivia *TriviaView, vis ast.VisibilityKind) {
	if comment := visibilityComment(vis); comment != "" {
		if len(trivia.TrailingTrivia) == 0 {
			trivia.TrailingTrivia = []string{comment}
		} else {
			trivia.TrailingTrivia[0] = comment + "; " + trivia.TrailingTrivia[0]
		}
	}
}

func toFieldView(owner *resolver.EntitySymbol, field *dialect.GoField, fileView *FileView) FieldView {
	collectImports(field.Type, fileView)
	trivia := toTriviaView(field.Trivia)
	attachVisibilityComment(&trivia, field.Visibility)
	return FieldView{
		Name:       ensureCorrectCase(owner.AST.Identifier, field.Name, field.Visibility),
		Type:       field.Type.String(),
		TriviaView: trivia,
		NotesView:  toNotesView(owner.MemberNotes[field.Name]),
	}
}

func toMethodView(owner *resolver.EntitySymbol, method *dialect.GoMethod, fileView *FileView) MethodView {
	for _, param := range method.Parameters {
		collectImports(param.Type, fileView)
	}
	for _, ret := range method.ReturnType {
		collectImports(ret.Type, fileView)
	}
	trivia := toTriviaView(method.Trivia)
	attachVisibilityComment(&trivia, method.Visibility)
	return MethodView{
		Name:       ensureCorrectCase(owner.AST.Identifier, method.Name, method.Visibility),
		Signature:  method.Signature(),
		TriviaView: trivia,
		NotesView:  toNotesView(owner.MemberNotes[method.Name]),
	}
}

func toNotesView(note []*ast.Note) NotesView {
	var notes []string
	for _, n := range note {
		raw := strings.TrimSpace(n.Text)
		if raw == "" {
			continue
		}
		var headerSet bool
		for line := range strings.SplitSeq(raw, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if !headerSet {
				notes = append(notes, "NOTE: "+trimmed)
				headerSet = true
			} else {
				notes = append(notes, trimmed)
			}
		}
	}
	return NotesView{
		Notes: notes,
	}
}

func toTriviaView(t ast.Trivia) TriviaView {
	var leadingTrivia []string
	for _, tok := range t.GetLeadingTrivia() {
		lines := strings.SplitSeq(tok.Literal, "\n")
		for line := range lines {
			leadingTrivia = append(leadingTrivia, strings.TrimSpace(line))
		}
	}

	var trailingTrivia []string
	var lineTokens []string
	var currentLine uint

	for _, tok := range t.GetTrailingTrivia() {
		lit := strings.TrimSpace(tok.Literal)
		lines := strings.Split(lit, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSpace(line)
			if len(lines[i]) == 0 {
				continue
			}
		}
		if len(lineTokens) == 0 {
			lineTokens = append(lineTokens, lines...)
			currentLine = tok.Pos.Line
		} else if tok.Pos.Line == currentLine {
			lineTokens = append(lineTokens, lines...)
		} else {
			trailingTrivia = append(trailingTrivia, strings.Join(lineTokens, "; "))
			lineTokens = lines
			currentLine = tok.Pos.Line
		}
	}
	if len(lineTokens) > 0 {
		trailingTrivia = append(trailingTrivia, strings.Join(lineTokens, "; "))
	}

	if len(trailingTrivia) > 2 {
		// defensive check to spot issues and bugs with trailing trivia aggregation
		panic(fmt.Sprintf("too many trailing trivia: %d", len(trailingTrivia)))
	}

	return TriviaView{
		LeadingTrivia:  leadingTrivia,
		TrailingTrivia: trailingTrivia,
	}
}
