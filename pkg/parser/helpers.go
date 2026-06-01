package parser

import "yur4uwe/pac/pkg/tokenizer"

func unamb(tok tokenizer.TokenType) tokenizer.Token {
	return tokenizer.Token{Type: tok}
}

func amb(tok tokenizer.TokenType, literal string) tokenizer.Token {
	return tokenizer.Token{Type: tok, Literal: literal}
}

func (p *Parser) parseDirective(tok tokenizer.Token) (any, error) {
	if tok.Type == tokenizer.EXCLAMATION {
		// Probably an include directive
	}
	return nil, nil
}

func (p *Parser) parseTitle() error {
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		// We are in luck and ints single line title
		p.ast.Title = p.stream.ReadRawUntilNewline()
		return nil
	}
	if _, ok := p.stream.ConsumeType(tokenizer.NEWLINE); !ok {
		return NewParserError("Expected newline after multiline title", p.stream.PeekTokenAt(0).Pos)
	}

	// Otherwise we are inside a multiline title block
	title, err := p.stream.ReadBlock(unamb(tokenizer.END_BLOCK), unamb(tokenizer.TITLE))
	if err != nil {
		return NewParserError("Expected title block to end", p.stream.PeekTokenAt(0).Pos)
	}

	p.ast.Title = title
	return nil
}

func (p *Parser) parseStyles(tok tokenizer.Token) error {
	if tok.Type != tokenizer.LANGLE {
		// <style> token sequence for now simply read it as a string
		// and ignore it
		// I probebly should actually assert internal value of the IDENTIFIER token
		seq := []tokenizer.Token{unamb(tokenizer.LANGLE), unamb(tokenizer.SLASH), amb(tokenizer.IDENTIFIER, "style"), unamb(tokenizer.RANGLE)}
		styles, err := p.stream.ReadBlock(seq...)
		if err != nil {
			return NewParserError("Expected style block to end", p.stream.PeekTokenAt(0).Pos)
		}
		p.styles = append(p.styles, styles)
		return nil
	}
	return nil
}
