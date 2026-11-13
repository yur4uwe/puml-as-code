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

func (d *Diagram) RecordDiagnostic(diag *Diagnostic) *Diagnostic {
	if d == nil {
		return diag
	}
	diag.ID = len(d.Diagnostics)
	d.Diagnostics = append(d.Diagnostics, diag)
	return diag
}

func (d *Diagram) RecordError(err error, span tokenizer.SourceSpan, node Node) *Diagnostic {
	return d.RecordDiagnostic(&Diagnostic{
		Severity: SeverityError,
		Message:  err.Error(),
		Span:     span,
		Err:      err,
		Node:     node,
	})
}

// RecordErrorMessage records a new error at span.
func (d *Diagram) RecordErrorMessage(span tokenizer.SourceSpan, node Node, message string) *Diagnostic {
	return d.RecordError(errors.New(message), span, node)
}

// RecordErrorf records a formatted error at span. It supports %w wrapping.
func (d *Diagram) RecordErrorf(span tokenizer.SourceSpan, node Node, format string, args ...any) *Diagnostic {
	return d.RecordError(fmt.Errorf(format, args...), span, node)
}

func (d *Diagram) RecordWarning(err error, span tokenizer.SourceSpan, node Node) *Diagnostic {
	return d.RecordDiagnostic(&Diagnostic{
		Severity: SeverityWarning,
		Message:  err.Error(),
		Span:     span,
		Err:      err,
		Node:     node,
	})
}

// RecordWarningMessage records a warning at span.
func (d *Diagram) RecordWarningMessage(span tokenizer.SourceSpan, node Node, message string) *Diagnostic {
	return d.RecordWarning(errors.New(message), span, node)
}

// RecordWarningf records a formatted warning at span. It supports %w wrapping.
func (d *Diagram) RecordWarningf(span tokenizer.SourceSpan, node Node, format string, args ...any) *Diagnostic {
	return d.RecordWarning(fmt.Errorf(format, args...), span, node)
}

func (d *Diagram) RecordInfo(err error, span tokenizer.SourceSpan, node Node) *Diagnostic {
	return d.RecordDiagnostic(&Diagnostic{
		Severity: SeverityInfo,
		Message:  err.Error(),
		Span:     span,
		Err:      err,
		Node:     node,
	})
}
