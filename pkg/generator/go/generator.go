package gogenerator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"log"
	"strconv"
	"strings"
	"text/template"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"
)

type GoCodeGenerator struct{}

type GeneratedFile struct {
	Path    string
	Content []byte
}

//go:embed templates/*.go.tmpl
var templateFS embed.FS

func (GoCodeGenerator) GenerateFromClassDiagram(tbl *resolver.SymbolTable) ([]*GeneratedFile, error) {
	var files []*GeneratedFile
	// map package path to file writer
	viewMap := map[string]*FileView{}

	for _, ent := range tbl.Entities {
		filePath := "types.go"
		if len(ent.PackagePath) > 0 {
			filePath = strings.Join(ent.PackagePath, "/") + "/" + filePath
		}
		view, ok := viewMap[filePath]
		if !ok {
			pkgName := "root"
			if len(ent.PackagePath) > 0 {
				pkgName = ent.PackagePath[len(ent.PackagePath)-1]
			}
			view = &FileView{
				PackageName: pkgName,
			}
			viewMap[filePath] = view
		}

		switch {
		case isStruct(ent.AST):
			view.Structs = append(view.Structs, toStructView(tbl, ent))
		case isInterface(ent.AST):
			view.Interfaces = append(view.Interfaces, toInterfaceView(tbl, ent))
		case isEnum(ent.AST):
			view.Enums = append(view.Enums, toEnumView(ent))
		}
	}

	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.go.tmpl"))

	for path, file := range viewMap {
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "file.go.tmpl", file); err != nil {
			return nil, fmt.Errorf("template execution failed: %w", err)
		}

		// F4: Run go/format
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("go/format failed on %s: %w\nRaw source:\n%s", path, err,
				buf.String())
		}

		files = append(files, &GeneratedFile{
			Path:    path,
			Content: formatted,
		})
	}

	return files, nil
}

func renderSourceStructRel(view *StructView, rel *resolver.RelationshipSymbol) {
	switch rel.Type {
	case ast.RelationInheritance:
		if isInterface(rel.Target.AST) {
			view.Implements = append(view.Implements, rel.Target.AST.Identifier)
		} else {
			view.Embeds = append(view.Embeds, rel.Target.AST.Identifier)
		}
	case ast.RelationRealization:
		view.Implements = append(view.Implements, rel.Target.AST.Identifier)
	}
}

func formatCompFieldType(ownerName string, mult ast.Cardinality) string {
	if mult == ast.UnknownCardinality {
		return ownerName
	}
	if mult.Min == 0 && mult.Max == 1 {
		return "*" + ownerName
	}
	if mult.Max == -1 {
		return "[]" + ownerName
	}
	if mult.Min == 1 && mult.Max == 1 {
		return ownerName
	}
	if mult.Max > 0 {
		return "[" + strconv.Itoa(mult.Max) + "]" + ownerName
	}
	return "unknown"
}

func formatAggFieldType(ownerName string, mult ast.Cardinality) string {
	if mult == ast.UnknownCardinality {
		return "*" + ownerName
	}
	if mult.Min == 0 && mult.Max == 1 {
		return "*" + ownerName
	}
	if mult.Max == -1 {
		return "[]*" + ownerName
	}
	if mult.Min == 1 && mult.Max == 1 {
		return "*" + ownerName
	}
	if mult.Max > 0 {
		return "[" + strconv.Itoa(mult.Max) + "]*" + ownerName
	}
	return "unknown"
}

func renderTargetStructRel(view *StructView, rel *resolver.RelationshipSymbol) {
	fieldView := FieldView{
		Name: rel.Source.AST.Identifier,
	}
	switch rel.Type {
	// need to clarify the difference between composition and aggregation
	case ast.RelationComposition:
		fieldView.Type = formatCompFieldType(rel.Source.AST.Identifier, rel.SourceMult)
	case ast.RelationAggregation:
		fieldView.Type = formatAggFieldType(rel.Source.AST.Identifier, rel.SourceMult)
	case ast.RelationInheritance:
		// do not output a warning, it will be handled in other struct
	default:
		log.Printf("warning: ignoring %s relationship between %s and %s", rel.Type, rel.Source.FQN, rel.Target.FQN)
	}
	view.Fields = append(view.Fields, fieldView)
}

