package git

import (
	"fmt"
	"path"
	"strings"
)

var ignoredFiles = []string{
	"go.sum",
	"package-lock.json",
	"yarn.lock",
	"pnpm-lock.yaml",
	"composer.lock",
	"Cargo.lock",
}

const maxDiffLength = 8000

func FilterDiff(rawDiff string) (string, error) {
	if strings.TrimSpace(rawDiff) == "" {
		return "", fmt.Errorf("diff is empty")
	}

	filteredChunks := SplitDiffByFile(rawDiff)

	if len(filteredChunks) == 0 {
		return "", fmt.Errorf("all changes were filtered out (ignored lockfiles or auto-generated files)")
	}

	result := strings.Join(filteredChunks, "")

	runes := []rune(result)
	if len(runes) > maxDiffLength {
		result = string(runes[:maxDiffLength]) + "\n\n... [Diff truncated to save tokens]"
	}

	return result, nil
}

func SplitDiffByFile(rawDiff string) []string {
	rawChunks := strings.Split(rawDiff, "diff --git ")
	var fileChunks []string

	for _, chunk := range rawChunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		lines := strings.SplitN(chunk, "\n", 2)
		header := lines[0]

		if isIgnoredFile(header) {
			continue
		}

		fileChunks = append(fileChunks, "diff --git "+chunk)
	}

	return fileChunks
}

func isIgnoredFile(header string) bool {
	for _, ignored := range ignoredFiles {
		if strings.Contains(header, ignored) {
			return true
		}
	}

	for _, pattern := range loadCustomIgnores() {
		if matchesIgnorePattern(header, pattern) {
			return true
		}
	}

	return false
}

// matchesIgnorePattern matches a .scribeignore pattern against a diff header, gitignore-style:
// a pattern with no "/" matches the filename at any depth; a pattern with "/" (or a leading "/")
// matches the full repo-relative path. Globs use path.Match (diff headers always use "/", so
// path.Match is used over filepath.Match to stay OS-independent); plain text falls back to substring match.
func matchesIgnorePattern(header, pattern string) bool {
	if !strings.ContainsAny(pattern, "*?[") {
		return strings.Contains(header, pattern)
	}

	anchored := strings.HasPrefix(pattern, "/")
	pattern = strings.TrimPrefix(pattern, "/")
	hasSlash := anchored || strings.Contains(pattern, "/")

	for _, name := range diffPaths(header) {
		if !hasSlash {
			name = path.Base(name)
		}
		if ok, err := path.Match(pattern, name); err == nil && ok {
			return true
		}
	}

	return false
}

// diffPaths extracts the "a/" and "b/" paths from a "diff --git a/X b/X" header.
func diffPaths(header string) []string {
	before, after, ok := strings.Cut(header, " b/")
	if !ok {
		return nil
	}
	return []string{strings.TrimPrefix(before, "a/"), after}
}
