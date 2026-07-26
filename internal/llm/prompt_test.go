package llm

import (
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		diff         string
		style        string
		ticketID     string
		expectSubstr []string
	}{
		{
			name:     "conventional style with ticket ID",
			diff:     "diff --git a/pkg/cache.go b/pkg/cache.go",
			style:    "conventional",
			ticketID: "PROJ-123",
			// عبارت زیر اصلاح شد تا دقیقاً با پرامپت اصلی مطابقت داشته باشد:
			expectSubstr: []string{"Conventional Commits", "PROJ-123", "diff --git"},
		},
		{
			name:         "concise style without ticket ID",
			diff:         "diff --git a/README.md b/README.md",
			style:        "concise",
			ticketID:     "",
			expectSubstr: []string{"concise", "README.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildSystemPrompt(tt.diff, tt.style, tt.ticketID)

			for _, expected := range tt.expectSubstr {
				if !strings.Contains(prompt, expected) {
					t.Errorf("BuildPrompt() result missing expected substring %q.\nGot:\n%s", expected, prompt)
				}
			}
		})
	}
}
