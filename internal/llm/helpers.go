package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func sendJSONRequest(ctx context.Context, client *http.Client, method, url string, headers map[string]string, reqBody interface{}, respTarget interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if respTarget != nil {
		if err := json.NewDecoder(resp.Body).Decode(respTarget); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return nil
}

func cleanResponseText(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	return strings.TrimSpace(cleaned)
}

func parseCandidates(rawText string) ([]string, error) {
	rawText = strings.TrimSpace(rawText)
	if rawText == "" {
		return nil, fmt.Errorf("empty response received")
	}

	lines := strings.Split(rawText, "\n")
	var candidates []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		trimmed = strings.TrimPrefix(trimmed, "- ")
		trimmed = strings.TrimPrefix(trimmed, "* ")

		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && trimmed[1] == '.' {
			trimmed = strings.TrimSpace(trimmed[2:])
		}

		if trimmed != "" {
			candidates = append(candidates, trimmed)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("failed to parse candidate commit messages")
	}

	return candidates, nil
}
