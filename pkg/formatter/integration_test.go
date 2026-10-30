package formatter

import (
	"strings"
	"testing"

	testdata "yur4uwe/pac/input/generator_testdata"
	"yur4uwe/pac/internal/testutil"
	"yur4uwe/pac/pkg/parser"
	"yur4uwe/pac/pkg/parser/dialect"

	"github.com/stretchr/testify/require"
)

func TestIntegration_RoundTripAndIdempotence(t *testing.T) {
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
			src := string(pumlBytes)

			p := parser.NewParser(dialect.LaxDialect{})
			tree1, err := p.Parse(src)
			require.NoError(t, err, "initial parse failed for %s", pumlFileName)

			formatted1, err := Format(src)
			require.NoError(t, err, "Format(src) failed for %s", pumlFileName)

			// Idempotence check: Format(Format(src)) == Format(src)
			formatted2, err := Format(formatted1)
			require.NoError(t, err, "Format(out1) failed for %s", pumlFileName)
			require.Equal(t, formatted1, formatted2, "Formatting is not idempotent for %s", pumlFileName)

			// Round-Trip AST Invariance: AST(Format(src)) == AST(src)
			p2 := parser.NewParser(dialect.LaxDialect{})
			tree2, err := p2.Parse(formatted1)
			require.NoError(t, err, "parse formatted output failed for %s", pumlFileName)

			testutil.SemanticEqual(t, tree1, tree2)
		})
	}
}

func TestIntegration_SyntheticComplexDiagram(t *testing.T) {
	complexInput := `@startuml
title Enterprise System Architecture
scale 1.5
left to right direction
hide empty members

skinparam backgroundColor #EEEBDC
skinparam class {
  BackgroundColor PaleGreen
  ArrowColor SeaGreen
}

package "Core Domain" as core <<Domain>> #lightblue {
  interface Repository<T> <<Generic>> {
    +FindByID(id string) (T, error)
    +Save(item T) error
  }
  class User {
    {static} +DefaultRole string
    +id string
    +name string
    -- Contact Information --
    +email string
    -phone string
    == Methods ==
    +GetDisplayName() string
  }
  enum Role {
    +Admin
    +User
    +Guest
  }
}

namespace External.Auth {
  class OAuthProvider {
    +ValidateToken(token string) bool
  }
}

note left of User : Core user aggregate root
note "Authentication bridge" as N1
note on link : TLS Encrypted

User "1" *-- "0..*" Role : has
User "1" --> "1" External.Auth.OAuthProvider : delegates auth

@enduml
`

	formatted1, err := Format(complexInput)
	require.NoError(t, err)

	formatted2, err := Format(formatted1)
	require.NoError(t, err)
	require.Equal(t, formatted1, formatted2, "Formatting complex diagram must be idempotent")

	p := parser.NewParser(dialect.LaxDialect{})
	tree1, err := p.Parse(complexInput)
	require.NoError(t, err)

	p2 := parser.NewParser(dialect.LaxDialect{})
	tree2, err := p2.Parse(formatted1)
	require.NoError(t, err)

	testutil.SemanticEqual(t, tree1, tree2)
}
