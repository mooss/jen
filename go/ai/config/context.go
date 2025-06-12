package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
)

///////////////////
// Context class //

// Context handles inclusion of files and directories as context.
type Context struct {
	Files       []string
	Dirs        []string
	Above       bool
	LineNumbers bool
}

func (c *Context) Empty() bool { return len(c.Files) == 0 && len(c.Dirs) == 0 }

// Build generates the context string from included files and directories.
// Empty should be checked before calling this because the "additional context" header is always
// present,
func (c *Context) Build() (string, error) {
	var buf bytes.Buffer
	buf.WriteString("# Additional context (files)")

	for path, err := range c.allPaths {
		if err != nil {
			return "", c.wrap(err)
		}

		if err := fileContent(&buf, path, c.LineNumbers); err != nil {
			return "", c.wrap(err)
		}
	}

	return buf.String(), nil
}

// allPaths returns an iterator on all paths contained in the context.
func (c *Context) allPaths(yield func(string, error) bool) {
	for _, path := range c.Files {
		if !yield(path, nil) {
			return
		}
	}

	for _, dir := range c.Dirs {
		for path, err := range iterFiles(dir) {
			if !yield(path, err) {
				return
			}
		}
	}
}

func (*Context) wrap(err error) error {
	return fmt.Errorf("failed to build file context: %s", err)
}

///////////////////////
// Utility functions //

// fileContent reads a file and fills the buffer with its content, along with a header.
//
//nolint:revive
func fileContent(buf *bytes.Buffer, path string, linum bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", path, err)
	}
	defer file.Close()

	buf.WriteString("\n\n====> START OF " + path + " <====\n\n")

	if linum {
		scanner := bufio.NewScanner(file)
		lineNumber := 1
		for scanner.Scan() {
			buf.WriteString(fmt.Sprintf("%d: %s\n", lineNumber, scanner.Text()))
			lineNumber++
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading file %s line by line: %w", path, err)
		}
	} else {
		_, err := io.Copy(buf, file)
		if err != nil {
			return fmt.Errorf("error writing file content to buffer for %s: %w", path, err)
		}
	}

	buf.WriteString("\n\n====> END OF " + path + " <====")
	return nil
}

// iterFiles returns an iterator that yields all regular file paths under the given root directory.
//
//nolint:revive
func iterFiles(root string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				if !yield(path, err) {
					// The caller decides whether to stop.
					return err
				}

				return nil
			}

			if info.IsDir() {
				return nil
			}

			if !yield(path, nil) {
				return errors.New("IGNORED")
			}

			return nil
		})
	}
}
