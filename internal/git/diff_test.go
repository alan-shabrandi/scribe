package git

import (
	"strings"
	"testing"
)

func TestSplitDiffByFile(t *testing.T) {
	tests := []struct {
		name        string
		sampleDiff  string
		wantCount   int
		wantInChunk []string
	}{
		{
			name: "Standard diff with multiple files",
			sampleDiff: `diff --git a/main.go b/main.go
index 1234567..89abcdef 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
diff --git a/README.md b/README.md
index 1111111..2222222 100644
--- a/README.md
+++ b/README.md
@@ -1,1 +1,2 @@
 # Scribe
+Added docs`,
			wantCount:   2,
			wantInChunk: []string{"a/main.go", "a/README.md"},
		},
		{
			name: "Diff containing ignored lockfiles",
			sampleDiff: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
diff --git a/go.sum b/go.sum
--- a/go.sum
+++ b/go.sum`,
			wantCount:   1,
			wantInChunk: []string{"a/main.go"},
		},
		{
			name:       "Empty diff",
			sampleDiff: "",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := SplitDiffByFile(tt.sampleDiff)

			if len(chunks) != tt.wantCount {
				t.Fatalf("SplitDiffByFile() expected %d chunks, got %d", tt.wantCount, len(chunks))
			}

			if tt.wantCount > 0 {
				for i, expectedText := range tt.wantInChunk {
					if !strings.Contains(chunks[i], expectedText) {
						t.Errorf("Chunk %d missing expected text %q", i, expectedText)
					}
				}
			}
		})
	}
}

func TestExtractTicketID(t *testing.T) {
	tests := []struct {
		branchName string
		wantTicket string
	}{
		{"feature/PROJ-123-add-cache", "PROJ-123"},
		{"fix/ABC-9999-fix-bug", "ABC-9999"},
		{"main", ""},
		{"user/john/no-ticket", ""},
		{"hotfix/abc-12-lowercase", "ABC-12"},
	}

	for _, tt := range tests {
		t.Run(tt.branchName, func(t *testing.T) {
			got := ExtractTicketID(tt.branchName)
			if got != tt.wantTicket {
				t.Errorf("ExtractTicketID(%q) = %q; want %q", tt.branchName, got, tt.wantTicket)
			}
		})
	}
}
