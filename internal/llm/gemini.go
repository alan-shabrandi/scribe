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
type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func buildSystemPrompt(diff string) string {
	return fmt.Sprintf(`You are an automated Git commit message generator.
Your job is to analyze the provided git diff and write a concise, professional commit message.

STRICT RULES:
1. Follow the Conventional Commits format: <type>(<scope>): <short summary>
   - Valid types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
   - Scope is optional, but encouraged if changes are isolated to a module/package.
2. Keep the first line (header) under 72 characters.
3. Write in the imperative mood, present tense (e.g., "add feature", NOT "added feature" or "adds feature").
4. Do NOT include markdown code blocks (e.g., no triple `+"````"+`).
5. Do NOT include introductory words, explanations, or concluding remarks (e.g., NO "Here is your commit message:").
6. Output raw text ONLY.

Git Diff:
%s`, diff)
}

func (c *Client) GenerateCommitMessage(diff string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is missing")
	}

	prompt := buildSystemPrompt(diff)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", c.APIKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("received empty response from Gemini API")
	}

	commitMessage := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	
	commitMessage = strings.TrimPrefix(commitMessage, "```")
	commitMessage = strings.TrimSuffix(commitMessage, "```")
	commitMessage = strings.TrimSpace(commitMessage)

	return commitMessage, nil
}
