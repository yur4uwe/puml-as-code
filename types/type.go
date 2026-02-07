package types

import "yur4uwe/pac/pkg/parser/ast"

type CodeGenerator interface {
	GenerateFromClassDiagram(diagram *ast.ClassDiagram) (string, error)
}
