package parser

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
	"yur4uwe/pac/pkg/tokenizer/keyword"
)

// lineCursor is a view of a single statement's tokens. The terminator
// remains in the stream so a command cannot consume the next statement.
type lineCursor struct {
	tokens []tokenizer.Token
	end    tokenizer.Token
}

func (c *lineCursor) take() tokenizer.Token {
	if len(c.tokens) == 0 {
		return c.end
	}
	token := c.tokens[0]
	c.tokens = c.tokens[1:]
	return token
}

func (p *Parser) commandLine() lineCursor {
	tokens := p.syncToLineEnd()
	return lineCursor{tokens: tokens, end: p.stream.PeekTokenAt(0)}
}

type commandProblem struct {
	token   tokenizer.Token
	message string
}

func parseVisibilityLine(kind keyword.KeywordKind, target string, end tokenizer.Token) (*ast.VisibilityCommand, *commandProblem) {
	cmd := &ast.VisibilityCommand{Target: target}
	switch kind {
	case keyword.Hide:
		cmd.Kind = ast.VisibilityCMDHide
	case keyword.Show:
		cmd.Kind = ast.VisibilityCMDShow
	case keyword.Remove:
		cmd.Kind = ast.VisibilityCMDRemove
	case keyword.Restore:
		cmd.Kind = ast.VisibilityCMDRestore
	}
	if target == "" {
		return cmd, &commandProblem{end, "empty target for visibility command"}
	}
	return cmd, nil
}

func parseDirectionLine(first tokenizer.Token, line lineCursor) (*ast.DirectionCommand, *commandProblem) {
	cmd := &ast.DirectionCommand{}
	var to string
	switch first.Literal {
	case "left":
		to = "right"
	case "top":
		to = "bottom"
	default:
		return cmd, &commandProblem{first, "expected left or top for diagram direction"}
	}
	for _, want := range [...]string{"to", to, "direction"} {
		token := line.take()
		if token.Type != tokenizer.IDENTIFIER || token.Literal != want {
			return cmd, &commandProblem{token, "unexpected diagram direction modifier"}
		}
	}
	if first.Literal == "left" {
		cmd.Direction = ast.LeftToRightDirection
	} else {
		cmd.Direction = ast.TopToBottomDirection
	}
	if len(line.tokens) != 0 {
		return cmd, &commandProblem{line.tokens[0], "unexpected tokens after direction command"}
	}
	return cmd, nil
}

func parseSetLine(line lineCursor) (*ast.SetCommand, *commandProblem) {
	cmd := &ast.SetCommand{}
	key := line.take()
	if key.Type != tokenizer.IDENTIFIER {
		return cmd, &commandProblem{key, "expected key after set"}
	}
	cmd.Key = key.Literal
	if len(line.tokens) == 0 {
		return cmd, &commandProblem{line.end, "expected value after set key"}
	}
	var value strings.Builder
	for _, tok := range line.tokens {
		value.WriteString(tok.Literal)
	}
	cmd.Value = value.String()
	return cmd, nil
}
