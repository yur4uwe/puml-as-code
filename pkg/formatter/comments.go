package formatter

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (s *formatterState) emitComment(comment tokenizer.Token) {
	s.emitRawSpan(comment.Span)
}

func (s *formatterState) emitBlockStartTrivia(block ast.Statement) {
	for _, comment := range block.GetTrailingTrivia() {
		if comment.Span.Start.Line == block.Span().Start.Line {
			s.emitComment(comment)
		}
	}
}

func (s *formatterState) emitBlockEndTrivia(block ast.Statement) {
	for _, comment := range block.GetTrailingTrivia() {
		// To account for blocked statements
		// only emit trailing trivia on the same line as the end of the statement
		if comment.Span.Start.Line != block.Span().Start.Line {
			s.emitComment(comment)
		}
	}
}
