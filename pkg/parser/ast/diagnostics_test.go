package ast

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/tokenizer"
)

func TestRecordDiagnosticMessages(t *testing.T) {
	diagram := &Diagram{}
	node := &Entity{}
	span := tokenizer.SourceSpan{Start: tokenizer.Pos{Line: 3, Col: 4}}
	plain := diagram.RecordErrorMessage(span, node, "invalid % name")
	warning := diagram.RecordWarningf(span, node, "check class: %w", ErrMissingIdentifier)
	require.Equal(t, "invalid % name", plain.Message)
	require.Equal(t, SeverityError, plain.Severity)
	require.Equal(t, SeverityWarning, warning.Severity)
	require.True(t, errors.Is(warning, ErrMissingIdentifier))
	require.Same(t, node, warning.Node)
	require.Equal(t, span, warning.Span)
	require.Equal(t, 1, warning.ID)
	require.Equal(t, DiagnosticList{plain, warning}, diagram.Diagnostics)
}
