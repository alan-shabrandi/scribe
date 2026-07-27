package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const openAIEndpoint = "https://api.openai.com/v1/chat/completions"

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

func (c *OpenAIClient) headers() map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.APIKey,
	}
}

func (c *OpenAIClient) GenerateCommitMessage(ctx context.Context, diff, style, ticketID string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is missing")
	}

	prompt := BuildSystemPrompt(diff, style, ticketID)
	reqBody := openAIRequest{
		Model: c.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	var res openAIResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", openAIEndpoint, c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	return cleanResponseText(res.Choices[0].Message.Content), nil
}

func (c *OpenAIClient) SummarizeFile(ctx context.Context, fileDiff string) (string, error) {
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

	var res openAIResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", openAIEndpoint, c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	return cleanResponseText(res.Choices[0].Message.Content), nil
}

func (c *OpenAIClient) GenerateMultipleCommitMessages(ctx context.Context, diff, style, ticketID string) ([]string, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is missing")
	}

	prompt := BuildMultiChoicePrompt(diff, style, ticketID)
	reqBody := openAIRequest{
		Model: c.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	var res openAIResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", openAIEndpoint, c.headers(), reqBody, &res); err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	return parseCandidates(res.Choices[0].Message.Content)
}
