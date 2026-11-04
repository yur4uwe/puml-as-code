package parser

import (
	"errors"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (p *Parser) parseSkinparamRoot(tok tokenizer.Token) ast.Statement {
	return p.parseSkinparamStatement(tok, true)
}

func (p *Parser) parseSkinparamTarget(firstTok tokenizer.Token) ([]tokenizer.Token, *ast.Stereotype, error) {
	var tokens []tokenizer.Token
	var stereo *ast.Stereotype

	switch firstTok.Type {
	case tokenizer.DOT, tokenizer.UNDERSCORE, tokenizer.DASH, tokenizer.IDENTIFIER,
		tokenizer.NUMBER:
		tokens = append(tokens, firstTok)
	default:
		return nil, nil, NewParserError(firstTok, "expected valid skinparam target name")
	}
	prevTokEnd := firstTok.EndOffset()

	// Collect contiguous key tokens (e.g. "class.BackgroundColor")
	for {
		nextTok := p.stream.PeekTokenAt(0)
		if nextTok.Span.Start.Offset != prevTokEnd {
			break
		}
		switch nextTok.Type {
		case tokenizer.NEWLINE, tokenizer.SEMICOLON, tokenizer.EOF, tokenizer.LBRACE:
			return tokens, stereo, nil
		case tokenizer.DOT, tokenizer.UNDERSCORE, tokenizer.DASH, tokenizer.IDENTIFIER, tokenizer.NUMBER:
			tokens = append(tokens, p.stream.Emit())
			prevTokEnd = nextTok.EndOffset()
		case tokenizer.LANGLE:
			if stereo != nil {
				return tokens, stereo, NewParserError(nextTok, "multiple stereotypes")
			}
			var err error
			stereo, _, err = p.tryParseStereotype()
			if errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
				return tokens, stereo, NewParserError(nextTok, "unexpected token in stereotype target declaration")
			}
			if err != nil {
				return tokens, stereo, err
			}
			prevTokEnd = stereo.FullSpan.End.Offset
		default:
			return tokens, stereo, NewParserError(nextTok, "unexpected token in stereotype target declaration")
		}
	}

	if stereo != nil && p.stream.AssertTypeSeq(tokenizer.LANGLE, tokenizer.LANGLE) {
		return tokens, stereo, NewParserError(p.stream.PeekTokenAt(0), "multiple stereotypes")
	}
	if stereo == nil {
		var err error
		stereo, _, err = p.tryParseStereotype()
		if err != nil && !errors.Is(err, tokenizer.ErrStartMarkerNotFound) {
			return tokens, nil, err
		}
	}

	return tokens, stereo, nil
}

func (p *Parser) parseSkinparamStatement(tok tokenizer.Token, isRoot bool) ast.Statement {
	var keyword string
	var targetTokens []tokenizer.Token
	var stereo *ast.Stereotype
	var err error
	var nameStartTok tokenizer.Token

	if isRoot {
		keyword = tok.Literal

		// Check for anonymous block: `skinparam { ... }`
		if lbraceTok, ok := p.stream.TryConsumeType(tokenizer.LBRACE); ok {
			p.stream.EmitCommentToks()
			trailingTrivia := p.dumpTrivia()
			block := &ast.SkinparamBlock{
				BaseNode: ast.BaseNode{
					TrailingTrivia: trailingTrivia,
				},
				Keyword: keyword,
				LBrace:  lbraceTok.Span.Start,
			}
			p.parseSkinparamBlockMembers(block)
			return block
		}

		if p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
			nextTok := p.stream.PeekTokenAt(0)
			err := NewParserError(nextTok, "Expected target or parameter after skinparam")
			setting := &ast.SkinparamSetting{
				Keyword: keyword,
			}
			p.ast.RecordError(err, nextTok.Span, setting)
			return setting
		}

		nameStartTok = p.stream.Emit()
	} else {
		nameStartTok = tok
	}

	targetTokens, stereo, err = p.parseSkinparamTarget(nameStartTok)
	if err != nil {
		if isRoot {
			setting := &ast.SkinparamSetting{
				Keyword: keyword,
			}
			p.ast.RecordError(err, nameStartTok.Span, setting)
			p.syncToLineEnd()
			return setting
		}
		p.ast.RecordError(err, nameStartTok.Span, nil)
		p.syncToLineOrBlockEnd()
		return nil
	}

	if lbraceTok, ok := p.stream.TryConsumeType(tokenizer.LBRACE); ok {
		p.stream.EmitCommentToks()
		trailingTrivia := p.dumpTrivia()
		block := &ast.SkinparamBlock{
			BaseNode: ast.BaseNode{
				TrailingTrivia: trailingTrivia,
			},
			Keyword:    keyword,
			Name:       ast.IdentFromTokens(targetTokens...),
			Stereotype: stereo,
			LBrace:     lbraceTok.Span.Start,
		}
		p.parseSkinparamBlockMembers(block)
		return block
	}

	toks := p.stream.ConsumeUntilType(tokenizer.NEWLINE, tokenizer.SEMICOLON)
	value := p.stream.SliceInputEnclosingTokens(toks...)
	p.stream.TryConsumeType(tokenizer.SEMICOLON)
	return &ast.SkinparamSetting{
		Keyword:    keyword,
		Name:       ast.IdentFromTokens(targetTokens...),
		Value:      value,
		Stereotype: stereo,
	}
}

func (p *Parser) parseSkinparamBlockMembers(block *ast.SkinparamBlock) {
	lbraceSpan := tokenizer.SourceSpan{
		Start: block.LBrace,
		End:   tokenizer.Pos{Offset: block.LBrace.Offset + 1, Line: block.LBrace.Line, Col: block.LBrace.Col + 1},
	}
	unclosedErr := ast.NewUnclosedBlockError("skinparam", "}")

	for {
		if p.isEndMarker() || p.stream.AssertType(tokenizer.EOF) {
			p.ast.RecordError(unclosedErr, lbraceSpan, block)
			break
		}
		if p.stream.AssertType(tokenizer.RBRACE) {
			break
		}

		tok := p.stream.Emit()
		switch tok.Type {
		case tokenizer.NEWLINE, tokenizer.SEMICOLON:
			continue
		case tokenizer.EXCLAMATION:
			memberMark := p.Mark(tok)
			leadingTrivia := p.dumpTrivia()
			member := p.parseDirective(tok)
			if member != nil {
				p.stream.EmitCommentToks()
				member.SetMetadata(p.Span(memberMark), leadingTrivia, p.dumpTrivia())
				block.Statements = append(block.Statements, member)
			}
			continue
		}

		memberMark := p.Mark(tok)
		leadingTrivia := p.dumpTrivia()

		member := p.parseSkinparamStatement(tok, false)
		if member != nil {
			p.stream.EmitCommentToks()
			member.SetMetadata(p.Span(memberMark), leadingTrivia, p.dumpTrivia())
			block.Statements = append(block.Statements, member)
		}
	}

	if rbraceTok, ok := p.stream.TryConsumeType(tokenizer.RBRACE); ok {
		block.RBrace = rbraceTok.Span.Start
		p.stream.EmitCommentToks()
		trailing := p.dumpTrivia()
		if len(trailing) > 0 {
			block.TrailingTrivia = append(block.TrailingTrivia, trailing...)
		}
	}
}
