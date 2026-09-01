package gogenerator

import (
	"log"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"
)

func toStructView(tbl *resolver.SymbolTable, ent *resolver.EntitySymbol, fileView *FileView) StructView {
	view := StructView{
		Name:       ent.AST.Identifier,
		NotesView:  toNotesView(ent.Notes),
		TriviaView: toTriviaView(ent.AST.Trivia),
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

	for _, member := range ent.AST.Members {
		switch member := member.(type) {
		case *dialect.GoField:
			view.Fields = append(view.Fields, toFieldView(ent.AST, member, fileView))
		case *dialect.GoMethod:
			view.Methods = append(view.Methods, toMethodView(ent.AST, member, fileView))
		case ast.ClassSeparator:
		}
	}

	return view
}

func toEnumView(ent *resolver.EntitySymbol) EnumView {
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

func toFieldView(owner *ast.Entity, field *dialect.GoField, fileView *FileView) FieldView {
	return FieldView{
		Name: ensureCorrectCase(owner.Identifier, field.Name, field.Visibility),
		Type: field.Type.String(),
	}
}

func toMethodView(owner *ast.Entity, method *dialect.GoMethod, fileView *FileView) MethodView {
	return MethodView{
		Name:      ensureCorrectCase(owner.Identifier, method.Name, method.Visibility),
		Signature: method.Signature(),
	}
}

func toInterfaceView(tbl *resolver.SymbolTable, ent *resolver.EntitySymbol, fileView *FileView) InterfaceView {
	view := InterfaceView{
		Name: ent.AST.Identifier,
	}
	for _, rel := range tbl.Relationships {
		if rel.Source == ent {
			switch rel.Type {
			case ast.RelationInheritance, ast.RelationRealization:
				view.Embeds = append(view.Embeds, rel.Target.AST.Identifier)
			}
		} else if rel.Target == ent {
			switch rel.Type {
			// need to clarify the difference between composition and aggregation
			case ast.RelationComposition:
			case ast.RelationAggregation:
			case ast.RelationInheritance, ast.RelationRealization:
				// do not output a warning, it will be handled in other struct
			default:
				log.Printf("warning: ignoring %s relationship between %s and %s", rel.Type, rel.Source.FQN, rel.Target.FQN)
			}
		}
	}

	for _, member := range ent.AST.Members {
		view.Methods = append(
			view.Methods,
			toMethodView(ent.AST, member.(*dialect.GoMethod), fileView),
		)
	}
	return view
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
