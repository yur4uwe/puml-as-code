package main

type LspCommand struct {
	stdio bool
}

var _ Command = (*LspCommand)(nil)

func (c *LspCommand) Name() string     { return "lsp" }
func (c *LspCommand) Synopsis() string { return "Start the Language Server" }
func (c *LspCommand) Usage() string    { return "pac lsp [flags]" }

func (c *LspCommand) Run(args []string) error {
	return nil
}
