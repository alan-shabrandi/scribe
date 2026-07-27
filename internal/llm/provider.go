package llm

import (
	"context"
	"fmt"
	"strings"
)

type Provider interface {
	GenerateCommitMessage(ctx context.Context, diff, style, ticketID string) (string, error)
	GenerateMultipleCommitMessages(ctx context.Context, diff, style, ticketID string) ([]string, error)
	SummarizeFile(ctx context.Context, fileDiff string) (string, error)
}

func NewProvider(providerType, apiKey, model string) (Provider, error) {
	normalized := strings.ToLower(strings.TrimSpace(providerType))

	switch normalized {
	case "gemini":
		return NewGeminiClient(apiKey, model), nil
	case "openai":
		return NewOpenAIClient(apiKey, model), nil
	case "ollama":
		return NewOllamaClient(model), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider '%s' (supported providers: gemini, openai, ollama)", providerType)
	}
}
