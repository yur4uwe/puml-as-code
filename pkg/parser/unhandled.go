package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseUnhandled(tok tokenizer.Token) (ast.Statement, error) {
	leadingTrivia := p.stream.DumpCollectedTrivia()
	var lineToks []tokenizer.Token
	for !p.stream.AssertType(tokenizer.NEWLINE) && !p.stream.AssertType(tokenizer.EOF) {
		lineToks = append(lineToks, p.stream.Emit())
	}
	endOffset := tok.Pos.Offset + uint(len(tok.Literal))
	endPos := tok.Pos
	if len(lineToks) > 0 {
		lastTok := lineToks[len(lineToks)-1]
		endOffset = lastTok.Pos.Offset + uint(len(lastTok.Literal))
		endPos = lastTok.Pos
	}
	p.stream.TryConsumeType(tokenizer.NEWLINE)
	p.stream.EmitCommentToks()
	trailingTrivia := p.stream.DumpCollectedTrivia()

	fullText := p.stream.SliceInput(tok.Pos.Offset, endOffset)
	return ast.UnhandledStatement{
		Text: fullText,
		Span: tokenizer.SourceSpan{
			Start: tok.Pos,
			End:   endPos,
		},
		Trivia: ast.Trivia{
			LeadingTrivia:  leadingTrivia,
			TrailingTrivia: trailingTrivia,
		},
	}, nil
}

func (p *Parser) parseUnhandledDirective(dirNameTok tokenizer.Token) (ast.Statement, error) {
	unhandledStmt := ast.UnhandledStatement{
		Trivia: ast.Trivia{
			LeadingTrivia: p.stream.DumpCollectedTrivia(),
		},
		Span: tokenizer.SourceSpan{
			Start: dirNameTok.Pos,
		},
	}

	return unhandledStmt, nil
}
