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

	wrapErr := func(err error) error {
		return fmt.Errorf("cannot format file: %w", err)
	}

	files := flagSet.Args()
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			return wrapErr(err)
		}

		if fileInfo.IsDir() {
			return fmt.Errorf("cannot format file: %s is a directory", file)
		}

		// files larger than 10MB are not supported
		if fileInfo.Size() >= 10*1024*1024 {
			return errors.New("cannot format file: file size is too large")
		} else if fileInfo.Size() == 0 {
			return nil // Nothing to format
		}

		fm := fileInfo.Mode()
		if !fm.IsRegular() {
			return fmt.Errorf("cannot format file: %s is not a regular file", file)
		}

		if c.write && (fm.Perm()&0o200 == 0) {
			return fmt.Errorf("cannot format file: %s is read-only", file)
		}

		fileReader, err := os.Open(file)
		if err != nil {
			return wrapErr(err)
		}

		// Peek first 512 bytes for a NULL byte
		buf := make([]byte, 512)
		n, _ := fileReader.Read(buf)
		if bytes.IndexByte(buf[:n], 0) != -1 || !utf8.Valid(buf[:n]) {
			return fmt.Errorf("cannot format file: %s appears to be a binary file", file)
		}

		// Reset the position
		fileReader.Seek(0, 0)

		// Read the whole file
		content, err := io.ReadAll(fileReader)
		if err != nil {
			return wrapErr(err)
		}

		formatted, err := formatter.Format(string(content))
		if err != nil {
			return fmt.Errorf("cannot format file: %w", err)
		}

		if c.write {
			tmpFReader, err := os.CreateTemp(filepath.Dir(file), ".pac-tmp-*")
			if err != nil {
				return wrapErr(err)
			}

			io.Copy(tmpFReader, strings.NewReader(formatted))
			err = tmpFReader.Sync()
			if err != nil {
				return wrapErr(err)
			}

			err = os.Chmod(tmpFReader.Name(), fm.Perm())
			if err != nil {
				return wrapErr(err)
			}
			err = os.Rename(tmpFReader.Name(), file)
			if err != nil {
				return wrapErr(err)
			}

			tmpFReader.Close()
		} else {
			fmt.Println(formatted)
		}

		fileReader.Close()
	}

	return nil
}
