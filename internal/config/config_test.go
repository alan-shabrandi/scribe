package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestFallbackAPIKey(t *testing.T) {
	fakeOPENAI := "sk-proj-1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOP"
	fakeGEMINI := "AIzaSyFakeKeyTemplate_1234567890abcdefGhI"

	tests := []struct {
		name      string
		provider  string
		openaiEnv string
		geminiEnv string
		want      string
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

// setupConfigEnv points LoadConfig at an isolated home directory and clears every
// environment variable it consults, so no case can inherit developer state. It also
// resets viper, whose package-level globals otherwise leak config paths and defaults
// between calls. Returns the temp home directory.
func setupConfigEnv(t *testing.T) string {
	t.Helper()

	viper.Reset()
	t.Cleanup(viper.Reset)

	dir := t.TempDir()
	// os.UserHomeDir reads USERPROFILE on Windows and HOME elsewhere.
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)

	for _, key := range []string{
		"SCRIBE_PROVIDER", "SCRIBE_API_KEY", "SCRIBE_MODEL", "SCRIBE_STYLE",
		"OPENAI_API_KEY", "GEMINI_API_KEY",
	} {
		t.Setenv(key, "")
	}

	return dir
}

// writeConfigFile checks that file writes can still happen correctly
func writeConfigFile(t *testing.T, dir, contents string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, ".scribe.yaml"), []byte(contents), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

// TestLoadConfig tests the default create when no config file is found
func TestLoadConfig_DefaultsWhenNoConfigFile(t *testing.T) {
	setupConfigEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	want := Config{Provider: "gemini", Model: "gemini-1.5-flash", Style: "conventional"}
	if *cfg != want {
		t.Errorf("LoadConfig() = %+v, want %+v", *cfg, want)
	}
}

func TestLoadConfig_ReadsConfigFile(t *testing.T) {
	dir := setupConfigEnv(t)
	writeConfigFile(t, dir, `
provider: openai
api_key: sk-from-file
model: gpt-4o
style: descriptive
`)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	want := Config{Provider: "openai", APIKey: "sk-from-file", Model: "gpt-4o", Style: "descriptive"}
	if *cfg != want {
		t.Errorf("LoadConfig() = %+v, want %+v", *cfg, want)
	}
}

func TestLoadConfig_PartialConfigFileKeepsDefaults(t *testing.T) {
	dir := setupConfigEnv(t)
	writeConfigFile(t, dir, "provider: ollama\n")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if cfg.Provider != "ollama" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "ollama")
	}
	if cfg.Model != "gemini-1.5-flash" {
		t.Errorf("Model = %q, want default %q", cfg.Model, "gemini-1.5-flash")
	}
	if cfg.Style != "conventional" {
		t.Errorf("Style = %q, want default %q", cfg.Style, "conventional")
	}
}

func TestLoadConfig_EnvOverridesConfigFile(t *testing.T) {
	dir := setupConfigEnv(t)
	writeConfigFile(t, dir, "provider: openai\nmodel: gpt-4o\nstyle: descriptive\n")

	t.Setenv("SCRIBE_PROVIDER", "ollama")
	t.Setenv("SCRIBE_MODEL", "llama3")
	t.Setenv("SCRIBE_STYLE", "conventional")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if cfg.Provider != "ollama" {
		t.Errorf("Provider = %q, want %q from SCRIBE_PROVIDER", cfg.Provider, "ollama")
	}
	if cfg.Model != "llama3" {
		t.Errorf("Model = %q, want %q from SCRIBE_MODEL", cfg.Model, "llama3")
	}
	if cfg.Style != "conventional" {
		t.Errorf("Style = %q, want %q from SCRIBE_STYLE", cfg.Style, "conventional")
	}
}

func TestLoadConfig_MalformedConfigFileReturnsError(t *testing.T) {
	dir := setupConfigEnv(t)
	writeConfigFile(t, dir, "provider: [unclosed\n")

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() returned nil error for malformed config file, want error")
	}
}

func TestLoadConfig_FallsBackToProviderEnvKey(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		envVar   string
		envValue string
		want     string
	}{
		{
			name:     "gemini falls back to GEMINI_API_KEY",
			provider: "gemini",
			envVar:   "GEMINI_API_KEY",
			envValue: "gemini-fallback-key",
			want:     "gemini-fallback-key",
		},
		{
			name:     "openai falls back to OPENAI_API_KEY",
			provider: "openai",
			envVar:   "OPENAI_API_KEY",
			envValue: "openai-fallback-key",
			want:     "openai-fallback-key",
		},
		{
			name:     "claude has no fallback",
			provider: "claude",
			envVar:   "ANTHROPIC_API_KEY",
			envValue: "anthropic-key",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupConfigEnv(t)
			writeConfigFile(t, dir, "provider: "+tt.provider+"\n")
			t.Setenv(tt.envVar, tt.envValue)

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() returned error: %v", err)
			}

			if cfg.APIKey != tt.want {
				t.Errorf("APIKey = %q, want %q", cfg.APIKey, tt.want)
			}
		})
	}
}

func TestLoadConfig_ConfigFileKeyBeatsEnvFallback(t *testing.T) {
	dir := setupConfigEnv(t)
	writeConfigFile(t, dir, "provider: gemini\napi_key: key-from-file\n")
	t.Setenv("GEMINI_API_KEY", "key-from-env")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if cfg.APIKey != "key-from-file" {
		t.Errorf("APIKey = %q, want %q (config file should win over env fallback)", cfg.APIKey, "key-from-file")
	}
}
