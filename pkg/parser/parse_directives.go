package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseDirective(tok1 tokenizer.Token) ast.Statement {
	directiveNameTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		err := NewParserError(p.stream.PeekTokenAt(0), "Expected directive name")
		p.ast.RecordError(err, tok1.Span, nil)
		p.syncToLineEnd()
		return &ast.UnhandledStatement{
			Raw: tok1.Literal,
			BaseNode: ast.BaseNode{
				LeadingTrivia: p.dumpTrivia(),
			},
		}
	}

	if directiveNameTok.Span.Start.Offset != tok1.Span.Start.Offset+1 {
		err := NewParserError(directiveNameTok, "Expected directive name right after !")
		p.ast.RecordError(err, directiveNameTok.Span, nil)
		p.syncToLineEnd()
		return &ast.UnhandledStatement{
			Raw: p.stream.SliceInput(tok1.Span.Start.Offset, directiveNameTok.Span.End.Offset),
			BaseNode: ast.BaseNode{
				LeadingTrivia: p.dumpTrivia(),
			},
		}
	}

	if strings.HasPrefix(directiveNameTok.Literal, "include") {
		return p.parseIncludeDirective(directiveNameTok)
	}
	return p.parseUnhandledDirective(tok1, directiveNameTok)
}

func (p *Parser) parseIncludeDirective(tok tokenizer.Token) *ast.IncludeDirective {
	var kind ast.IncludeKind
	switch tok.Literal {
	case "include_many":
		kind = ast.IncludeMany
	case "include_once", "include":
		kind = ast.IncludeOnce
	default:
		kind = ast.IncludeOnce
		err := NewParserError(tok, "Unknown include directive")
		p.ast.RecordError(err, tok.Span, nil)
	}

	dir := &ast.IncludeDirective{
		Kind: kind,
	}

	filePathToks := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.EXCLAMATION)
	dir.Path = p.stream.TokensToString(filePathToks)
	if p.stream.AssertType(tokenizer.EXCLAMATION) {
		p.stream.Emit() // consume '!'
		// Id or order must be a single token
		if !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
			lastDirTok := p.stream.Emit()
			dir.Tag = lastDirTok.Literal
		}
	}

	return dir
}
