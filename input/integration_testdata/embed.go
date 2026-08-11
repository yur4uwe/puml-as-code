// Package testdata provides test files for integration tests.
package testdata

import (
	"embed"
)

//go:embed *
var TestFiles embed.FS