func toStructView(tbl *resolver.SymbolTable, ent *resolver.EntitySymbol) StructView {
	view := StructView{
		Name:       ent.AST.Identifier,
		NotesView:  toNotes(ent.Notes),
		TriviaView: toTrivia(ent.AST.Trivia),
	}

	for _, rel := range tbl.Relationships {
		if rel.Source == ent {
			renderSourceStructRel(&view, rel)
		} else if rel.Target == ent {
			renderTargetStructRel(&view, rel)
		}
	}

	for _, member := range ent.AST.Members {
		switch member := member.(type) {
		case *dialect.GoField:
			view.Fields = append(view.Fields, toFieldView(ent.AST, member))
		case *dialect.GoMethod:
			view.Methods = append(view.Methods, toMethodView(ent.AST, member))
		}
	}

	return view
}

func toInterfaceView(tbl *resolver.SymbolTable, ent *resolver.EntitySymbol) InterfaceView {
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
			case ast.RelationInheritance:
				// do not output a warning, it will be handled in other struct
			default:
				log.Printf("warning: ignoring %s relationship between %s and %s", rel.Type, rel.Source.FQN, rel.Target.FQN)
			}
		}
	}

	for _, member := range ent.AST.Members {
		view.Methods = append(
			view.Methods,
			toMethodView(ent.AST, member.(*dialect.GoMethod)),
		)
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

func toNotes(note []*ast.Note) NotesView {
	var notes []string
	for _, n := range note {
		notes = append(notes, n.Text)
	}
	return NotesView{
		Notes: notes,
	}
}

func toTrivia(t ast.Trivia) TriviaView {
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

func ensureUpper(s string) string {
	firstByte := string(s[0])
	firstByte = strings.ToUpper(firstByte)
	return firstByte + s[1:]
}

func ensureLower(s string) string {
	firstByte := string(s[0])
	firstByte = strings.ToLower(firstByte)
	return firstByte + s[1:]
}

func isLower(s byte) bool {
	return s >= 'a' && s <= 'z'
}

func isUpper(s byte) bool {
	return s >= 'A' && s <= 'Z'
}

func ensureCorrectCase(ownerName string, fieldName string, visibility ast.VisibilityKind) string {
	newName := fieldName

	switch visibility {
	case ast.VisibilityProtected, ast.VisibilityPrivate:
		log.Printf("warning: %s field %s.%s is not supported in Go, using package encapsulation instead", visibility, ownerName, fieldName)
		fallthrough
	case ast.VisibilityPackage:
		if !isUpper(newName[0]) {
			return fieldName
		}
		newName = ensureLower(newName)
		log.Printf("warning: changing field name capitalization %s.%s to %s", ownerName, fieldName, newName)
	case ast.VisibilityPublic:
		if !isLower(newName[0]) {
			return fieldName
		}
		newName = ensureUpper(newName)
		log.Printf("warning: changing field name capitalization %s.%s to %s", ownerName, fieldName, newName)
	}

	return newName
}

func toFieldView(owner *ast.Entity, field *dialect.GoField) FieldView {
	return FieldView{
		Name: ensureCorrectCase(owner.Identifier, field.Name, field.Visibility),
		Type: field.Type.String(),
	}
}

func toMethodView(owner *ast.Entity, method *dialect.GoMethod) MethodView {
	return MethodView{
		Name:      ensureCorrectCase(owner.Identifier, method.Name, method.Visibility),
		Signature: method.Signature(),
	}
}
