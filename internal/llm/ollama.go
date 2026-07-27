package llm

import (
	"context"
	"fmt"
	"net/http"
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

func (c *OllamaClient) headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

func (c *OllamaClient) url() string {
	return c.BaseURL + "/api/generate"
}

func (c *OllamaClient) GenerateCommitMessage(ctx context.Context, diff, style, ticketID string) (string, error) {
	prompt := BuildSystemPrompt(diff, style, ticketID)

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	var res ollamaResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", c.url(), c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call Ollama API (ensure Ollama is running locally): %w", err)
	}

	return cleanResponseText(res.Response), nil
}

func (c *OllamaClient) SummarizeFile(ctx context.Context, fileDiff string) (string, error) {
	prompt := buildFileSummaryPrompt(fileDiff)

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	var res ollamaResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", c.url(), c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call Ollama API (ensure Ollama is running locally): %w", err)
	}

	return cleanResponseText(res.Response), nil
}

func (c *OllamaClient) GenerateMultipleCommitMessages(ctx context.Context, diff, style, ticketID string) ([]string, error) {
	prompt := BuildMultiChoicePrompt(diff, style, ticketID)

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	var res ollamaResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", c.url(), c.headers(), reqBody, &res); err != nil {
		return nil, fmt.Errorf("failed to call Ollama API (ensure Ollama is running locally): %w", err)
	}

	return parseCandidates(res.Response)
}
