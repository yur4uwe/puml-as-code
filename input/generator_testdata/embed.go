// Package testdata provides test files for generator integration tests.
package testdata

import (
	"embed"
)

//go:embed *
var TestFiles embed.FS
