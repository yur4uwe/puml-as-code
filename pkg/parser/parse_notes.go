package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseRelativeNote(note *ast.Note, dirTok tokenizer.Token) error {
	note.Direction = p.mapTokenToDirection(dirTok)
	if relativeTok, ok := p.stream.TryConsumeKW(keyword.Position); ok {
		target, err := p.parseTargetRef(p.stream.Emit()) // consume target
		if err != nil {
			return err
		}
		if strings.ToLower(target.Entity) != "link" && relativeTok.Literal == "on" {
			return NewParserError("Unexpected identifier for a note link target", relativeTok)
		}
		note.Target = &target
	} else if tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER); ok {
		return NewParserError("Unexpected identifier after direction", tok)
	}
	p.tryParseColor()
	return p.parseNoteBody(note)
}

func (p *Parser) parseInlineIdentNote(note *ast.Note, stringTok tokenizer.Token) error {
	note.Text = stringTok.Literal
	if aliasTok, ok := p.stream.TryConsumeKW(keyword.Alias); !ok {
		return NewParserError("Expected alias keyword after note text", aliasTok)
	}
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return NewParserError("Expected identifier after alias keyword", tok)
	}
	note.Identifier = tok.Literal
	p.tryParseColor()
	if !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
		return NewParserError("Unexpected tokens after inline alias note", p.stream.PeekTokenAt(0))
	}
	return nil
}

func (p *Parser) parseMultilineAliasNote(note *ast.Note) error {
	tok, ok := p.stream.TryConsumeType(tokenizer.IDENTIFIER)
	if !ok {
		return NewParserError("Expected identifier after alias keyword", tok)
	}
	p.tryParseColor()
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		return NewParserError("Expected newline after alias keyword", tok)
	}
	return p.parseNoteBody(note)
}

func (p *Parser) parseLinkNote(note *ast.Note, onTok tokenizer.Token) error {
	if onTok.Literal != "on" {
		return NewParserError("Unexpected identifier after 'note'", onTok)
	}
	if _, ok := p.stream.TryConsume(amb(tokenizer.IDENTIFIER, "link")); !ok {
		return NewParserError("Expected 'link' after 'note on'", onTok)
	}
	note.Target = &ast.TargetRef{Entity: "link"}
	p.tryParseColor()
	return p.parseNoteBody(note)
}

func (p *Parser) parseNoteBody(note *ast.Note) error {
	tok := p.stream.PeekTokenAt(0)
	switch tok.Type {
	case tokenizer.COLON:
		p.stream.Emit()
		note.Text = p.stream.ReadUntilNewline()
		return nil
	case tokenizer.NEWLINE:
		note.TrailingTrivia = p.stream.DumpCollectedTrivia()
		body, err := p.stream.ConsumeTextBlock("end", "note")
		if err != nil {
			return err
		}
		note.Text = body
		return nil
	default:
		p.stream.Emit()
		return NewParserError("Expected ':' or newline after note definition", tok)
	}
}
