package resolver

import (
	"errors"
	"fmt"
	"os"
)

// MaxFileSize defines the maximum allowed size for a single diagram file (10MB).
// If you encounter an actual diagram file > 10MB, please open an issue.
const MaxFileSize = 10 * 1024 * 1024

var ErrFileTooLarge = errors.New("file exceeds maximum allowed size (10MB)")

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type OSFileReader struct{}

func (OSFileReader) ReadFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("cannot read non-regular file: %s", path)
	}

	if info.Size() > MaxFileSize {
		// If you encounter an actual diagram file > 10MB, please open an issue.
		return nil, fmt.Errorf("%w: %s (%d bytes)", ErrFileTooLarge, path, info.Size())
	}

	return os.ReadFile(path)
}

var DefaultFS FileReader = OSFileReader{}

type MapFS map[string][]byte

func (m MapFS) ReadFile(path string) ([]byte, error) {
	data, ok := m[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}
