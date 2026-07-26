package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type CacheData struct {
	DiffHash   string   `json:"diff_hash"`
	Candidates []string `json:"candidates"`
}

func getCacheFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".scribe_cache.json"), nil
}

func ComputeHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func GetCachedCandidates(currentDiff string) ([]string, bool) {
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

	currentHash := ComputeHash(currentDiff)
	if cache.DiffHash == currentHash && len(cache.Candidates) > 0 {
		return cache.Candidates, true
	}

	return nil, false
}

func SaveCandidates(currentDiff string, candidates []string) error {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return err
	}

	cache := CacheData{
		DiffHash:   ComputeHash(currentDiff),
		Candidates: candidates,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0600)
}
