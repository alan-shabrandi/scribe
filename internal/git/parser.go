package git

import (
	"fmt"
	"path/filepath"
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

// matchesIgnorePattern matches a .scribeignore pattern against a diff header; globs use filepath.Match, plain text falls back to substring match.
func matchesIgnorePattern(header, pattern string) bool {
	if !strings.ContainsAny(pattern, "*?[") {
		return strings.Contains(header, pattern)
	}

	for field := range strings.FieldsSeq(header) {
		name := strings.TrimPrefix(strings.TrimPrefix(field, "a/"), "b/")
		if ok, err := filepath.Match(pattern, name); err == nil && ok {
			return true
		}
	}

	return false
}
