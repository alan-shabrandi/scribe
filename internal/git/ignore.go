package git

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// scribeIgnoreFile is a repo-root-only, non-recursive subset of .gitignore syntax:
// plain names match the filename at any depth, path/glob patterns match the full
// relative path, and a leading "/" anchors a pattern to the repo root. "**" and "!"
// negation are not supported.
const scribeIgnoreFile = ".scribeignore"

var (
	customIgnoredOnce sync.Once
	customIgnored     []string
)

// loadCustomIgnores reads .scribeignore from the repo root once per process; missing file or no git repo just means no custom patterns.
func loadCustomIgnores() []string {
	customIgnoredOnce.Do(func() {
		root, err := repoRoot()
		if err != nil {
			return
		}

		f, err := os.Open(filepath.Join(root, scribeIgnoreFile))
		if err != nil {
			return
		}
		defer f.Close()

		customIgnored = parseIgnoreLines(f)
	})
	return customIgnored
}

// parseIgnoreLines reads .scribeignore-format lines, skipping blank lines and "#" comments.
func parseIgnoreLines(r io.Reader) []string {
	var patterns []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
