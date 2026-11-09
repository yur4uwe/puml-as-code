package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"
)

func (p *Parser) parseNoteStyle(n *ast.Note) *ast.NoteStyle {
	firstHash, ok := p.stream.TryConsumeType(tokenizer.HASH)
	if !ok {
		return nil
	}
	mark := p.Mark(firstHash)

	noteStyle := &ast.NoteStyle{
		Span: firstHash.Span,
	}

	colorTok := p.stream.Emit()
	expr, err := p.parseColorExpr(colorTok)
	noteStyle.Span = p.Span(mark)
	if err != nil {
		p.ast.RecordError(err, colorTok.Span, n)
		return noteStyle
	}

	noteStyle.Background = expr
	return noteStyle
}

func (p *Parser) parseNote() *ast.Note {
	// has the parameter to match signature
	// tok is a keyword 'note'
	note := &ast.Note{}

	tok := p.stream.Emit()
	switch tok.Type {
	case tokenizer.STRING:
		p.parseInlineIdentNote(note, tok)
		return note
	case tokenizer.IDENTIFIER:
		class := keyword.Classify(tok.Literal)
		switch class {
		case keyword.Direction:
			p.parseDirectedNote(note, tok)
		case keyword.Position:
			p.parseLinkNote(note, tok)
		case keyword.Alias:
			p.parseMultilineAliasNote(note)
		default:
			err := NewParserErrorf(tok, "expected direction (left, right etc.), position or alias (as) after 'note', got %s", class.String())
			p.ast.RecordError(
				err, tok.Span, note,
			)
			p.syncToNoteBody()
			return note
		}
	default:
		err := NewParserError(tok, "unexpected token in note definition")
		p.ast.RecordError(
			err, tok.Span, note,
		)
		p.syncToLineEnd()
		return note
	}
	p.parseNoteBody(note)
	return note
}

// parseDirectedNote can resolve into either a NoteTargeted or NoteLink.
func (p *Parser) parseDirectedNote(note *ast.Note, dirTok tokenizer.Token) {
	note.Kind = ast.NoteTargeted
	note.Direction = p.mapTokenToDirection(dirTok)
	if relativeTok, ok := p.stream.TryConsumeKW(keyword.Position); ok {
		if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.STRING) {
			tok := p.stream.PeekTokenAt(0)
			p.ast.RecordError(
				NewParserError(tok, "Expected identifier for target"),
				tok.Span,
				note,
			)
			p.syncToNoteBody()
			return
		}
		target, err := p.parseTargetRef(p.stream.Emit())
		if err != nil {
			p.ast.RecordError(
				err, target.Span(), note,
			)
			p.syncToLineEnd()
			return
		}
		if relativeTok.Literal == "on" {
			if strings.ToLower(target.EntityName()) != "link" {
				err := NewParserError(relativeTok, "Unexpected identifier for a note link target")
				p.ast.RecordError(
					err, target.Span(), note,
				)
				p.syncToNoteBody()
				return
			}
			note.Kind = ast.NoteLink
			p.ast.RecordWarning(
				NewParserErrorf(dirTok, "direction %q is ignored for note on links", dirTok.Literal),
				dirTok.Span,
				note,
			)
		} else {
			note.Kind = ast.NoteTargeted
		}
		note.Target = &target
	} else if tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER); ok {
		err := NewParserError(tok, "Unexpected identifier after direction")
		p.ast.RecordError(
			err, tok.Span, note,
		)
		p.syncToNoteBody()
		return
	}
	note.Style = p.parseNoteStyle(note)
}

func (p *Parser) parseInlineIdentNote(note *ast.Note, stringTok tokenizer.Token) {
	note.Text = strings.Trim(stringTok.Literal, "\"")
	note.Kind = ast.NoteAlias
	if aliasTok, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		err := NewParserError(aliasTok, "expected alias keyword after note text")
		p.ast.RecordError(err, aliasTok.Span, note)
		p.syncToLineEnd()
		return
	}
	if !p.stream.AssertType(tokenizer.IDENTIFIER) {
		tok := p.stream.PeekTokenAt(0)
		err := NewParserError(tok, "Expected identifier after alias keyword")
		p.ast.RecordError(err, tok.Span, note)
		p.syncToLineEnd()
		return
	}
	note.Identifier = p.parseQualifiedName(p.stream.Emit())
	note.Style = p.parseNoteStyle(note)
	if !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
		nextTok := p.stream.PeekTokenAt(0)
		err := NewParserError(nextTok, "Unexpected tokens after inline alias note")
		p.ast.RecordError(err, nextTok.Span, note)
		p.syncToLineEnd()
		return
	}
}

func (p *Parser) parseMultilineAliasNote(note *ast.Note) {
	note.Kind = ast.NoteAlias
	if !p.stream.AssertType(tokenizer.IDENTIFIER) {
		tok := p.stream.PeekTokenAt(0)
		p.ast.RecordError(
			NewParserError(tok, "Expected identifier after alias keyword"),
			tok.Span,
			note,
		)
		p.syncToLineEnd()
		return
	}
	note.Identifier = p.parseQualifiedName(p.stream.Emit())
	note.Style = p.parseNoteStyle(note)
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		lastLineTok := p.stream.PeekTokenAt(0)
		p.ast.RecordError(
			NewParserError(p.stream.PeekTokenAt(0), "Expected newline after alias keyword"),
			lastLineTok.Span,
			note,
		)
		p.syncToLineEnd()
		return
	}
}

func (p *Parser) parseLinkNote(note *ast.Note, onTok tokenizer.Token) {
	if onTok.Literal != "on" {
		p.ast.RecordError(
			NewParserError(onTok, "Unexpected identifier after 'note'"),
			onTok.Span,
			note,
		)
		p.syncToNoteBody()
		return
	}
	linkTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok || linkTok.Literal != "link" {
		errTok := linkTok
		if !ok {
			errTok = p.stream.PeekTokenAt(0)
		}
		p.ast.RecordError(
			NewParserError(errTok, "Expected 'link' after 'note on'"),
			errTok.Span,
			note,
		)
		p.syncToNoteBody()
		return
	}
	note.Target = &ast.TargetRef{Name: ast.IdentFromToken(linkTok)}
	note.Kind = ast.NoteLink
	note.Style = p.parseNoteStyle(note)
}

func (p *Parser) parseNoteBody(note *ast.Note) {
	tok := p.stream.PeekTokenAt(0)
	switch tok.Type {
	case tokenizer.COLON:
		p.stream.Emit()
		note.Text = p.stream.ReadUntilNewline()
	case tokenizer.NEWLINE:
		note.TrailingTrivia = p.dumpTrivia()
		if !p.stream.HasCloserBeforeBoundary("end", "note") {
			p.ast.RecordError(
				ast.NewUnclosedBlockError("note", "end note"),
				tok.Span,
				note,
			)
			return
		}
		body, err := p.stream.ConsumeTextBlock("end", "note")
		if err != nil {
			return
		}
		note.Text = body
	default:
		p.stream.Emit()
		err := NewParserError(tok, "Expected ':' or newline after note definition")
		p.ast.RecordError(
			err, tok.Span, note,
		)
	}
}
