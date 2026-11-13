package main

import (
	"errors"
	"fmt"
	"os"
)

type LspCommand struct {
	stdio bool
}

var _ Command = (*LspCommand)(nil)

func (c *LspCommand) Name() string     { return "lsp" }
func (c *LspCommand) Synopsis() string { return "Language server (planned)" }
func (c *LspCommand) Usage() string    { return "pac lsp [flags]" }

func (c *LspCommand) Run(args []string) error {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "-help") {
		fmt.Fprintln(os.Stderr, "Usage:", c.Usage())
		fmt.Fprintln(os.Stderr, "The language server is not implemented yet.")
		return nil
	}
	return errors.New("the language server is not implemented yet")
}
