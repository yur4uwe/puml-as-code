package generator

import (
	"fmt"
	gogenerator "yur4uwe/pac/pkg/generator/go"
	"yur4uwe/pac/pkg/parser/ast"
)

type CodeGenerator interface {
	GenerateFromClassDiagram(diagram *ast.Diagram) (string, error)
}

var _ CodeGenerator = gogenerator.GoCodeGenerator{}

func CodeGeneratorByLang(lang string) (CodeGenerator, error) {
	switch lang {
	case "go":
		return gogenerator.GoCodeGenerator{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}
