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
)

type FmtCommand struct {
	indentSize int
	useTabs    bool
	check      bool
	write      bool
}

var _ Command = (*FmtCommand)(nil)

func (c *FmtCommand) Name() string     { return "fmt" }
func (c *FmtCommand) Synopsis() string { return "Format PlantUML diagrams" }
func (c *FmtCommand) Usage() string    { return "pac fmt [flags] <file...>" }

func (c *FmtCommand) Run(args []string) error {
	flagSet := flag.NewFlagSet(c.Name(), flag.ExitOnError)
	flagSet.IntVar(&c.indentSize, "indent-size", 2, "Spaces per indent")
	flagSet.BoolVar(&c.useTabs, "use-tabs", false, "Indent with tabs")
	flagSet.BoolVar(&c.write, "w", false, "Write result in-place")
	flagSet.BoolVar(&c.check, "check", false, "Check formatting without modifying")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	opts := formatter.FormatOptions{
		IndentSize: c.indentSize,
		UseTabs:    c.useTabs,
	}

	for _, file := range flagSet.Args() {
		if err := c.processFile(file, opts); err != nil {
			return err
		}
	}
	return nil
}

func (c *FmtCommand) processFile(file string, opts formatter.FormatOptions) error {
	content, info, err := validateAndReadFile(file, c.write)
	if err != nil {
		return fmt.Errorf("cannot format file: %w", err)
	}
	if info.Size() == 0 {
		return nil
	}

	formatted, err := formatter.Format(content, opts)
	if err != nil {
		return fmt.Errorf("cannot format file: %w", err)
	}

	if c.check {
		if formatted != content {
			return fmt.Errorf("file %s is not formatted", file)
		}
		return nil
	}

	if c.write {
		if err := writeAtomic(file, formatted, info.Mode().Perm()); err != nil {
			return fmt.Errorf("cannot format file: %w", err)
		}
		return nil
	}

	fmt.Print(formatted)
	return nil
}

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
	if info.Size() >= 10*1024*1024 {
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

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if bytes.IndexByte(buf[:n], 0) != -1 || !utf8.Valid(buf[:n]) {
		return "", nil, fmt.Errorf("%s appears to be a binary file", file)
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return "", nil, err
	}
	return string(data), info, nil
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
