package helpers

import "fmt"

func SubstrRunes(s string, start, end int) (string, error) {
	r := []rune(s)
	if start < 0 || end < 0 {
		return "", fmt.Errorf("start/end must be >= 0")
	}
	if start > len(r) || end > len(r) {
		return "", fmt.Errorf("index out of range: len=%d start=%d end=%d", len(r), start, end)
	}
	if start >= end {
		return "", nil
	}
	return string(r[start:end]), nil
}
