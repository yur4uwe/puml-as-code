package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
)

var commands = map[string]Command{
	"fmt": &FmtCommand{},
	"gen": &GenCommand{},
	"lsp": &LspCommand{},
}

func printHelp() {
	fmt.Fprintf(os.Stderr, "Usage: pac <command> [flags]\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	for _, name := range slices.Sorted(maps.Keys(commands)) {
		fmt.Fprintf(os.Stderr, "  %s\t%s\n", name, commands[name].Synopsis())
	}
	fmt.Fprintln(os.Stderr, "\nRun 'pac <command> -help' for command options.")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmdName := os.Args[1]
	if cmdName == "-h" || cmdName == "--help" || cmdName == "-help" || cmdName == "help" {
		printHelp()
		return
	}
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmdName)
		os.Exit(1)
	}

	if err := cmd.Run(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "pac %s: %v\n", cmdName, err)
		if cmdName == "fmt" {
			os.Exit(formatExitCode(err))
		}
		os.Exit(1)
	}
}
