package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"
)

func (p *Parser) parseVisibilityCommand(tok tokenizer.Token) *ast.VisibilityCommand {
	cmd := &ast.VisibilityCommand{}
	switch keyword.Classify(tok.Literal) {
	case keyword.Hide:
		cmd.Kind = ast.VisibilityCMDHide
	case keyword.Show:
		cmd.Kind = ast.VisibilityCMDShow
	case keyword.Remove:
		cmd.Kind = ast.VisibilityCMDRemove
	case keyword.Restore:
		cmd.Kind = ast.VisibilityCMDRestore
	default:
		panic("visibility command: unreachable")
	}
	cmd.Target = p.stream.ReadUntilNewline()
	if len(cmd.Target) == 0 {
		nextTok := p.stream.PeekTokenAt(0)
		p.recordError(nextTok, cmd, "empty target for visibility command")
	}
	return cmd
}

func (p *Parser) parseDiagDirection(tok tokenizer.Token) *ast.DirectionCommand {
	var to string
	switch tok.Literal {
	case "left":
		to = "right"
	case "top":
		to = "bottom"
	}

	expectedSeq := []tokenizer.Token{
		{
			Type:    tokenizer.IDENTIFIER,
			Literal: "to",
		},
		{
			Type:    tokenizer.IDENTIFIER,
			Literal: to,
		},
		{
			Type:    tokenizer.IDENTIFIER,
			Literal: "direction",
		},
	}

	cmd := &ast.DirectionCommand{}

	for _, token := range expectedSeq {
		if _, ok := p.stream.TryConsume(token); ok {
			continue
		}

		nextToken := p.stream.PeekTokenAt(0)
		p.recordError(nextToken, cmd, "Unexpected diagram direction modifier")
		p.syncToLineEnd()
		return cmd
	}
	switch tok.Literal {
	case "left":
		cmd.Direction = ast.LeftToRightDirection
	case "top":
		cmd.Direction = ast.TopToBottomDirection
	}
	if !p.stream.AssertType(tokenizer.NEWLINE) {
		nextToken := p.stream.PeekTokenAt(0)
		p.recordError(nextToken, cmd, "Unexpected tokens after direction command")
		p.syncToLineEnd()
	}
	return cmd
}

func (p *Parser) parseScale(start tokenizer.Token) *ast.ScaleCommand {
	cmd := &ast.ScaleCommand{}
	_, cmd.IsMax = p.stream.TryConsume(amb(tokenizer.IDENTIFIER, "max"))
	tokens := extractScaleTokens(p.syncToLineEnd())
	span := p.Span(p.Mark(start))
	if message := readScaleOperands(cmd, tokens); message != "" {
		p.ast.RecordError(NewParserError(start, message), span, cmd)
	}
	return cmd
}

// The lexer reads 200x100 as one identifier. Split this form and the
// one-sided variants (200x and x100) into ordinary operand tokens.
func extractScaleTokens(tokens []tokenizer.Token) []tokenizer.Token {
	for i, tok := range tokens {
		if tok.Type != tokenizer.IDENTIFIER {
			continue
		}
		left, right, ok := strings.Cut(tok.Literal, "x")
		if !ok || left == "" && right == "" ||
			left != "" && !isScaleInteger(left) || right != "" && !isScaleInteger(right) {
			continue
		}
		normalized := make([]tokenizer.Token, 0, len(tokens)+2)
		normalized = append(normalized, tokens[:i]...)
		if left != "" {
			normalized = append(normalized, tokenizer.Token{Type: tokenizer.NUMBER, Literal: left, Span: tok.Span})
		}
		normalized = append(normalized, tokenizer.Token{Type: tokenizer.IDENTIFIER, Literal: "x", Span: tok.Span})
		if right != "" {
			normalized = append(normalized, tokenizer.Token{Type: tokenizer.NUMBER, Literal: right, Span: tok.Span})
		}
		return append(normalized, tokens[i+1:]...)
	}
	return tokens
}

