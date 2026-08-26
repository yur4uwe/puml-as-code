package gogenerator

import (
	"yur4uwe/pac/pkg/resolver"
)

type GoCodeGenerator struct{}

func (GoCodeGenerator) GenerateFromClassDiagram(tbl *resolver.SymbolTable) error {
	return nil
}
