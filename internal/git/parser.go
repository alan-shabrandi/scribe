package git

import (
	"fmt"
	"strings"
)

var ignoredFiles = map[string]bool{
	"go.sum":            true,
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"composer.lock":     true,
	"Cargo.lock":        true,
}

const maxDiffLength = 8000

func FilterDiff(rawDiff string) (string, error) {
	if strings.TrimSpace(rawDiff) == "" {
		return "", fmt.Errorf("diff is empty")
	}
	chunks := strings.Split(rawDiff, "diff --git ")
	var filteredChunks []string

	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		lines := strings.SplitN(chunk, "\n", 2)
		header := lines[0]

		if isIgnoredFile(header) {
			continue
		}

		filteredChunks = append(filteredChunks, "diff --git "+chunk)
	}
	if len(filteredChunks) == 0 {
		return "", fmt.Errorf("all changes were filtered out (ignored lockfiles or auto-generated files)")
	}

	result := strings.Join(filteredChunks, "")

	if len(result) > maxDiffLength {
		result = result[:maxDiffLength] + "\n\n... [Diff truncated to save tokens]"
	}

	return result, nil
}

func isIgnoredFile(header string) bool {
	for ignored := range ignoredFiles {
		if strings.Contains(header, ignored) {
			return true
		}
	}
	return false
}