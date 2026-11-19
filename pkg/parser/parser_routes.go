package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) fallbackRoute(
	tok tokenizer.Token,
	stmt ast.Statement,
	err error,
	mark mark,
	leadingTrivia []ast.Comment,
	syncToBlockEnd bool,
) ast.Statement {
	if err == nil && stmt != nil {
		leadingTrivia = stmt.GetLeadingTrivia()
	}
	if err == nil {
		err = NewParserError(tok, "Unexpected token")
	}

	if syncToBlockEnd {
		p.syncToLineOrBlockEnd()
	} else {
		p.syncToLineEnd()
	}

	if !p.stream.AssertType(tokenizer.EOF) && !p.isEndMarker() {
		if p.stream.LastSemanticToken().Span.End.Offset < tok.Span.End.Offset {
			p.stream.Emit()
		}
	}

	span := p.span(mark)

	dgnst := &ast.Diagnostic{
		ID:      len(p.ast.Diagnostics),
		Message: err.Error(),
		Span:    span,
		Err:     err,
	}

	p.stream.EmitCommentToks()
	badStmt := &ast.BadStatement{
		Diagnostic: dgnst,
		BaseNode: ast.BaseNode{
			NodeSpan:       span,
			LeadingTrivia:  leadingTrivia,
			TrailingTrivia: p.dumpTrivia(),
		},
	}

	dgnst.Node = badStmt

	p.ast.RecordDiagnostic(dgnst)
	return badStmt
}

func (p *Parser) routeDiagramStatement(tok tokenizer.Token) ast.Statement {
	mark := p.mark(tok)
	leadingTrivia := p.dumpTrivia()

	stmt, err := p.parseDiagramOnlyStatement(tok)
	if err != nil {
		return p.fallbackRoute(tok, stmt, err, mark, leadingTrivia, false)
	}
	if stmt != nil {
		p.stream.EmitCommentToks()
		stmt.SetMetadata(p.span(mark), leadingTrivia, p.dumpTrivia())
		return stmt
	}

	stmt, err = p.parseContainerStatement(tok)
	if err != nil {
		return p.fallbackRoute(tok, stmt, err, mark, leadingTrivia, false)
	}
	if stmt != nil {
		p.stream.EmitCommentToks()
		stmt.SetMetadata(p.span(mark), leadingTrivia, p.dumpTrivia())
		return stmt
	}

	return p.fallbackRoute(tok, nil, nil, mark, leadingTrivia, false)
}

func (p *Parser) routeContainerStatement(tok tokenizer.Token) ast.Statement {
	mark := p.mark(tok)
	leadingTrivia := p.dumpTrivia()

	stmt, err := p.parseContainerStatement(tok)
	if err != nil {
		return p.fallbackRoute(tok, stmt, err, mark, leadingTrivia, true)
	}
	if stmt != nil {
		p.stream.EmitCommentToks()
		stmt.SetMetadata(p.span(mark), leadingTrivia, p.dumpTrivia())
		return stmt
	}

	return p.fallbackRoute(tok, stmt, err, mark, leadingTrivia, true)
}
