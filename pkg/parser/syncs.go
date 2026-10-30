package parser

import "yur4uwe/pac/pkg/tokenizer"

func (p *Parser) syncToLineEnd() []tokenizer.Token {
	var consumed []tokenizer.Token
	for !p.stream.AssertAnyType(
		tokenizer.NEWLINE,
		tokenizer.EOF,
	) && !p.isEndMarker() {
		consumed = append(consumed, p.stream.Emit())
	}
	return consumed
}

func (p *Parser) syncToLineOrBlockEnd() []tokenizer.Token {
	var consumed []tokenizer.Token
	for !p.stream.AssertAnyType(
		tokenizer.NEWLINE,
		tokenizer.EOF,
		tokenizer.RBRACE,
	) && !p.isEndMarker() {
		consumed = append(consumed, p.stream.Emit())
	}
	return consumed
}

func (p *Parser) syncToBlockStart() []tokenizer.Token {
	var consumed []tokenizer.Token
	for !p.stream.AssertAnyType(
		tokenizer.NEWLINE,
		tokenizer.EOF,
		tokenizer.LBRACE,
	) && !p.isEndMarker() {
		consumed = append(consumed, p.stream.Emit())
	}
	return consumed
}

func (p *Parser) syncToNoteBody() []tokenizer.Token {
	var consumed []tokenizer.Token
	for !p.stream.AssertAnyType(
		tokenizer.NEWLINE,
		tokenizer.EOF,
	) && !p.isEndMarker() {
		if p.stream.AssertType(tokenizer.COLON) && p.stream.PeekTokenAt(1).Type != tokenizer.COLON {
			break
		}
		consumed = append(consumed, p.stream.Emit())
	}
	return consumed
}
