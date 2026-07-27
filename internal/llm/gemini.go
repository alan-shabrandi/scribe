package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const geminiBaseURLTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}
type GeminiClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &GeminiClient{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *GeminiClient) headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

func (c *GeminiClient) url() string {
	return fmt.Sprintf(geminiBaseURLTemplate, c.Model, c.APIKey)
}

func (c *GeminiClient) GenerateCommitMessage(ctx context.Context, diff, style, ticketID string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("API Key is missing. Please set it in ~/.scribe.yaml or SCRIBE_API_KEY environment variable")
	}

	prompt := BuildSystemPrompt(diff, style, ticketID)
	reqBody := GeminiRequest{
		Contents: []Content{
			{Parts: []Part{{Text: prompt}}},
		},
	}

	var res GeminiResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", c.url(), c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("received empty response from Gemini API")
	}

	return cleanResponseText(res.Candidates[0].Content.Parts[0].Text), nil
}

func (c *GeminiClient) SummarizeFile(ctx context.Context, fileDiff string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("API Key is missing")
	}

	prompt := buildFileSummaryPrompt(fileDiff)
	reqBody := GeminiRequest{
		Contents: []Content{
			{Parts: []Part{{Text: prompt}}},
		},
	}

	var res GeminiResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", c.url(), c.headers(), reqBody, &res); err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini API")
	}

	return cleanResponseText(res.Candidates[0].Content.Parts[0].Text), nil
}

func (c *GeminiClient) GenerateMultipleCommitMessages(ctx context.Context, diff, style, ticketID string) ([]string, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("API Key is missing")
	}

	prompt := BuildMultiChoicePrompt(diff, style, ticketID)
	reqBody := GeminiRequest{
		Contents: []Content{
			{Parts: []Part{{Text: prompt}}},
		},
	}

	var res GeminiResponse
	if err := sendJSONRequest(ctx, c.HTTPClient, "POST", c.url(), c.headers(), reqBody, &res); err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	return parseCandidates(res.Candidates[0].Content.Parts[0].Text)
}
