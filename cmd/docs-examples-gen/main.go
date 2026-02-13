package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"yur4uwe/pac/pkg/tokenizer"
)

type diagram struct {
	name  string
	input string
}

type jsonExample struct {
	Name     string      `json:"n"`
	Input    string      `json:"i"`
	Expected []jsonToken `json:"e"`
}

type jsonToken struct {
	Type    string `json:"t"`
	Literal string `json:"l"`
}

func main() {
	docsDir := flag.String("docs-dir", "docs/class-diagram", "directory with class-diagram markdown docs")
	outDir := flag.String("out-dir", "pkg/tokenizer/examples", "output directory for JSON files")
	flag.Parse()

	// Collect diagrams grouped by source file
	diagramsByFile, err := collectDiagramsByFile(*docsDir)
	if err != nil {
		fail(err)
	}
	if len(diagramsByFile) == 0 {
		fail(fmt.Errorf("no <plantuml> blocks found under %s", *docsDir))
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fail(err)
	}

	// Generate a JSON file for each markdown file
	totalDiagrams := 0
	for mdFile, diagrams := range diagramsByFile {
		if len(diagrams) == 0 {
			continue
		}

		content, err := renderJSON(diagrams)
		if err != nil {
			fail(err)
		}

		// Convert markdown filename to JSON filename (e.g., "class-body.md" -> "class-body.json")
		jsonFilename := strings.TrimSuffix(mdFile, ".md") + ".json"
		outPath := filepath.Join(*outDir, jsonFilename)

		if err := os.WriteFile(outPath, content, 0o644); err != nil {
			fail(err)
		}

		fmt.Printf("Generated %s (%d diagrams)\n", outPath, len(diagrams))
		totalDiagrams += len(diagrams)
	}

	fmt.Printf("\nTotal: %d diagrams across %d files\n", totalDiagrams, len(diagramsByFile))
}

func collectDiagramsByFile(root string) (map[string][]diagram, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]diagram)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(root, entry.Name())
		diagrams, err := extractPlantumlBlocks(filePath)
		if err != nil {
			return nil, err
		}

		if len(diagrams) > 0 {
			result[entry.Name()] = diagrams
		}
	}

	return result, nil
}

func collectDiagrams(root string) ([]diagram, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, filepath.Join(root, entry.Name()))
		}
	}
	sort.Strings(files)

	var out []diagram
	for _, filePath := range files {
		diagrams, err := extractPlantumlBlocks(filePath)
		if err != nil {
			return nil, err
		}
		out = append(out, diagrams...)
	}
	return out, nil
}

func extractPlantumlBlocks(path string) ([]diagram, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var blocks []diagram
	var buf []string
	inBlock := false
	blockIndex := 0

	for _, line := range lines {
		switch {
		case strings.TrimSpace(line) == "<plantuml>":
			inBlock = true
			buf = buf[:0]
		case strings.TrimSpace(line) == "</plantuml>":
			if inBlock {
				blockIndex++
				name := fmt.Sprintf("%s#%02d", filepath.Base(path), blockIndex)
				blocks = append(blocks, diagram{name: name, input: strings.Join(buf, "\n")})
			}
			inBlock = false
		default:
			if inBlock {
				buf = append(buf, line)
			}
		}
	}

	return blocks, nil
}

func renderJSON(diagrams []diagram) ([]byte, error) {
	examples := make([]jsonExample, 0, len(diagrams))

	for _, d := range diagrams {
		tokens := collectTokens(d.input)
		jsonTokens := make([]jsonToken, len(tokens))

		for i, tok := range tokens {
			jsonTokens[i] = jsonToken{
				Type:    tok.Type.String(),
				Literal: tok.Literal,
			}
		}

		examples = append(examples, jsonExample{
			Name:     d.name,
			Input:    d.input,
			Expected: jsonTokens,
		})
	}

	// Minify JSON (no indentation)
	return json.MarshalIndent(examples, "", "  ")
}

func collectTokens(input string) []tokenizer.Token {
	lex := tokenizer.NewLexer(input)
	out := make([]tokenizer.Token, 0, 64)
	for tok := lex.NextToken(); tok.Type != tokenizer.EOF; tok = lex.NextToken() {
		out = append(out, tok)
	}
	return out
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
