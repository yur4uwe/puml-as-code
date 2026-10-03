package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"yur4uwe/pac/pkg/generator"
	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"
)

type GenCommand struct {
	in   string
	outd string
	lang string
	id   string
}

var _ Command = (*GenCommand)(nil)

func (c *GenCommand) Name() string     { return "gen" }
func (c *GenCommand) Synopsis() string { return "Generate code from diagrams" }
func (c *GenCommand) Usage() string    { return "pac gen [flags] <file...>" }

func (c *GenCommand) Run(args []string) error {
	flagSet := flag.NewFlagSet(c.Name(), flag.ExitOnError)
	flagSet.StringVar(&c.in, "in", "", "Path to the input file")
	flagSet.StringVar(&c.outd, "outd", "", "Path to the output dir")
	flagSet.StringVar(&c.lang, "lang", "go", "Language to generate the code in (default: go)")
	flagSet.StringVar(&c.id, "id", "", "Diagram's index in the target file or literal ID (optional)")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if c.in == "" {
		fmt.Fprintf(os.Stderr, "Please, provide non-empty file path with -in\n")
		os.Exit(1)
	}

	if c.outd == "" {
		fmt.Fprintf(os.Stderr, "Output directory not specified, please provide it with -out\n")
		os.Exit(1)
	}

	absPath, err := filepath.Abs(c.in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	fd, err := os.Open(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer fd.Close()

	content, err := io.ReadAll(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	p := parser.NewParser(dialect.Factory(c.lang)).
		WithTargetID(c.id)
	AST, err := p.Parse(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing PUML content: %v\n", err)
		os.Exit(1)
	}

	err = resolver.ResolveImports(AST, c.in, c.id, resolver.OSFileReader{}, c.lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving imports: %v\n", err)
		os.Exit(1)
	}

	tbl, err := resolver.ResolveSymbols(AST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving symbols: %v\n", err)
		os.Exit(1)
	}

	codeGen, ok := generator.CodeGeneratorByLang(c.lang)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unsupported language: %s\n", c.lang)
		os.Exit(1)
	}

	if err := codeGen.SemanticPass(tbl); err != nil {
		fmt.Fprintf(os.Stderr, "Error in semantic validation: %v\n", err)
		os.Exit(1)
	}

	files, err := codeGen.GenerateFromClassDiagram(tbl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		outPath := filepath.Join(c.outd, file.Path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory for %s: %v\n", outPath, err)
			os.Exit(1)
		}
		if err := os.WriteFile(outPath, file.Content, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", outPath, err)
			os.Exit(1)
		}
	}

	return nil
}
