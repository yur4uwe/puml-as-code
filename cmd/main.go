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

func main() {
	inputf := flag.String("in", "", "Path to the input file")
	outdir := flag.String("out", "", "Path to the output dir")
	lang := flag.String("lang", "go", "Language to generate code in (default: go)")
	id := flag.String("id", "", "Diagram's index in the target file or literal ID (optional)")
	flag.Parse()

	if *inputf == "" {
		fmt.Fprintf(os.Stderr, "Please, provide non-empty file path with -in\n")
		os.Exit(1)
	}

	if *outdir == "" {
		fmt.Fprintf(os.Stderr, "Output directory not specified, please provide it with -out\n")
		os.Exit(1)
	}

	absPath, err := filepath.Abs(*inputf)
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

	p := parser.NewParser(dialect.Factory(*lang)).
		WithTargetID(*id)
	AST, err := p.Parse(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing PUML content: %v\n", err)
		os.Exit(1)
	}

	err = resolver.ResolveImports(AST, *inputf, *id, resolver.OSFileReader{}, *lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving imports: %v\n", err)
		os.Exit(1)
	}

	tbl, err := resolver.ResolveSymbols(AST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving symbols: %v\n", err)
		os.Exit(1)
	}

	codeGen, ok := generator.CodeGeneratorByLang(*lang)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unsupported language: %s\n", *lang)
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
		outPath := filepath.Join(*outdir, file.Path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory for %s: %v\n", outPath, err)
			os.Exit(1)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", outPath, err)
			os.Exit(1)
		}
	}
}
