package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"yur4uwe/pac/pkg/formatter"
	"yur4uwe/pac/pkg/resolver"
)

type FmtCommand struct {
	indentSize int
	useTabs    bool
	check      bool
	write      bool
	stdinPath  string
}

var errNeedsFormatting = errors.New("input needs formatting")

func formatExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errNeedsFormatting):
		return 1
	default:
		return 2
	}
}

var _ Command = (*FmtCommand)(nil)

func (c *FmtCommand) Name() string     { return "fmt" }
func (c *FmtCommand) Synopsis() string { return "Format PlantUML diagrams" }
func (c *FmtCommand) Usage() string    { return "pac fmt [flags] <file... | ->" }

func (c *FmtCommand) Run(args []string) error {
	return c.run(args, os.Stdin, os.Stdout, os.Stderr)
}

func (c *FmtCommand) run(args []string, in io.Reader, out, stderr io.Writer) error {
	flagSet := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	// main prints parse errors once; help is printed by Usage.
	flagSet.SetOutput(io.Discard)
	flagSet.Usage = func() {
		fmt.Fprintln(stderr, "Usage:", c.Usage())
		fmt.Fprintln(stderr, "Use - to read a diagram from stdin.")
		flagSet.SetOutput(stderr)
		flagSet.PrintDefaults()
		flagSet.SetOutput(io.Discard)
	}
	flagSet.IntVar(&c.indentSize, "indent-size", 2, "Spaces per indent")
	flagSet.BoolVar(&c.useTabs, "use-tabs", false, "Indent with tabs")
	flagSet.BoolVar(&c.write, "w", false, "Write result in-place")
	flagSet.BoolVar(&c.check, "check", false, "Check formatting without modifying")
	flagSet.StringVar(&c.stdinPath, "stdin-filepath", "", "Name used for diagnostics when reading stdin")

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	files := flagSet.Args()
	if len(files) == 0 {
		return fmt.Errorf("provide a file or - for stdin (see pac fmt -help)")
	}
	if c.indentSize < 0 {
		return fmt.Errorf("indent-size must be non-negative")
	}
	if c.write && c.check {
		return fmt.Errorf("-w and -check cannot be combined")
	}

	opts := formatter.FormatOptions{
		IndentSize:           c.indentSize,
		UseTabs:              c.useTabs,
		RequireSingleDiagram: true,
	}

	for _, file := range files {
		if file != "-" {
			continue
		}
		if len(files) != 1 {
			return fmt.Errorf("- must be the only input")
		}
		if c.write {
			return fmt.Errorf("-w cannot be used with stdin")
		}
		name := c.stdinPath
		if name == "" {
			name = "<stdin>"
		}
		content, err := readFormatInput(in, name)
		if err != nil {
			return fmt.Errorf("reading %s: %w", name, err)
		}
		return c.processInput(name, content, 0, out, opts)
	}
	if c.stdinPath != "" {
		return fmt.Errorf("-stdin-filepath requires - as the input")
	}
	var needsFormatting error
	for _, file := range files {
		if err := c.processFile(file, out, opts); errors.Is(err, errNeedsFormatting) {
			if needsFormatting == nil {
				needsFormatting = err
			}
		} else if err != nil {
			return err
		}
	}
	return needsFormatting
}

func (c *FmtCommand) processFile(file string, out io.Writer, opts formatter.FormatOptions) error {
	content, info, err := validateAndReadFile(file, c.write)
	if err != nil {
		return fmt.Errorf("cannot format file: %w", err)
	}
	return c.processInput(file, content, info.Mode().Perm(), out, opts)
}

func (c *FmtCommand) processInput(name, content string, perm os.FileMode, out io.Writer, opts formatter.FormatOptions) error {
	src := strings.TrimPrefix(content, "\ufeff")
	if src == "" {
		if !c.check && !c.write {
			_, err := io.WriteString(out, content)
			return err
		}
		return nil
	}

	// Format LF text and retain the original encoding and newline convention.
	normalized := strings.ReplaceAll(src, "\r\n", "\n")
	formatted, err := formatter.Format(normalized, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if !strings.HasSuffix(normalized, "\n") {
		formatted = strings.TrimSuffix(formatted, "\n")
	}
	if strings.Contains(src, "\r\n") && strings.Count(src, "\r\n") == strings.Count(src, "\n") {
		formatted = strings.ReplaceAll(formatted, "\n", "\r\n")
	}
	formatted = content[:len(content)-len(src)] + formatted

	if c.check {
		if formatted != content {
			return notFormattedError{name: name}
		}
		return nil
	}

	if c.write {
		if formatted == content {
			return nil
		}
		if err := writeAtomic(name, formatted, perm); err != nil {
			return fmt.Errorf("cannot format file: %w", err)
		}
		return nil
	}

	_, err = io.WriteString(out, formatted)
	return err
}

type notFormattedError struct{ name string }

func (e notFormattedError) Error() string { return e.name + " is not formatted" }
func (e notFormattedError) Unwrap() error { return errNeedsFormatting }

func validateAndReadFile(file string, isWrite bool) (string, os.FileInfo, error) {
	info, err := os.Stat(file)
	if err != nil {
		return "", nil, err
	}
	if info.IsDir() {
		return "", nil, fmt.Errorf("%s is a directory", file)
	}
	if !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("%s is not a regular file", file)
	}
	if info.Size() >= resolver.MaxFileSize {
		return "", nil, errors.New("file size is too large")
	}
	if info.Size() == 0 {
		return "", info, nil
	}
	if isWrite && (info.Mode().Perm()&0o200 == 0) {
		return "", nil, fmt.Errorf("%s is read-only", file)
	}

	f, err := os.Open(file)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	data, err := readFormatInput(f, file)
	if err != nil {
		return "", nil, err
	}
	return data, info, nil
}

func readFormatInput(in io.Reader, name string) (string, error) {
	data, err := io.ReadAll(io.LimitReader(in, resolver.MaxFileSize))
	if err != nil {
		return "", err
	}
	if len(data) >= resolver.MaxFileSize {
		return "", errors.New("input size is too large")
	}
	if bytes.IndexByte(data, 0) != -1 || !utf8.Valid(data) {
		return "", fmt.Errorf("%s appears to be a binary file", name)
	}
	return string(data), nil
}

func writeAtomic(file string, content string, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(file), ".pac-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, strings.NewReader(content)); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, file)
}