func isScaleInteger(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func readScaleNumber(tokens []tokenizer.Token) (string, []tokenizer.Token) {
	if len(tokens) == 0 {
		return "", tokens
	}
	first := tokens[0]
	if first.Type == tokenizer.DOT && len(tokens) > 1 {
		second := tokens[1]
		if second.Type == tokenizer.NUMBER && isAdjacent(first, second) && isScaleInteger(second.Literal) {
			return "." + second.Literal, tokens[2:]
		}
	}
	if first.Type != tokenizer.NUMBER {
		return "", tokens
	}
	left, right, decimal := strings.Cut(first.Literal, ".")
	if !isScaleInteger(left) || decimal && right != "" && !isScaleInteger(right) {
		return "", tokens
	}
	return first.Literal, tokens[1:]
}

func readScaleOperands(cmd *ast.ScaleCommand, tokens []tokenizer.Token) string {
	cmd.Lhs, tokens = readScaleNumber(tokens)
	if cmd.Lhs == "" {
		return "expected a decimal number after scale"
	}

	if len(tokens) > 0 {
		tok := tokens[0]
		if tok.Type == tokenizer.ASTERISK || tok.Type == tokenizer.SLASH ||
			tok.Type == tokenizer.IDENTIFIER && tok.Literal == "x" {
			cmd.Sep = tok.Literal
			cmd.Rhs, tokens = readScaleNumber(tokens[1:])
			if !isScaleInteger(cmd.Lhs) || !isScaleInteger(cmd.Rhs) {
				return "expected decimal integers on both sides of the scale separator"
			}
		}
	}
	if len(tokens) > 0 && tokens[0].Type == tokenizer.IDENTIFIER {
		switch tokens[0].Literal {
		case "width", "height":
			cmd.Unit = tokens[0].Literal
			tokens = tokens[1:]
		}
	}
	switch {
	case len(tokens) > 0:
		return "unexpected tokens after scale operands"
	case cmd.Sep != "" && cmd.Unit != "":
		return "cannot specify a unit with a box or fraction"
	case cmd.IsMax && (!isScaleInteger(cmd.Lhs) || cmd.Sep == "/" || cmd.Sep == "" && cmd.Unit == ""):
		return "scale max requires an integer width, height, or box"
	}
	// Preserve the established AST representation of a bare full decimal.
	if cmd.Unit == "" && cmd.Sep == "" {
		if lhs, rhs, ok := strings.Cut(cmd.Lhs, "."); ok && lhs != "" && rhs != "" {
			cmd.Lhs, cmd.Sep, cmd.Rhs = lhs, ".", rhs
		}
	}
	return ""
}

func (p *Parser) parseSetCommand() *ast.SetCommand {
	keyTok := p.stream.Emit() // consume "separator"

	toks := p.stream.ConsumeUntilType(tokenizer.NEWLINE)
	var sb strings.Builder
	for _, t := range toks {
		sb.WriteString(t.Literal)
	}
	directiveVal := sb.String()
	if keyTok.Literal == "separator" {
		if directiveVal == "none" {
			p.stream.PackageSeparator = ""
		} else {
			p.stream.PackageSeparator = directiveVal
		}
	}

	return &ast.SetCommand{
		Key:   keyTok.Literal,
		Value: directiveVal,
	}
}

func parseLayoutAlignmentToken(tok tokenizer.Token, blockKind ast.TextBlockKind) (horiz string, vert string, isAlign bool) {
	switch strings.ToLower(tok.Literal) {
	case "left", "right", "center":
		return strings.ToLower(tok.Literal), "", true
	case "top", "bottom":
		isAlign = true
		if blockKind != ast.BlockLegend {
			isAlign = false
		}
		return "", strings.ToLower(tok.Literal), isAlign
	default:
		return "", "", false
	}
}

func (p *Parser) parseLayoutStatement(kwTok tokenizer.Token, prefixAlignment *tokenizer.Token) *ast.TextBlock {
	blockKind := mapKwTokToTextBlockKind(keyword.Classify(kwTok.Literal))

	block := &ast.TextBlock{
		Kind: blockKind,
	}

	// Process prefix alignment if provided (e.g. "left header", "center footer")
	if prefixAlignment != nil && blockKind != ast.BlockTitle {
		h, v, isAlign := parseLayoutAlignmentToken(*prefixAlignment, blockKind)
		if !isAlign {
			var err error
			if len(v) > 0 {
				err = NewParserError(*prefixAlignment, "Vertical alignment only supported for legend")
			} else {
				err = NewParserError(*prefixAlignment, "Invalid alignment")
			}
			p.ast.RecordError(err, *&prefixAlignment.Span, block)
		}
		block.HorizontalAlignment = h
		block.VerticalAlignment = v
	} else if prefixAlignment != nil && blockKind == ast.BlockTitle {
		err := NewParserError(*prefixAlignment, "Title alignment not supported")
		p.ast.RecordError(err, *&prefixAlignment.Span, block)
	}

	// Consume any trailing alignment modifiers on the opener line (e.g. "legend top left", "header center")
	for !p.stream.AssertAnyType(tokenizer.NEWLINE, tokenizer.EOF) {
		peekTok := p.stream.PeekTokenAt(0)
		h, v, isAlign := parseLayoutAlignmentToken(peekTok, blockKind)
		if !isAlign {
			break
		}
		if blockKind == ast.BlockTitle {
			break // For title, any tokens on the same line are title text
		}
		if h != "" {
			if block.HorizontalAlignment != "" {
				p.recordError(peekTok, block, "Horizontal alignment already set")
			} else {
				block.HorizontalAlignment = h
			}
			p.stream.Emit()
		} else if v != "" {
			if block.VerticalAlignment != "" {
				p.recordError(peekTok, block, "Vertical alignment already set")
			} else {
				block.VerticalAlignment = v
			}
			p.stream.Emit()
		}
	}

	lineContentToks := p.stream.ConsumeUntilType(tokenizer.NEWLINE)

	var body string
	if len(lineContentToks) == 0 {
		// --- MULTI-LINE BLOCK FORM ---
		var err error
		body, err = p.stream.ConsumeTextBlock("end", kwTok.Literal)
		if err != nil {
			err := ast.NewUnclosedBlockError(kwTok.Literal, "end "+kwTok.Literal)
			p.ast.RecordError(err, kwTok.Span, block)
			return block
		}
	} else {
		// --- SINGLE-LINE INLINE FORM ---
		body = p.stream.SliceInputEnclosingTokens(lineContentToks...)
	}

	block.Text = body
	if blockKind == ast.BlockTitle {
		p.ast.Title = block.Text
	}
	return block
}
