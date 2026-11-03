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

func (p *Parser) newDiagnostic(severity ast.Severity, err error, span tokenizer.SourceSpan, node ast.Node) *ast.Diagnostic {
	if p.ast == nil {
		p.ast = &ast.Diagram{}
	}
	return &ast.Diagnostic{
		ID:       len(p.ast.Diagnostics),
		Severity: severity,
		Message:  err.Error(),
		Span:     span,
		Err:      err,
		Node:     node,
	}
}

func (p *Parser) recordError(err error, span tokenizer.SourceSpan, node ast.Node) *ast.Diagnostic {
	d := p.newDiagnostic(ast.SeverityError, err, span, node)
	p.recordDiagnostic(d)
	return d
}

func (p *Parser) recordWarning(err error, span tokenizer.SourceSpan, node ast.Node) *ast.Diagnostic {
	d := p.newDiagnostic(ast.SeverityWarning, err, span, node)
	p.recordDiagnostic(d)
	return d
}

func (p *Parser) recordInfo(err error, span tokenizer.SourceSpan, node ast.Node) *ast.Diagnostic {
	d := p.newDiagnostic(ast.SeverityInfo, err, span, node)
	p.recordDiagnostic(d)
	return d
}
