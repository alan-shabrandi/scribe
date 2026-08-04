package config

import "testing"

func TestFallbackAPIKey(t *testing.T) {
	fakeOPENAI := "sk-proj-1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOP"
	fakeGEMINI := "AIzaSyFakeKeyTemplate_1234567890abcdefGhI"

	tests := []struct {
		name     string
		provider string
		openaiEnv string
		geminiEnv string
		want     string
	}{
		{
			name:     "empty provider",
			provider: "",
			want:     "",
		},
		{
			name:      "openai reads OPENAI_API_KEY",
			provider:  "openai",
			openaiEnv: fakeOPENAI,
			want:      fakeOPENAI,
		},
		{
			name:      "gemini reads GEMINI_API_KEY",
			provider:  "gemini",
			geminiEnv: fakeGEMINI,
			want:      fakeGEMINI,
		},
		{
			name:      "provider only reads its own env var",
			provider:  "gemini",
			openaiEnv: fakeOPENAI,
			want:      "",
		},
		{
			name:     "openai with unset env",
			provider: "openai",
			want:     "",
		},
		{
			name:      "claude has no fallback",
			provider:  "claude",
			openaiEnv: fakeOPENAI,
			geminiEnv: fakeGEMINI,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set both unconditionally so an ambient key in the developer's
			// environment can't leak into cases that expect "".
			t.Setenv("OPENAI_API_KEY", tt.openaiEnv)
			t.Setenv("GEMINI_API_KEY", tt.geminiEnv)

			got := fallbackAPIKey(tt.provider)
			if got != tt.want {
				t.Errorf("fallbackAPIKey(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}
