package parser

import "yur4uwe/pac/pkg/tokenizer"

func (p *Parser) parseDirective(tok tokenizer.Token) (any, error) {
	if tok.Type == tokenizer.EXCLAMATION {
		// Probably an include directive
	}
	return nil, nil
}

func (p *Parser) parseTitle() error {
	if !p.stream.Assert(tokenizer.NEWLINE) {
		// We are in luck and ints single line title
		p.ast.Title = p.stream.ReadRawUntilNewline()
		return nil
	}
	if _, ok := p.stream.Consume(tokenizer.NEWLINE); !ok {
		return NewParserError("Expected newline after multiline title", p.stream.PeekTokenAt(0).Pos)
	}

	// Otherwise we are inside a multiline title block
	title, found := p.stream.ReadBlock(tokenizer.TITLE)
	if !found {
		return NewParserError("Expected title block to end", p.stream.PeekTokenAt(0).Pos)
	}

	p.ast.Title = title
	return nil
}
