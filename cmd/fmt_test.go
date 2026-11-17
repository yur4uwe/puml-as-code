package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"yur4uwe/pac/pkg/resolver"
)

func TestFmtStdinCLI(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "pac.exe")
	build := exec.Command("go", "build", "-o", bin, ".")
	output, err := build.CombinedOutput()
	require.NoError(t, err, "%s", output)
	unformatted := "@startuml\nclass User {\n+id string\n}\n@enduml\n"
	formatted := "@startuml\nclass User {\n  +id string\n}\n@enduml\n"
	for _, tt := range []struct {
		name, input, stdout, diagnostic string
		status                          int
		args                            []string
	}{
		{name: "format", input: unformatted, stdout: formatted, args: []string{"-"}},
		{name: "idempotent", input: formatted, stdout: formatted, args: []string{"-"}},
		{name: "empty", args: []string{"-"}},
		{name: "bom, crlf and no final newline", input: "\ufeff" + strings.ReplaceAll(strings.TrimSuffix(unformatted, "\n"), "\n", "\r\n"), stdout: "\ufeff" + strings.ReplaceAll(strings.TrimSuffix(formatted, "\n"), "\n", "\r\n"), args: []string{"-"}},
		{name: "valid utf8 across old sniff boundary", input: "@startuml\n'" + strings.Repeat("a", 500) + "é\n@enduml\n", stdout: "@startuml\n'" + strings.Repeat("a", 500) + "é\n@enduml\n", args: []string{"-"}},
		{name: "syntax error", input: "@startuml\nclass User {\n@enduml\n", status: 2, diagnostic: "models/user.puml", args: []string{"-stdin-filepath", "models/user.puml", "-"}},
		{name: "invalid utf8 after first 512 bytes", input: formatted + strings.Repeat(" ", 512) + "\xff", status: 2, diagnostic: "binary file", args: []string{"-"}},
		{name: "extra diagram is not silently discarded", input: formatted + formatted, status: 2, diagnostic: "expected one diagram", args: []string{"-"}},
		{name: "check clean", input: formatted, args: []string{"-check", "-"}},
		{name: "check dirty", input: unformatted, status: 1, diagnostic: "is not formatted", args: []string{"-check", "-"}},
		{name: "check file error", status: 2, diagnostic: "cannot format file", args: []string{"-check", filepath.Join(t.TempDir(), "missing.puml")}},
		{name: "size limit", input: strings.Repeat(" ", resolver.MaxFileSize), status: 2, diagnostic: "input size is too large", args: []string{"-"}},
		{name: "mixed inputs", status: 2, diagnostic: "only input", args: []string{"file.puml", "-"}},
		{name: "in-place stdin", status: 2, diagnostic: "-w cannot be used with stdin", args: []string{"-w", "-"}},
		{name: "no implicit stdin", status: 2, diagnostic: "provide a file or -"},
		{name: "help", diagnostic: "-stdin-filepath", args: []string{"-help"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(bin, append([]string{"fmt"}, tt.args...)...)
			cmd.Stdin = strings.NewReader(tt.input) // os/exec connects this reader through a pipe.
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			stdout, err := cmd.Output()
			if tt.status == 0 {
				require.NoError(t, err, "%s", &stderr)
			} else {
				var exitErr *exec.ExitError
				require.ErrorAs(t, err, &exitErr)
				require.Equal(t, tt.status, exitErr.ExitCode(), "%s", &stderr)
			}
			require.Equal(t, tt.stdout, string(stdout))
			if tt.diagnostic != "" {
				require.Contains(t, stderr.String(), tt.diagnostic)
			} else {
				require.Empty(t, stderr.String())
			}
		})
	}
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	runErr := fn()
	w.Close()
	os.Stdout = oldStdout
	out := <-outChan
	return out, runErr
}

func TestFmtCommand_StdoutMode(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "diagram.puml")
	unformatted := "@startuml\nclass User {\n+id string\n}\n@enduml\n"
	err := os.WriteFile(filePath, []byte(unformatted), 0o644)
	require.NoError(t, err)

	cmd := &FmtCommand{}
	out, err := captureStdout(t, func() error {
		return cmd.Run([]string{filePath})
	})
	require.NoError(t, err)

	expected := "@startuml\nclass User {\n  +id string\n}\n@enduml\n"
	require.Equal(t, expected, out)

	// File on disk must remain unchanged
	diskContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, unformatted, string(diskContent))
}

