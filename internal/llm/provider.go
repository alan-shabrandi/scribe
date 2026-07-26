package llm

import "fmt"

type LLMProvider interface {
	GenerateCommitMessage(diff, style, ticketID string) (string, error)
	GenerateMultipleCommitMessages(diff, style, ticketID string) ([]string, error)
	SummarizeFile(fileDiff string) (string, error)
}

func NewProvider(providerType, apiKey, model string) (LLMProvider, error) {
	switch providerType {
	case "gemini":
		return NewGeminiClient(apiKey, model), nil
	case "openai":
		return NewOpenAIClient(apiKey, model), nil
	case "ollama":
		return NewOllamaClient(model), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider '%s'", providerType)
	}
}
