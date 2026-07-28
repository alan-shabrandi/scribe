package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CacheData struct {
	DiffHash   string   `json:"diff_hash"`
	Candidates []string `json:"candidates"`
}

func getCacheFilePath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to locate .git directory: %w", err)
	}

	gitDir := strings.TrimSpace(string(output))

	return filepath.Join(gitDir, "scribe_cache.json"), nil
}

func ComputeHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func GetCachedCandidates(diffHash string) ([]string, bool) {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return nil, false
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var cache CacheData
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false
	}

	if cache.DiffHash == diffHash && len(cache.Candidates) > 0 {
		return cache.Candidates, true
	}

	return nil, false
}

func SaveCandidates(diffHash string, candidates []string) error {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return err
	}

	cache := CacheData{
		DiffHash:   diffHash,
		Candidates: candidates,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0600)
}
