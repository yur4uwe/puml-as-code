package stream

import "yur4uwe/pac/pkg/tokenizer"

type TokenSink interface {
	Receive(tokenizer.Token)
}

type TokenCollector struct {
	tokens []tokenizer.Token
}

func (c *TokenCollector) Receive(t tokenizer.Token) { c.tokens = append(c.tokens, t) }
func (c *TokenCollector) Tokens() []tokenizer.Token { return c.tokens }
func (c *TokenCollector) Len() int                  { return len(c.tokens) }
