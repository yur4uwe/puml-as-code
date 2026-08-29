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
		fmt.Printf("Please, provide non-empty file path")
		return
	}

	if *outdir == "" {
		fmt.Printf("Output directory not specified, please, provide it\n")
		return
	}

	absPath, err := filepath.Abs(*inputf)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return
	}

	fd, err := os.Open(absPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	content, err := io.ReadAll(fd)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fd.Close()

	p := parser.NewParser(dialect.Factory(*lang)).
		WithTargetID(*id)
	AST, err := p.Parse(string(content))
	if err != nil {
		fmt.Printf("Error parsing PUML content: %v\n", err)
		return
	}

	err = resolver.ResolveImports(AST, *inputf, *id, resolver.OSFileReader{}, *lang)
	if err != nil {
		fmt.Printf("Error resolving imports: %v\n", err)
		return
	}

	tbl, err := resolver.ResolveSymbols(AST)
	if err != nil {
		fmt.Printf("Error resolving symbols: %v\n", err)
		return
	}

	generator, ok := generator.CodeGeneratorByLang(*lang)
	if !ok {
		fmt.Printf("Unsupported language: %s\n", *lang)
		return
	}
	_, err = generator.GenerateFromClassDiagram(tbl)
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
		return
	}
}
