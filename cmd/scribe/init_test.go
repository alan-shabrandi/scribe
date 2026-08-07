package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{
			name:     "gemini",
			provider: "gemini",
			want:     "gemini-1.5-flash",
		},
		{
			name:     "openai",
			provider: "openai",
			want:     "gpt-4o-mini",
		},
		{
			name:     "ollama",
			provider: "ollama",
			want:     "llama3",
		},
		{
			name:     "Unknown provider",
			provider: "anthropic",
			want:     "",
		},
		{
			name:     "Empty provider",
			provider: "",
			want:     "",
		},
		{
			name:     "Match is case sensitive",
			provider: "OpenAI",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultModel(tt.provider)
			if got != tt.want {
				t.Errorf("getDefaultModel(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "normal capitilize function test",
			input: "word",
			want:  "Word",
		},
		{
			name:  "Uppercase input is normalized",
			input: "WORD",
			want:  "Word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capitalizeFirst(tt.input)
			if got != tt.want {
				t.Errorf("capitalizeFirst() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	//To avoid overwriting actual scribe config file, we must redirect home first.
	// t.Setenv works on mac/linux/windows
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)        //unix
	t.Setenv("USERPROFILE", tmp) //windows

	cfg := &ConfigFile{
		Provider: "openai",
		APIKey:   "sk-test",
		Model:    "gpt-4o-mini",
		Style:    "conventional",
	}

	gotPath, err := saveConfig(cfg)
	if err != nil {
		t.Fatalf("saveConfig() error = %v", err)
	}

	wantPath := filepath.Join(tmp, configFileName)
	if gotPath != wantPath {
		t.Errorf("saveConfig() Path = %q, want %q", gotPath, wantPath)
	}

	data, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("reading written config: %v", err)
	}
	var got ConfigFile
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("written file is not valid YAML: %v", err)
	}
	if got != *cfg {
		t.Errorf("round-tripped config = %+v, want %+v", got, *cfg)
	}

}
