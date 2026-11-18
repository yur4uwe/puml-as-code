package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"
)

func (p *Parser) parseNoteStyle(n *ast.Note) (*ast.NoteStyle, error) {
	firstHash, ok := p.stream.TryConsumeType(tokenizer.HASH)
	if !ok {
		return nil, nil
	}
	mark := p.Mark(firstHash)

	noteStyle := &ast.NoteStyle{
		Span: firstHash.Span,
	}

	colorTok := p.stream.Emit()
	expr, err := p.parseColorExpr(colorTok)
	noteStyle.Span = p.Span(mark)
	if err != nil {
		return nil, err
	}

	noteStyle.Background = expr
	return noteStyle, nil
}

func (p *Parser) parseNote() (ast.Statement, error) {
	// tok is a keyword 'note'
	note := &ast.Note{}

	tok := p.stream.Emit()
	switch tok.Type {
	case tokenizer.STRING:
		if err := p.parseInlineIdentNote(note, tok); err != nil {
			return nil, err
		}
		return note, nil
	case tokenizer.IDENTIFIER:
		class := keyword.Classify(tok.Literal)
		switch class {
		case keyword.Direction:
			if err := p.parseDirectedNote(note, tok); err != nil {
				return nil, err
			}
		case keyword.Position:
			if err := p.parseLinkNote(note, tok); err != nil {
				return nil, err
			}
		case keyword.Alias:
			if err := p.parseMultilineAliasNote(note); err != nil {
				return nil, err
			}
		default:
			return nil, NewParserErrorf(tok, "expected direction (left, right etc.), position or alias (as) after 'note', got %s", class.String())
		}
	default:
		return nil, NewParserError(tok, "unexpected token in note definition")
	}
	if err := p.parseNoteBody(note); err != nil {
		return nil, err
	}
	return note, nil
}

// parseDirectedNote can resolve into either a NoteTargeted or NoteLink.
func (p *Parser) parseDirectedNote(note *ast.Note, dirTok tokenizer.Token) error {
	note.Kind = ast.NoteTargeted
	note.Direction = p.mapTokenToDirection(dirTok)
	if relativeTok, ok := p.stream.TryConsumeKW(keyword.Position); ok {
		if !p.stream.AssertAnyType(tokenizer.IDENTIFIER, tokenizer.STRING) {
			tok := p.stream.PeekTokenAt(0)
			return NewParserError(tok, "Expected identifier for target")
		}
		target, err := p.parseTargetRef(p.stream.Emit())
		if err != nil {
			return err
		}
		if relativeTok.Literal == "on" {
			if strings.ToLower(target.EntityName()) != "link" {
				return NewParserError(relativeTok, "Unexpected identifier for a note link target")
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
		return NewParserError(tok, "Unexpected identifier after direction")
	}
	style, err := p.parseNoteStyle(note)
	if err != nil {
		return err
	}
	note.Style = style
	return nil
}

func (p *Parser) parseInlineIdentNote(note *ast.Note, stringTok tokenizer.Token) error {
	note.Text = strings.Trim(stringTok.Literal, "\"")
	note.Kind = ast.NoteAlias
	if aliasTok, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		return NewParserError(aliasTok, "expected alias keyword after note text")
	}
	if !p.stream.AssertType(tokenizer.IDENTIFIER) {
		tok := p.stream.PeekTokenAt(0)
		return NewParserError(tok, "Expected identifier after alias keyword")
	}
	note.Identifier = p.parseQualifiedName(p.stream.Emit())
	style, err := p.parseNoteStyle(note)
	if err != nil {
		return err
	}
	note.Style = style
	if !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
		nextTok := p.stream.PeekTokenAt(0)
		return NewParserError(nextTok, "Unexpected tokens after inline alias note")
	}
	return nil
}

func (p *Parser) parseMultilineAliasNote(note *ast.Note) error {
	note.Kind = ast.NoteAlias
	if !p.stream.AssertType(tokenizer.IDENTIFIER) {
		tok := p.stream.PeekTokenAt(0)
		return NewParserError(tok, "Expected identifier after alias keyword")
	}
	note.Identifier = p.parseQualifiedName(p.stream.Emit())
	style, err := p.parseNoteStyle(note)
	if err != nil {
		return err
	}
	note.Style = style
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		return NewParserError(p.stream.PeekTokenAt(0), "Expected newline after alias keyword")
	}
	return nil
}

func (p *Parser) parseLinkNote(note *ast.Note, onTok tokenizer.Token) error {
	if onTok.Literal != "on" {
		return NewParserError(onTok, "Unexpected identifier after 'note'")
	}
	linkTok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok || linkTok.Literal != "link" {
		errTok := linkTok
		if !ok {
			errTok = p.stream.PeekTokenAt(0)
		}
		return NewParserError(errTok, "Expected 'link' after 'note on'")
	}
	note.Target = &ast.TargetRef{Name: ast.IdentFromToken(linkTok)}
	note.Kind = ast.NoteLink
	style, err := p.parseNoteStyle(note)
	if err != nil {
		return err
	}
	note.Style = style
	return nil
}

func (p *Parser) parseNoteBody(note *ast.Note) error {
	tok := p.stream.PeekTokenAt(0)
	switch tok.Type {
	case tokenizer.COLON:
		p.stream.Emit()
		note.Text = p.stream.ReadUntilNewline()
		return nil
	case tokenizer.NEWLINE:
		note.TrailingTrivia = p.dumpTrivia()
		if !p.stream.HasCloserBeforeBoundary("end", "note") {
			p.ast.RecordError(
				ast.NewUnclosedBlockError("note", "end note"),
				tok.Span,
				note,
			)
			return nil
		}
		body, err := p.stream.ConsumeTextBlock("end", "note")
		if err != nil {
			return err
		}
		note.Text = body
		return nil
	default:
		return NewParserError(tok, "Expected ':' or newline after note definition")
	}
}
