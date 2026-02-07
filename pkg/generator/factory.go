package generator

import (
	"fmt"
	gogenerator "yur4uwe/pac/pkg/generator/go"
	"yur4uwe/pac/types"
)

func CodeGeneratorByLang(lang string) (types.CodeGenerator, error) {
	switch lang {
	case "go":
		return gogenerator.GoCodeGenerator{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}
