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
		fileImportPath := strings.Join(ent.PackagePath, "/")
		if len(ent.PackagePath) > 0 {
			filePath = fileImportPath + "/" + filePath
		}
		view, ok := viewMap[filePath]
		if !ok {
			pkgName := "root"
			if len(ent.PackagePath) > 0 {
				pkgName = ent.PackagePath[len(ent.PackagePath)-1]
			}
			view = &FileView{
				PackageName: pkgName,
				ImportPath:  fileImportPath,
			}
			viewMap[filePath] = view
		}

		switch {
		case isStruct(ent.AST):
			aux, err := toStructView(tbl, ent, view)
			if err != nil {
				return nil, err
			}
			view.Structs = append(view.Structs, aux)
		case isInterface(ent.AST):
			aux, err := toInterfaceView(tbl, ent, view)
			if err != nil {
				return nil, err
			}
			view.Interfaces = append(view.Interfaces, aux)
		case isEnum(ent.AST):
			view.Enums = append(view.Enums, toEnumView(ent))
		}
	}

	for _, note := range tbl.Notes {
		raw := strings.TrimSpace(note.Text)
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "Package ") {
			fields := strings.Fields(raw)
			var targetView *FileView
			if len(fields) >= 2 {
				pkgName := fields[1]
				for _, view := range viewMap {
					if view.PackageName == pkgName {
						targetView = view
						break
					}
				}
			}
			if targetView == nil {
				targetView = getRootOrFirstView(viewMap)
			}
			for line := range strings.SplitSeq(raw, "\n") {
				targetView.PackageComments = append(targetView.PackageComments, strings.TrimSpace(line))
			}
		} else {
			targetView := getRootOrFirstView(viewMap)
			nv := toNotesView([]*ast.Note{note})
			targetView.FileNotes = append(targetView.FileNotes, nv.Notes...)
		}
	}

	for _, file := range viewMap {
		// collapse incomplete import paths
		cleanImports := make([]string, 0, len(file.Imports))
		for _, incompleteImp := range file.Imports {
			if strings.Contains(incompleteImp, "/") || stdlib.IsStdlibPackage(strings.Split(incompleteImp, "/")) {
				cleanImports = append(cleanImports, incompleteImp)
				continue
			}
			var matchFound bool
			for _, fullImp := range file.Imports {
				if !strings.Contains(fullImp, "/") {
					// It is a partial import path
					continue
				}
				if strings.HasSuffix(fullImp, incompleteImp) {
					// we have found an incomplete import paths
					// that matches complete one so we ignore it
					matchFound = true
					break
				}
			}
			if !matchFound {
				// we have found an incomplete import path
				// that does not match any complete one
				// so we add it to the list
				if len(incompleteImp) > 0 {
					cleanImports = append(cleanImports, incompleteImp)
				}
			}
		}
		slices.Sort(cleanImports)
		file.Imports = slices.Clip(cleanImports)
	}

	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.go.tmpl"))

	for path, file := range viewMap {
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "file.go.tmpl", file); err != nil {
			return nil, fmt.Errorf("template execution failed: %w", err)
		}

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

func getRootOrFirstView(viewMap map[string]*FileView) *FileView {
	if view, ok := viewMap["types.go"]; ok {
		return view
	}
	for _, view := range viewMap {
		return view
	}
	root := &FileView{PackageName: "root"}
	viewMap["types.go"] = root
	return root
}

func fillSourceStructByRel(view *StructView, rel *resolver.RelationshipSymbol, fileView *FileView) {
	// i have not
	trivia := toTriviaView(rel.AST.Trivia)
	targetType := targetTypeName(rel.Source.PackagePath, rel.Target)
	fieldName := ""
	if customName, vis := reparseLabel(rel.AST.Label); customName != "" {
		fieldName = customName
		attachVisibilityComment(&trivia, vis)
	} else {
		fieldName = targetFieldName(rel.Target)
	}

	fieldView := FieldView{
		Name:       fieldName,
		NotesView:  toNotesView(rel.Notes),
		TriviaView: trivia,
	}

	if impPath, ok := stdlib.LookupImportPath(rel.Target.PackagePath); ok {
		fileView.AddImport(impPath)
	} else if len(rel.Target.PackagePath) > 0 && !slices.Equal(rel.Source.PackagePath, rel.Target.PackagePath) {
		fileView.AddImport(strings.Join(rel.Target.PackagePath, "/"))
	}

	switch rel.Type {
	case ast.RelationInheritance:
		if isInterface(rel.Target.AST) {
			view.Implements = append(view.Implements, targetType)
		} else {
			view.Embeds = append(view.Embeds, targetType)
		}
		if len(rel.Notes) > 0 {
			view.Notes = append(view.Notes, toNotesView(rel.Notes).Notes...)
		}
	case ast.RelationRealization:
		view.Implements = append(view.Implements, targetType)
		if len(rel.Notes) > 0 {
			view.Notes = append(view.Notes, toNotesView(rel.Notes).Notes...)
		}
	case ast.RelationComposition:
		fieldView.Type = formatCompFieldType(targetType, rel.TargetMult)
		view.Fields = append(view.Fields, fieldView)
	case ast.RelationAggregation, ast.RelationAssociation:
		fieldView.Type = formatAggFieldType(targetType, rel.TargetMult)
		view.Fields = append(view.Fields, fieldView)
	case ast.RelationDependency:
		var sb strings.Builder
		sb.WriteString("Depends on ")
		sb.WriteString(targetType)
		if rel.AST != nil && rel.AST.Label != "" {
			sb.WriteString(" (")
			sb.WriteString(rel.AST.Label)
			sb.WriteString(")")
		}
		view.LeadingTrivia = append(view.LeadingTrivia, sb.String())
		if len(rel.Notes) > 0 {
			view.Notes = append(view.Notes, toNotesView(rel.Notes).Notes...)
		}
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
