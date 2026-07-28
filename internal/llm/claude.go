package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const claudeEndpoint = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

type ClaudeClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewClaudeClient(apiKey, model string) *ClaudeClient {
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &ClaudeClient{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeResponse struct {
	Content []claudeContentBlock `json:"content"`
}

func (c *ClaudeClient) headers() map[string]string {
	return map[string]string{
		"Content-Type":      "application/json",
		"x-api-key":         c.APIKey,
		"anthropic-version": anthropicVersion,
	}
}

func (c *ClaudeClient) GenerateCommitMessage(ctx context.Context, diff, style, ticketID string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("Claude API key is missing. Please set it in ~/.scribe.yaml or SCRIBE_API_KEY environment variable")
	}

	prompt := BuildSystemPrompt(diff, style, ticketID)
	reqBody := claudeRequest{
		Model:     c.Model,
		MaxTokens: 1000,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	var res claudeResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", claudeEndpoint, c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}

	if len(res.Content) == 0 {
		return "", fmt.Errorf("received empty response from Claude API")
	}

	return cleanResponseText(res.Content[0].Text), nil
}

func (c *ClaudeClient) SummarizeFile(ctx context.Context, fileDiff string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("Claude API key is missing")
	}

	prompt := buildFileSummaryPrompt(fileDiff)
	reqBody := claudeRequest{
		Model:     c.Model,
		MaxTokens: 1000,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	var res claudeResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", claudeEndpoint, c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}

	if len(res.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude API")
	}

	return cleanResponseText(res.Content[0].Text), nil
}

func (c *ClaudeClient) GenerateMultipleCommitMessages(ctx context.Context, diff, style, ticketID string) ([]string, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is missing")
	}

	prompt := BuildMultiChoicePrompt(diff, style, ticketID)
	reqBody := claudeRequest{
		Model:     c.Model,
		MaxTokens: 1000,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	var res claudeResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", claudeEndpoint, c.headers(), reqBody, &res); err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}

	if len(res.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude API")
	}

	return parseCandidates(res.Content[0].Text)
}
