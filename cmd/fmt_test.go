package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	err := os.WriteFile(filePath, []byte(unformatted), 0644)
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
	err := os.WriteFile(filePath, []byte(unformatted), 0644)
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
	require.Equal(t, os.FileMode(0644), info.Mode().Perm())

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
		err := os.WriteFile(filePath, []byte(unformatted), 0644)
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
		err := os.WriteFile(filePath, []byte(formatted), 0644)
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
		err := os.WriteFile(filePath, []byte(unformatted), 0644)
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
		err := os.WriteFile(filePath, []byte(unformatted), 0644)
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

	err := os.WriteFile(f1, []byte("@startuml\nclass A {\n+x int\n}\n@enduml\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(f2, []byte("@startuml\nclass B {\n+y int\n}\n@enduml\n"), 0644)
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
		err := os.Mkdir(subDir, 0755)
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
		err := os.WriteFile(filePath, []byte("@startuml\n@enduml\n"), 0400)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{"-w", filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "is read-only")
	})

	t.Run("binary file with null byte", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "binary.bin")
		err := os.WriteFile(filePath, []byte("@startuml\x00class A\n@enduml"), 0644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "appears to be a binary file")
	})

	t.Run("binary file with invalid utf-8", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "invalid_utf8.puml")
		err := os.WriteFile(filePath, []byte{0xff, 0xfe, 0xfd}, 0644)
		require.NoError(t, err)

		cmd := &FmtCommand{}
		err = cmd.Run([]string{filePath})
		require.Error(t, err)
		require.Contains(t, err.Error(), "appears to be a binary file")
	})

	t.Run("empty file (0-byte)", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "empty.puml")
		err := os.WriteFile(filePath, []byte(""), 0644)
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
