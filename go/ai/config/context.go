package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"net/http"
	"net/url"
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

// fileContent reads a file or URL and fills the buffer with its content, along with a header.
//
//nolint:revive
func fileContent(buf *bytes.Buffer, path string, linum bool) error {
	reader, err := readContent(path)
	if err != nil {
		return fmt.Errorf("error reading content from %s: %w", path, err)
	}
	defer reader.Close()

	buf.WriteString("\n\n====> START OF " + path + " <====\n\n")

	if linum {
		scanner := bufio.NewScanner(reader)
		lineNumber := 1
		for scanner.Scan() {
			buf.WriteString(fmt.Sprintf("%d: %s\n", lineNumber, scanner.Text()))
			lineNumber++
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading content from %s line by line: %w", path, err)
		}
	} else {
		_, err := io.Copy(buf, reader)
		if err != nil {
			return fmt.Errorf("error writing content to buffer for %s: %w", path, err)
		}
	}

	buf.WriteString("\n\n====> END OF " + path + " <====")
	return nil
}

// readContent returns an io.Reader for the given path, which can be a local file or a URL.
// It also returns a cleanup function to be called when the reader is no longer needed.
func readContent(path string) (io.ReadCloser, error) {
	if isURL(path) {
		return downloadContent(path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// isURL checks if the given string is a valid URL.
func isURL(path string) bool {
	u, err := url.Parse(path)
	return err == nil && u.Scheme != "" && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

// downloadContent downloads content from the given URL and returns an io.Reader.
func downloadContent(fileURL string) (io.ReadCloser, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download URL %s: %w", fileURL, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download URL %s, status code: %d", fileURL, resp.StatusCode)
	}

	return readCloser{resp.Body, resp.Body.Close}, nil
}

type readCloser struct {
	io.Reader
	close func() error
}

func (rc readCloser) Close() error { return rc.close() }

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
