package ast

import (
	"errors"
	"fmt"

	"yur4uwe/pac/pkg/tokenizer"
)

var (
	ErrUnclosedBlock         = errors.New("unclosed block")
	ErrMissingIdentifier     = errors.New("expected identifier")
	ErrIncompletePackagePath = errors.New("incomplete qualified package path")
	ErrMissingTarget         = errors.New("expected relationship target")
	ErrMalformedArrow        = errors.New("malformed relationship arrow")
)

//go:generate enumer -type=Severity -json -transform=lower -trimprefix=Severity
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
	SeverityHint
)

type Diagnostic struct {
	ID       int                  `json:"id"`
	Severity Severity             `json:"severity"`
	Message  string               `json:"message"`
	Span     tokenizer.SourceSpan `json:"span"`
	Err      error                `json:"-"`
	Node     Node                 `json:"-"`
}

func (d *Diagnostic) Error() string {
	return d.Message
}

func (d *Diagnostic) Unwrap() error {
	return d.Err
}

type DiagnosticList []*Diagnostic

func (dl DiagnosticList) Error() string {
	switch len(dl) {
	case 0:
		return "no errors"
	case 1:
		return dl[0].Error()
	default:
		return fmt.Sprintf("%s (and %d more errors)", dl[0].Error(), len(dl)-1)
	}
}
