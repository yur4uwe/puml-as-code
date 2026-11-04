package resolver

import (
	"fmt"

	"yur4uwe/pac/pkg/parser/ast"
)

type ResolutionError struct {
	Node ast.Node `json:"-"`
	Err  error    `json:"-"`
	Msg  string
}

func (re *ResolutionError) Error() string {
	return re.Msg
}

func (re *ResolutionError) Unwrap() error {
	return re.Err
}

func NewResolutionErrorf(node ast.Node, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	return &ResolutionError{
		Node: node,
		Msg:  err.Error(),
		Err:  err,
	}
}
