package parser

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yur4uwe/pac/input/integration_testdata"
	"yur4uwe/pac/pkg/parser/dialect"

	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update-golden", false, "update .golden.json files on disk")

func TestDiagramParsing_Golden(t *testing.T) {
	entries, err := testdata.TestFiles.ReadDir(".")
	require.NoError(t, err, "failed to read embedded testdata directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".puml") {
			continue
		}

		pumlFileName := entry.Name()
		testName := strings.TrimSuffix(pumlFileName, ".puml")

		t.Run(testName, func(t *testing.T) {
			pumlBytes, err := testdata.TestFiles.ReadFile(pumlFileName)
			require.NoError(t, err, "failed to read embedded puml file %s", pumlFileName)

			p := &Parser{
				dialect: dialect.NewGoDialect(),
			}

			diagram, err := p.Parse(string(pumlBytes))
			require.NoError(t, err, "failed to parse puml diagram %s", pumlFileName)

			actualJSON, err := json.MarshalIndent(diagram, "", "  ")
			require.NoError(t, err, "failed to marshal AST diagram to JSON")

			goldenFileName := testName + ".golden.json"
			diskGoldenPath := filepath.Join("../../input/integration_testdata", goldenFileName)

			if *updateGolden {
				err := os.WriteFile(diskGoldenPath, actualJSON, 0644)
				require.NoError(t, err, "failed to update golden file %s", diskGoldenPath)
				t.Logf("Updated golden file: %s", diskGoldenPath)
				return
			}

			goldenBytes, err := testdata.TestFiles.ReadFile(goldenFileName)
			require.NoError(t, err, "golden file missing: %s (run 'go test ./pkg/parser -update-golden' to create it)", goldenFileName)

			require.JSONEq(t, string(goldenBytes), string(actualJSON), "AST JSON mismatch for %s", pumlFileName)
		})
	}
}
