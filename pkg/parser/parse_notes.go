package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) finishInline(mark Mark, n ast.InlineNote) (ast.InlineNote, error) {
	n.NodeSpan = p.Span(mark)
	p.stream.EmitCommentToks()
	n.TrailingTrivia = append(n.TrailingTrivia, p.stream.DumpCollectedTrivia()...)
	return n, nil
}

func (p *Parser) finishBlock(mark Mark, n ast.BlockNote) (ast.BlockNote, error) {
	n.NodeSpan = p.Span(mark)
	p.stream.EmitCommentToks()
	n.TrailingTrivia = append(n.TrailingTrivia, p.stream.DumpCollectedTrivia()...)
	return n, nil
}

func (p *Parser) parseRelativeOrLinkNoteHeader(note *ast.InlineNote, dirTok tokenizer.Token) error {
	note.Kind = ast.NoteRelative
	note.Direction = p.mapTokenToDirection(dirTok)
	if relativeTok, ok := p.stream.TryConsumeKW(keyword.Position); ok {
		target, err := p.parseTargetRef(p.stream.Emit()) // consume target
		if err != nil {
			return err
		}
		if relativeTok.Literal == "on" {
			if strings.ToLower(target.Entity) != "link" {
				return NewParserError(relativeTok, "Unexpected identifier for a note link target")
			}
			note.Kind = ast.NoteLink
		} else {
			note.Kind = ast.NoteRelative
		}
		note.Target = &target
	} else if tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER); ok {
		return NewParserError(tok, "Unexpected identifier after direction")
	}
	p.tryParseColor()
	return nil
}

func (p *Parser) parseInlineIdentNoteHeader(note *ast.InlineNote, stringTok tokenizer.Token) error {
	note.Text = stringTok.Literal
	note.Kind = ast.NoteAlias
	if aliasTok, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		return NewParserError(aliasTok, "Expected alias keyword after note text")
	}
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return NewParserError(tok, "Expected identifier after alias keyword")
	}
	note.Identifier = tok.Literal
	p.tryParseColor()
	if !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
		return NewParserError(p.stream.PeekTokenAt(0), "Unexpected tokens after inline alias note")
	}
	return nil
}

func (p *Parser) parseMultilineAliasNoteHeader(note *ast.BlockNote) error {
	note.Kind = ast.NoteAlias
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return NewParserError(tok, "Expected identifier after alias keyword")
	}
	note.Identifier = tok.Literal
	p.tryParseColor()
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		return NewParserError(tok, "Expected newline after alias keyword")
	}
	return nil
}

func (p *Parser) parseLinkNoteHeader(note *ast.InlineNote, onTok tokenizer.Token) error {
	if onTok.Literal != "on" {
		return NewParserError(onTok, "Unexpected identifier after 'note'")
	}
	if _, ok := p.stream.TryConsume(amb(tokenizer.IDENTIFIER, "link")); !ok {
		return NewParserError(onTok, "Expected 'link' after 'note on'")
	}
	note.Target = &ast.TargetRef{Entity: "link"}
	note.Kind = ast.NoteLink
	p.tryParseColor()
	return nil
}

func (p *Parser) parseNoteBodyAndFinish(mark Mark, note ast.InlineNote) (ast.Statement, error) {
	tok := p.stream.PeekTokenAt(0)
	switch tok.Type {
	case tokenizer.COLON:
		p.stream.Emit()
		note.Text = p.stream.ReadUntilNewline()
		return p.finishInline(mark, note)
	case tokenizer.NEWLINE:
		note.TrailingTrivia = p.stream.DumpCollectedTrivia()
		body, err := p.stream.ConsumeTextBlock("end", "note")
		if err != nil {
			return nil, err
		}
		bn := ast.BlockNote(note)
		bn.Text = body
		return p.finishBlock(mark, bn)
	default:
		p.stream.Emit()
		return nil, NewParserError(tok, "Expected ':' or newline after note definition")
	}
}
