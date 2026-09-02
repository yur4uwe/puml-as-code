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

func fillSourceStructByRel(view *StructView, rel *resolver.RelationshipSymbol, fileView *FileView) {
	fieldView := FieldView{}
	targetType := targetTypeName(rel.Source.PackagePath, rel.Target)
	if customName := reparseLabel(rel.AST.Label); customName != "" {
		fieldView.Name = customName
	} else {
		fieldView.Name = targetFieldName(rel.Target)
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
	case ast.RelationRealization:
		view.Implements = append(view.Implements, targetType)
	case ast.RelationComposition:
		fieldView.Type = formatCompFieldType(targetType, rel.TargetMult)
		view.Fields = append(view.Fields, fieldView)
	case ast.RelationAggregation, ast.RelationAssociation:
		fieldView.Type = formatAggFieldType(targetType, rel.TargetMult)
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
