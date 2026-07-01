package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"yur4uwe/pac/pkg/generator"
	"yur4uwe/pac/pkg/parser"
)

func main() {
	inputf := flag.String("in", "", "Path to the input file")
	outdir := flag.String("out", "", "Path to the output dir (optional)")
	lang := flag.String("lang", "go", "Language to generate code in (default: go)")
	_ = flag.String("dialect", "go", "Dialect to use (default: go)")
	flag.Parse()

	if *inputf == "" {
		fmt.Printf("Please, provide non-empty file path")
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

	p := parser.Parser{}
	AST, err := p.Parse(string(content))
	if err != nil {
		fmt.Printf("Error parsing PUML content: %v\n", err)
		return
	}

	generator, err := generator.CodeGeneratorByLang(*lang)
	if err != nil {
		fmt.Printf("Error getting code generator: %v\n", err)
		return
	}

	parsedCode, err := generator.GenerateFromClassDiagram(AST)
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
		return
	}

	if *outdir == "" {
		fmt.Println(parsedCode)
		return
	}

	info, err := os.Stat(*outdir)
	switch {
	case os.IsNotExist(err):
		err = os.MkdirAll(*outdir, 0o755)
		if err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return
		}
	case !info.IsDir():
		fmt.Printf("Output path exists and is not a directory: %s\n", *outdir)
		return
	}

	err = os.MkdirAll(*outdir, 0o755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	fn := strings.TrimSuffix(filepath.Base(*inputf), filepath.Ext(*inputf)) + ".go"

	outFd, err := os.Create(filepath.Join(*outdir, fn))
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFd.Close()

	_, err = outFd.Write([]byte(parsedCode))
	if err != nil {
		fmt.Printf("Error writing to output file: %v\n", err)
		return
	}
}
