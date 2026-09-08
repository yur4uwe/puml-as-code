package gogenerator

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	testdata "yur4uwe/pac/input/generator_testdata"
	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"

	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update-golden", false, "update .golden files on disk")

func serializeGoldenArchive(files []*GeneratedFile) []byte {
	if len(files) == 1 && files[0].Path == "types.go" {
		return files[0].Content
	}

	sorted := make([]*GeneratedFile, len(files))
	copy(sorted, files)
	slices.SortFunc(sorted, func(a, b *GeneratedFile) int {
		// Reverse comparison to use descending order
		return strings.Compare(b.Path, a.Path)
	})

	var sb strings.Builder
	for i, f := range sorted {
		if i > 0 {
			sb.WriteString("\n")
		}
		fmt.Fprintf(&sb, "==> %s <==\n", f.Path)
		sb.Write(f.Content)
	}
	return []byte(sb.String())
}

func parseGoldenArchive(data []byte) map[string]string {
	// 1. Normalize CRLF / CR to standard LF
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Fast path: single file archive
	if !strings.HasPrefix(content, "==> ") {
		return map[string]string{
			"types.go": content,
		}
	}

	result := make(map[string]string)
	var currentPath string
	var currentBuilder strings.Builder

	for line := range strings.Lines(content) {
		trimmed := strings.TrimRight(line, "\r\n")

		if strings.HasPrefix(trimmed, "==> ") && strings.HasSuffix(trimmed, " <==") {
			if currentPath != "" {
				// Strip the extra separating newline added between sections by serializeGoldenArchive
				result[currentPath] = strings.TrimSuffix(currentBuilder.String(), "\n\n") + "\n"
				currentBuilder.Reset()
			}
			currentPath = strings.TrimSuffix(strings.TrimPrefix(trimmed, "==> "), " <==")
			currentPath = filepath.ToSlash(currentPath)
			continue
		}

		if currentPath != "" {
			currentBuilder.WriteString(trimmed)
			currentBuilder.WriteString("\n")
		}
	}

	if currentPath != "" {
		result[currentPath] = currentBuilder.String()
	}

	return result
}

func TestGenerator_Golden(t *testing.T) {
	entries, err := testdata.TestFiles.ReadDir(".")
	require.NoError(t, err, "failed to read embedded generator_testdata directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".puml") {
			continue
		}

		pumlFileName := entry.Name()
		testName := strings.TrimSuffix(pumlFileName, ".puml")

		t.Run(testName, func(t *testing.T) {
			pumlBytes, err := testdata.TestFiles.ReadFile(pumlFileName)
			require.NoError(t, err, "failed to read embedded puml file %s", pumlFileName)

			p := parser.NewParser(dialect.NewGoDialect())
			diagram, err := p.Parse(string(pumlBytes))
			require.NoError(t, err, "failed to parse puml diagram %s", pumlFileName)

			tbl, err := resolver.ResolveSymbols(diagram)
			require.NoError(t, err, "failed to resolve symbols for %s", pumlFileName)

			gen := GoCodeGenerator{}
			err = gen.SemanticPass(tbl)
			require.NoError(t, err, "semantic pass failed for %s", pumlFileName)

			files, err := gen.GenerateFromClassDiagram(tbl)
			require.NoError(t, err, "code generation failed for %s", pumlFileName)
			require.NotEmpty(t, files, "no files generated for %s", pumlFileName)

			actualMap := make(map[string]string, len(files))
			for _, f := range files {
				actualMap[f.Path] = string(f.Content)
			}

			goldenFileName := testName + ".golden"
			diskGoldenPath := filepath.Join("../../../input/generator_testdata", goldenFileName)

			if *updateGolden {
				archiveData := serializeGoldenArchive(files)
				err := os.WriteFile(diskGoldenPath, archiveData, 0o644)
				require.NoError(t, err, "failed to update golden file %s", diskGoldenPath)
				t.Logf("Updated golden file: %s", diskGoldenPath)
				return
			}

			goldenBytes, err := testdata.TestFiles.ReadFile(goldenFileName)
			require.NoError(t, err, "golden file missing: %s (run 'go test ./pkg/generator/go -update-golden' to create it)", goldenFileName)

			expectedMap := parseGoldenArchive(goldenBytes)

			require.Equal(t, expectedMap, actualMap, "Generated Go output mismatch for %s", pumlFileName)
		})
	}
}
