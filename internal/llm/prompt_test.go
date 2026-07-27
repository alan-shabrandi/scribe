package llm

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	tests := []struct {
		name         string
		diff         string
		style        string
		ticketID     string
		expectSubstr []string
	}{
		{
			name:         "conventional style with ticket ID",
			diff:         "diff --git a/pkg/cache.go b/pkg/cache.go",
			style:        "conventional",
			ticketID:     "PROJ-123",
			expectSubstr: []string{"Conventional Commits", "PROJ-123", "diff --git a/pkg/cache.go"},
		},
		{
			name:         "freeform style without ticket ID",
			diff:         "diff --git a/README.md b/README.md",
			style:        "freeform",
			ticketID:     "",
			expectSubstr: []string{"freeform style", "README.md"}, // کلمه freeform باید در خروجی باشد
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildSystemPrompt(tt.diff, tt.style, tt.ticketID)

			for _, expected := range tt.expectSubstr {
				if !strings.Contains(prompt, expected) {
					t.Errorf("BuildSystemPrompt() missing expected substring %q.\nGot:\n%s", expected, prompt)
				}
			}
		})
	}
}

func TestBuildFileSummaryPrompt(t *testing.T) {
	diff := "diff --git a/main.go b/main.go\n+ fmt.Println(\"Hello\")"
	prompt := buildFileSummaryPrompt(diff)

	expectedSubstrs := []string{
		"Summarize the key functional changes",
		"main.go",
		"fmt.Println",
	}

	for _, expected := range expectedSubstrs {
		if !strings.Contains(prompt, expected) {
			t.Errorf("buildFileSummaryPrompt() missing expected substring %q.", expected)
		}
	}
}

func TestBuildSummaryBasedPrompt(t *testing.T) {
	summaries := []string{
		"added new auth middleware",
		"fixed database connection timeout",
	}
	style := "conventional"
	ticketID := "AUTH-99"

	prompt := BuildSummaryBasedPrompt(summaries, style, ticketID)

	expectedSubstrs := []string{
		"Conventional Commits",
		"AUTH-99",
		"- added new auth middleware",
		"- fixed database connection timeout",
	}

	for _, expected := range expectedSubstrs {
		if !strings.Contains(prompt, expected) {
			t.Errorf("BuildSummaryBasedPrompt() missing expected substring %q.", expected)
		}
	}
}

func TestBuildMultiChoicePrompt(t *testing.T) {
	diff := "diff --git a/config.yaml b/config.yaml\n+ port: 8080"
	style := "freeform"
	ticketID := ""

	prompt := BuildMultiChoicePrompt(diff, style, ticketID)

	expectedSubstrs := []string{
		"EXACTLY 3 distinct candidate",
		"freeform style",
		"config.yaml",
	}

	for _, expected := range expectedSubstrs {
		if !strings.Contains(prompt, expected) {
			t.Errorf("BuildMultiChoicePrompt() missing expected substring %q.", expected)
		}
	}
}
