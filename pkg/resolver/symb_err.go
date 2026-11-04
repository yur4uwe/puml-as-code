package resolver

import "yur4uwe/pac/pkg/parser/ast"

type ResolutionError struct {
	Node ast.Node
	Msg  string
	Err  error
}

func (re *ResolutionError) Error() string {
	return re.Msg
}
