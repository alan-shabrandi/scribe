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

type OpenAIClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

func (c *OpenAIClient) GenerateCommitMessage(diff, style string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is missing")
	}

	prompt := buildSystemPrompt(diff, style)

	reqBody := openAIRequest{
		Model: c.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	commitMessage := strings.TrimSpace(res.Choices[0].Message.Content)
	commitMessage = strings.TrimPrefix(commitMessage, "```")
	commitMessage = strings.TrimSuffix(commitMessage, "```")
	return strings.TrimSpace(commitMessage), nil
}

func (c *OpenAIClient) SummarizeFile(fileDiff string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is missing")
	}

	prompt := buildFileSummaryPrompt(fileDiff)

	reqBody := openAIRequest{
		Model: c.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	summary := strings.TrimSpace(res.Choices[0].Message.Content)
	summary = strings.TrimPrefix(summary, "```")
	summary = strings.TrimSuffix(summary, "```")

	return strings.TrimSpace(summary), nil
}
