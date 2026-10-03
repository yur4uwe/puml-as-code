package main

import (
	"flag"
	"fmt"
	"os"
)

var commands = map[string]Command{
	"fmt": &FmtCommand{},
	"gen": &GenCommand{},
	"lsp": &LspCommand{},
}

func printHelp() {
	fmt.Fprintf(os.Stderr, "Usage: pac <command> [flags]\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	for name, cmd := range commands {
		fmt.Fprintf(os.Stderr, "  %s\t%s\n", name, cmd.Synopsis())
	}
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	flag.PrintDefaults()
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmdName := os.Args[1]
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmdName)
		os.Exit(1)
	}

	if err := cmd.Run(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "pac %s: %v\n", cmdName, err)
		os.Exit(1)
	}
}
