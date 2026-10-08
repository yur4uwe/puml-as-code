package formatter

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (s *formatterState) emitTrivia(trivia tokenizer.Token) {
	s.emitRawSpan(trivia.Span)
}

func (s *formatterState) emitBlockStartTrivia(block ast.Node) {
	for _, trivia := range block.GetTrailingTrivia() {
		if trivia.Span.Start.Line == block.Span().Start.Line {
			s.buf.WriteByte(' ')
			s.emitTrivia(trivia)
		}
	}
}

func (s *formatterState) emitBlockEndTrivia(block ast.Node) {
	for _, trivia := range block.GetTrailingTrivia() {
		// To account for blocked statements
		// only emit trailing trivia on the same line as the end of the statement
		if trivia.Span.Start.Line != block.Span().Start.Line {
			s.buf.WriteByte(' ')
			s.emitTrivia(trivia)
		}
	}
}
