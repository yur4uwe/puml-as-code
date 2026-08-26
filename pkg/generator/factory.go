package generator

import (
	gogen "yur4uwe/pac/pkg/generator/go"
	"yur4uwe/pac/pkg/resolver"
)

type CodeGenerator interface {
	SemanticPass(tbl *resolver.SymbolTable) error
	GenerateFromClassDiagram(tbl *resolver.SymbolTable) error
}

var _ CodeGenerator = gogen.GoCodeGenerator{}

func CodeGeneratorByLang(lang string) (CodeGenerator, bool) {
	switch lang {
	case "go":
		return gogen.GoCodeGenerator{}, true
	default:
		return nil, false
	}
}
