package git

import (
	"testing"
)

func TestSplitDiffByFile(t *testing.T) {
	sampleDiff := `diff --git a/main.go b/main.go
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
+Added docs`

	chunks := SplitDiffByFile(sampleDiff)

	if len(chunks) != 2 {
		t.Errorf("SplitDiffByFile() expected 2 chunks, got %d", len(chunks))
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
