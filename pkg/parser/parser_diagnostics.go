package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) recordDiagnostic(d *ast.Diagnostic) {
	if p.ast == nil {
		p.ast = &ast.Diagram{}
	}
	p.ast.Diagnostics = append(p.ast.Diagnostics, d)
}

func (p *Parser) newDiagnosticError(err error, span tokenizer.SourceSpan, node ast.Node) *ast.Diagnostic {
	if p.ast == nil {
		p.ast = &ast.Diagram{}
	}
	return &ast.Diagnostic{
		ID:       len(p.ast.Diagnostics),
		Severity: ast.SeverityError,
		Message:  err.Error(),
		Span:     span,
		Err:      err,
		Node:     node,
	}
}

func (p *Parser) newDiagnosticWarning(err error, span tokenizer.SourceSpan, node ast.Node) *ast.Diagnostic {
	if p.ast == nil {
		p.ast = &ast.Diagram{}
	}
	return &ast.Diagnostic{
		ID:       len(p.ast.Diagnostics),
		Severity: ast.SeverityWarning,
		Message:  err.Error(),
		Span:     span,
		Err:      err,
		Node:     node,
	}
}

func (p *Parser) newDiagnosticInfo(err error, span tokenizer.SourceSpan, node ast.Node) *ast.Diagnostic {
	if p.ast == nil {
		p.ast = &ast.Diagram{}
	}
	return &ast.Diagnostic{
		ID:       len(p.ast.Diagnostics),
		Severity: ast.SeverityInfo,
		Message:  err.Error(),
		Span:     span,
		Err:      err,
		Node:     node,
	}
}
