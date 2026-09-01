package stdlib

import (
	"os/exec"
	"path"
	"strings"
	"sync"

	"yur4uwe/pac/pkg/resolver"
)

var (
	once        sync.Once
	stdPackages map[string]struct{}
	shortToPath map[string]string
)

var canonicalOverrides = map[string]string{
	"rand":     "crypto/rand",
	"template": "text/template",
	"pprof":    "runtime/pprof",
	"trace":    "runtime/trace",
}

// BuiltinTypes is the set of Go predefined identifiers that require no import and should not be emitted.
var BuiltinTypes = map[string]struct{}{
	"any":        {},
	"bool":       {},
	"byte":       {},
	"comparable": {},
	"complex64":  {},
	"complex128": {},
	"error":      {},
	"float32":    {},
	"float64":    {},
	"int":        {},
	"int8":       {},
	"int16":      {},
	"int32":      {},
	"int64":      {},
	"rune":       {},
	"string":     {},
	"uint":       {},
	"uint8":      {},
	"uint16":     {},
	"uint32":     {},
	"uint64":     {},
	"uintptr":    {},
}

// NOTE: Currently, initStdlib is executed synchronously on first call via sync.Once
// during the generator pass. In the future, we could pre-warm this in a background
// goroutine concurrently while lexing, parsing, and resolving are happening to hide
// the subprocess latency.
// Potential considerations to address if implemented:
// 1. Cancellation via context.Context if parsing fails early.
// 2. Ensure it only triggers when the target generation dialect is Go.
// 3. Non-blocking fallback if the subprocess hangs or fails.
func initStdlib() {
	stdPackages = make(map[string]struct{})
	shortToPath = make(map[string]string)

	out, err := exec.Command("go", "list", "std").Output()
	if err != nil {
		populateFallback()
		return
	}

	lines := strings.SplitSeq(strings.TrimSpace(string(out)), "\n")
	for line := range lines {
		pkgPath := strings.TrimSpace(line)
		if pkgPath == "" ||
			strings.HasPrefix(pkgPath, "internal/") ||
			strings.Contains(pkgPath, "/internal/") ||
			strings.HasPrefix(pkgPath, "vendor/") ||
			strings.Contains(pkgPath, "/vendor/") ||
			strings.HasPrefix(pkgPath, "cmd/") {
			continue
		}

		stdPackages[pkgPath] = struct{}{}
		pkgName := path.Base(pkgPath)

		if override, ok := canonicalOverrides[pkgName]; ok {
			shortToPath[pkgName] = override
		} else if _, exists := shortToPath[pkgName]; !exists {
			shortToPath[pkgName] = pkgPath
		}
	}
}

func populateFallback() {
	fallbackPkgs := []string{
		"archive/tar", "archive/zip", "bufio", "bytes", "cmp", "compress/gzip",
		"context", "crypto", "crypto/aes", "crypto/cipher", "crypto/rand", "crypto/sha256",
		"crypto/tls", "crypto/x509", "database/sql", "embed", "encoding/base64",
		"encoding/csv", "encoding/hex", "encoding/json", "encoding/xml", "errors",
		"flag", "fmt", "hash", "html/template", "image", "image/color", "image/png",
		"io", "io/fs", "iter", "log", "log/slog", "maps", "math", "math/big", "math/rand",
		"mime", "net", "net/http", "net/url", "os", "path", "path/filepath", "reflect",
		"regexp", "runtime", "slices", "sort", "strconv", "strings", "structs", "sync",
		"sync/atomic", "syscall", "testing", "text/template", "time", "unicode", "unicode/utf8",
	}

	for _, p := range fallbackPkgs {
		stdPackages[p] = struct{}{}
		pkgName := path.Base(p)
		if override, ok := canonicalOverrides[pkgName]; ok {
			shortToPath[pkgName] = override
		} else if _, exists := shortToPath[pkgName]; !exists {
			shortToPath[pkgName] = p
		}
	}
}

func ensureLoaded() {
	once.Do(initStdlib)
}

// IsBuiltin checks if the given type name is a Go builtin type.
func IsBuiltin(name string) bool {
	_, ok := BuiltinTypes[name]
	return ok
}

// LookupImportPath returns the canonical standard library import path for a package identifier or path.
func LookupImportPath(pkgNameOrPath string) (string, bool) {
	ensureLoaded()

	// Exact match on full import path
	if _, ok := stdPackages[pkgNameOrPath]; ok {
		return pkgNameOrPath, true
	}
	// Match on short package name (e.g. "http" -> "net/http", "json" -> "encoding/json")
	if path, ok := shortToPath[pkgNameOrPath]; ok {
		return path, true
	}
	return "", false
}

// IsStdlibPackage checks if the given name or path belongs to the Go standard library.
func IsStdlibPackage(pkgNameOrPath string) bool {
	ensureLoaded()

	if _, ok := stdPackages[pkgNameOrPath]; ok {
		return true
	}
	if _, ok := shortToPath[pkgNameOrPath]; ok {
		return true
	}
	return false
}

// IsStdlibEntity checks if an entity symbol represents a Go builtin or standard library type.
func IsStdlibEntity(ent *resolver.EntitySymbol) bool {
	if ent == nil {
		return false
	}

	name := ""
	if ent.AST != nil {
		name = ent.AST.Identifier
	}
	if name == "" {
		name = SimpleName(ent.FQN)
	}

	// 1. Built-in types without package (e.g. error, any, string)
	if len(ent.PackagePath) == 0 && IsBuiltin(name) {
		return true
	}

	// 2. Package path matches standard library (e.g. ["time"], ["net", "http"])
	if len(ent.PackagePath) > 0 {
		importPath := strings.Join(ent.PackagePath, "/")
		if IsStdlibPackage(importPath) || IsStdlibPackage(ent.PackagePath[len(ent.PackagePath)-1]) {
			return true
		}
	}

	// 3. FQN starts with standard library package name (e.g. "time.Time", "http.Request")
	if strings.Contains(ent.FQN, ".") {
		parts := strings.Split(ent.FQN, ".")
		// Check first segment (e.g. "time" from "time.Time")
		if IsStdlibPackage(parts[0]) {
			return true
		}
		// Check joined segments except last (e.g. "net/http" from "net.http.Request")
		joined := strings.Join(parts[:len(parts)-1], "/")
		if IsStdlibPackage(joined) {
			return true
		}
	}

	return false
}

func SimpleName(fqn string) string {
	if idx := strings.LastIndex(fqn, "."); idx != -1 {
		return fqn[idx+1:]
	}
	return fqn
}
