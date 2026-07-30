package git

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

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

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			customIgnored = append(customIgnored, line)
		}
	})
	return customIgnored
}

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