func TestFmtCommand_InPlaceWriteMode(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "diagram.puml")
	unformatted := "@startuml\nclass User {\n+id string\n}\n@enduml\n"
	err := os.WriteFile(filePath, []byte(unformatted), 0o644)
	require.NoError(t, err)

	cmd := &FmtCommand{}
	err = cmd.Run([]string{"-w", filePath})
	require.NoError(t, err)

	expected := "@startuml\nclass User {\n  +id string\n}\n@enduml\n"
	diskContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, expected, string(diskContent))

	// File permissions must be 0644
	info, err := os.Stat(filePath)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		require.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	}

	// No leftover temporary files in directory
	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.False(t, strings.HasPrefix(entries[0].Name(), ".pac-tmp-"))
}

func TestFmtCommand_CheckMode(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("unformatted file returns error", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "unformatted.puml")
		unformatted := "@startuml\nclass User {\n+id string\n}\n@enduml\n"
		err := os.WriteFile(filePath, []byte(unformatted), 0o644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{"-check", filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not formatted")

		// File must remain unchanged
		diskContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, unformatted, string(diskContent))
	})

	t.Run("formatted file returns nil", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "formatted.puml")
		formatted := "@startuml\nclass User {\n  +id string\n}\n@enduml\n"
		err := os.WriteFile(filePath, []byte(formatted), 0o644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{"-check", filePath})
		require.NoError(t, err)
	})
}

func TestFmtCommand_FlagPropagation(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("custom indent-size", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "indent.puml")
		unformatted := "@startuml\nclass User {\n+id string\n}\n@enduml\n"
		err := os.WriteFile(filePath, []byte(unformatted), 0o644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		out, err := captureStdout(t, func() error {
			return cmd.Run([]string{"--indent-size", "4", filePath})
		})
		require.NoError(t, err)

		expected := "@startuml\nclass User {\n    +id string\n}\n@enduml\n"
		require.Equal(t, expected, out)
	})

	t.Run("use-tabs", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "tabs.puml")
		unformatted := "@startuml\nclass User {\n+id string\n}\n@enduml\n"
		err := os.WriteFile(filePath, []byte(unformatted), 0o644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		out, err := captureStdout(t, func() error {
			return cmd.Run([]string{"--use-tabs", filePath})
		})
		require.NoError(t, err)

		expected := "@startuml\nclass User {\n\t+id string\n}\n@enduml\n"
		require.Equal(t, expected, out)
	})
}

func TestFmtCommand_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()
	f1 := filepath.Join(tempDir, "f1.puml")
	f2 := filepath.Join(tempDir, "f2.puml")

	err := os.WriteFile(f1, []byte("@startuml\nclass A {\n+x int\n}\n@enduml\n"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(f2, []byte("@startuml\nclass B {\n+y int\n}\n@enduml\n"), 0o644)
	require.NoError(t, err)

	cmd := &FmtCommand{}
	err = cmd.Run([]string{"-w", f1, f2})
	require.NoError(t, err)

	content1, err := os.ReadFile(f1)
	require.NoError(t, err)
	require.Equal(t, "@startuml\nclass A {\n  +x int\n}\n@enduml\n", string(content1))

	content2, err := os.ReadFile(f2)
	require.NoError(t, err)
	require.Equal(t, "@startuml\nclass B {\n  +y int\n}\n@enduml\n", string(content2))
}

func TestFmtCommand_Safeguards(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("directory input", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "subdir")
		err := os.Mkdir(subDir, 0o755)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{subDir})
		require.Error(t, err)
		require.Contains(t, err.Error(), "is a directory")
	})

	t.Run("non-existent file", func(t *testing.T) {
		cmd := &FmtCommand{}
		err := cmd.Run([]string{filepath.Join(tempDir, "nonexistent.puml")})
		require.Error(t, err)
	})

	t.Run("read-only file with write flag", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "readonly.puml")
		err := os.WriteFile(filePath, []byte("@startuml\n@enduml\n"), 0o400)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{"-w", filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "is read-only")
	})

	t.Run("binary file with null byte", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "binary.bin")
		err := os.WriteFile(filePath, []byte("@startuml\x00class A\n@enduml"), 0o644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "appears to be a binary file")
	})

	t.Run("binary file with invalid utf-8", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "invalid_utf8.puml")
		err := os.WriteFile(filePath, []byte{0xff, 0xfe, 0xfd}, 0o644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "appears to be a binary file")
	})

	t.Run("empty file (0-byte)", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "empty.puml")
		err := os.WriteFile(filePath, []byte(""), 0o644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{filePath})
		require.NoError(t, err)
	})

	t.Run("file larger than 10MB", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "large.puml")
		f, err := os.Create(filePath)
		require.NoError(t, err)
		err = f.Truncate(11 * 1024 * 1024)
		require.NoError(t, err)
		f.Close()

		cmd := &FmtCommand{}
		err = cmd.Run([]string{filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "file size is too large")
	})
}
