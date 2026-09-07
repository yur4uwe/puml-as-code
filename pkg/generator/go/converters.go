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

	var pendingSeparators []string
	for _, member := range ent.AST.Members {
		switch member := member.(type) {
		case *dialect.GoField:
			fv := toFieldView(ent, member, fileView)
			if len(pendingSeparators) > 0 {
				fv.LeadingTrivia = append(pendingSeparators, fv.LeadingTrivia...)
				pendingSeparators = nil
			}
			view.Fields = append(view.Fields, fv)
		case *dialect.GoMethod:
			mv := toMethodView(ent, member, fileView)
			if len(pendingSeparators) > 0 {
				mv.LeadingTrivia = append(pendingSeparators, mv.LeadingTrivia...)
				pendingSeparators = nil
			}
			view.Methods = append(view.Methods, mv)
		case ast.ClassSeparator:
			pendingSeparators = append(pendingSeparators, formatSeparator(member)...)
		}
	}

	if ent.AST.Kind == ast.EntityException {
		if !slices.Contains(view.Implements, "error") {
			view.Implements = append(view.Implements, "error")
		}
		hasErrorMethod := slices.ContainsFunc(view.Methods, func(m MethodView) bool {
			return m.Name == "Error"
		})
		if !hasErrorMethod {
			view.Methods = append(view.Methods, MethodView{
				Name:      "Error",
				Signature: "() string",
			})
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

	var pendingSeparators []string
	for _, member := range ent.AST.Members {
		switch m := member.(type) {
		case *dialect.GoMethod:
			mv := toMethodView(ent, m, fileView)
			if len(pendingSeparators) > 0 {
				mv.LeadingTrivia = append(pendingSeparators, mv.LeadingTrivia...)
				pendingSeparators = nil
			}
			view.Methods = append(
				view.Methods,
				mv,
			)
		case ast.ClassSeparator:
			pendingSeparators = append(pendingSeparators, formatSeparator(m)...)
		}
	}
	return view, nil
}

func toEnumView(ent *resolver.EntitySymbol) EnumView {
	if ent == nil {
		panic("enums should always be explicitly defined")
	}
	var cases []string
	for _, member := range ent.AST.Members {
		switch m := member.(type) {
		case *dialect.GoField:
			cases = append(cases, m.Name)
		case ast.ClassSeparator:
		}
	}
	view := EnumView{
		Name:       ent.AST.Identifier,
		Values:     cases,
		NotesView:  toNotesView(ent.Notes),
		TriviaView: toTriviaView(ent.AST.Trivia),
	}
	return view
}

func formatSeparator(sep ast.ClassSeparator) []string {
	sepChar := sep.Type
	if sepChar == 0 {
		sepChar = '-'
	}
	fill := strings.Repeat(string(sepChar), 3)
	label := strings.TrimSpace(sep.Label)
	var text string
	if label != "" {
		text = fmt.Sprintf("%s %s %s", fill, label, fill)
	} else {
		text = fill
	}
	return []string{"", text, ""}
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
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				leadingTrivia = append(leadingTrivia, trimmed)
			}
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
