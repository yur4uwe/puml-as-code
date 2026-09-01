// Package gogenerator provides a code generator and semantic pass for Go code.
package gogenerator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"yur4uwe/pac/pkg/generator/go/stdlib"
	"yur4uwe/pac/pkg/parser/ast"
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
			view.Structs = append(view.Structs, toStructView(tbl, ent, view))
		case isInterface(ent.AST):
			view.Interfaces = append(view.Interfaces, toInterfaceView(tbl, ent, view))
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

func fillSourceStructByRel(view *StructView, rel *resolver.RelationshipSymbol, fileView *FileView) {
	fieldView := FieldView{}
	targetIdent := targetTypeName(rel.Source.PackagePath, rel.Target)
	if customName := reparseLabel(rel.AST.Label); customName != "" {
		fieldView.Name = customName
	} else {
		fieldView.Name = targetIdent
	}

	if !slices.Equal(rel.Source.PackagePath, rel.Target.PackagePath) {
		fileView.AddImport(rel.Target.PackagePath[len(rel.Target.PackagePath)-1])
	}

	if stdlib.IsStdlibPackage(strings.Join(rel.Target.PackagePath, "/")) {
		fileView.AddImport(rel.Target.PackagePath[len(rel.Target.PackagePath)-1])
	}

	switch rel.Type {
	case ast.RelationInheritance:
		if isInterface(rel.Target.AST) {
			view.Implements = append(view.Implements, targetIdent)
		} else {
			view.Embeds = append(view.Embeds, targetIdent)
		}
	case ast.RelationRealization:
		view.Implements = append(view.Implements, targetIdent)
	case ast.RelationComposition:
		fieldView.Type = formatCompFieldType(targetIdent, rel.TargetMult)
		view.Fields = append(view.Fields, fieldView)
	case ast.RelationAggregation, ast.RelationAssociation:
		fieldView.Type = formatAggFieldType(targetIdent, rel.TargetMult)
		view.Fields = append(view.Fields, fieldView)
	}
}

func formatCompFieldType(ownerName string, mult ast.Cardinality) string {
	if mult == ast.UnknownCardinality || mult.Raw == "" || (mult.Min == 0 && mult.Max == 0) {
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
	return ownerName
}

func formatAggFieldType(ownerName string, mult ast.Cardinality) string {
	if mult == ast.UnknownCardinality || mult.Raw == "" || (mult.Min == 0 && mult.Max == 0) {
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
	return "*" + ownerName
}

// func renderTargetStructRel(view *StructView, rel *resolver.RelationshipSymbol) {
// 	switch rel.Type {
// 	// need to clarify the difference between composition and aggregation
// 	case ast.RelationInheritance, ast.RelationRealization:
// 		// do not output a warning, it will be handled in other struct
// 	default:
// 		log.Printf("warning: ignoring %s relationship between %s and %s", rel.Type, rel.Source.FQN, rel.Target.FQN)
// 	}
// }
