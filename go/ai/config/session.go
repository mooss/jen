// This file handles jenai's session management.

package config

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mooss/jen/go/utils"
)

//////////////////////
// Session metadata //

type SessionMetadata struct {
	// Dir is the path to the aichat session dir.
	Dir string
	// Name is the name of the session.
	Name string
	// Requested is true when the session was explicitly requested.
	Requested bool
}

// prepare sets the proper the session name and dir and ensures the session is valid and ready to
// use.
func (ses *SessionMetadata) prepare() error {
	ses.Dir = sessionDir()
	ses.Requested = true

	switch ses.Name {
	case "":
		ses.Name = uniqueFilePrefix(ses.Dir, time.Now().Format("2006-01-02_15h04m"), ".yaml")
		ses.Requested = false
	case "/last":
		var err error
		ses.Name, err = mostRecentSession(ses.Dir)
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(sessionDir(), 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	return nil
}

// Path returns the path to the aichat session.
func (ses *SessionMetadata) Path() string {
	return filepath.Join(ses.Dir, ses.Name+".yaml")
}

//////////////////
// Session data //

// Conversation represents the entire session.
type Conversation struct {
	Model    string    `yaml:"model"`
	Messages []Message `yaml:"messages"`
}

// Message represents a single message in the session.
type Message struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

// Load loads a conversation from a YAML file.
func (ses *SessionMetadata) Load() (Conversation, error) {
	data, err := os.ReadFile(ses.Path())
	if err != nil {
		return Conversation{}, err
	}

	res, err := utils.FromYAML[Conversation](data)
	return utils.Wrapf(res, err, "failed to load session %s from YAML", ses.Path())
}

///////////////////////
// Utility functions //

// sessionDir return the path to the session directory.
func sessionDir() string {
	root := "."

	// If in a git repo, the session is at the root of the repo.
	gitRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		root = strings.TrimSpace(string(gitRoot))
	}

	return filepath.Join(root, ".jenai", "session")
}

// uniqueFilePrefix generates a unique file prefix based on a given directory, prefix, and suffix.
func uniqueFilePrefix(directory, prefix, suffix string) string {
	unique := prefix
	counter := 1

	for {
		filePath := filepath.Join(directory, unique+suffix)
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			return unique
		}

		unique = fmt.Sprintf("%s.%d", prefix, counter)
		counter++
	}
}

func mostRecentSession(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}

	var res fs.FileInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		res, err = mostRecent(res, entry)
		if err != nil {
			return "", err
		}
	}

	return strings.TrimSuffix(res.Name(), ".yaml"), nil
}

func mostRecent(ref fs.FileInfo, entry os.DirEntry) (fs.FileInfo, error) {
	candid, err := entry.Info()
	if err != nil || ref == nil {
		return candid, err
	}

	if candid.ModTime().Before(ref.ModTime()) {
		return candid, nil
	}

	return ref, nil
}
