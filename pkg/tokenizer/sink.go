package tokenizer

type TokenSink interface {
	Receive(Token)
}

type TokenCollector struct {
	tokens []Token
}

func (c *TokenCollector) Receive(t Token) { c.tokens = append(c.tokens, t) }
func (c *TokenCollector) Len() int        { return len(c.tokens) }
