package gogenerator

import (
	"yur4uwe/pac/pkg/parser/ast"
)

type GoCodeGenerator struct{}

func (GoCodeGenerator) GenerateFromClassDiagram(diagram *ast.Diagram) (string, error) {
	return "", nil
}
