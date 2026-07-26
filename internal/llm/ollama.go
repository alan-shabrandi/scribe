package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OllamaClient struct {
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

func NewOllamaClient(model string) *OllamaClient {
	if model == "" {
		model = "llama3"
	}
	return &OllamaClient{
		Model:   model,
		BaseURL: "http://localhost:11434",
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (c *OllamaClient) GenerateCommitMessage(diff, style, ticketID string) (string, error) {
	prompt := BuildSystemPrompt(diff, style, ticketID)

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama (ensure Ollama is running locally): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	commitMessage := strings.TrimSpace(res.Response)
	commitMessage = strings.TrimPrefix(commitMessage, "```")
	commitMessage = strings.TrimSuffix(commitMessage, "```")
	return strings.TrimSpace(commitMessage), nil
}

func (c *OllamaClient) SummarizeFile(fileDiff string) (string, error) {
	prompt := buildFileSummaryPrompt(fileDiff)

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama (ensure Ollama is running locally): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	summary := strings.TrimSpace(res.Response)
	summary = strings.TrimPrefix(summary, "```")
	summary = strings.TrimSuffix(summary, "```")

	return strings.TrimSpace(summary), nil
}

func (c *OllamaClient) GenerateMultipleCommitMessages(diff, style, ticketID string) ([]string, error) {
	prompt := BuildMultiChoicePrompt(diff, style, ticketID)

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		c.BaseURL+"/api/generate",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to Ollama (ensure Ollama is running locally): %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"ollama API returned status %d: %s",
			resp.StatusCode,
			string(bodyBytes),
		)
	}

	var res ollamaResponse

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	rawText := strings.TrimSpace(res.Response)

	if rawText == "" {
		return nil, fmt.Errorf("empty response from Ollama")
	}

	lines := strings.Split(rawText, "\n")

	var candidates []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		trimmed = strings.TrimPrefix(trimmed, "- ")
		trimmed = strings.TrimPrefix(trimmed, "* ")

		if len(trimmed) > 2 {
			if trimmed[1] == '.' && trimmed[0] >= '0' && trimmed[0] <= '9' {
				trimmed = strings.TrimSpace(trimmed[2:])
			}
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
