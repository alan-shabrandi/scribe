package git

import (
	"strings"
	"testing"
)

func TestMatchesIgnorePattern(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		pattern string
		want    bool
	}{
		{"literal filename match", "a/notes.txt b/notes.txt", "notes.txt", true},
		{"literal path prefix match", "a/vendor/dep.go b/vendor/dep.go", "vendor/", true},
		{"literal no match", "a/main.go b/main.go", "notes.txt", false},
		{"glob match", "a/foo.generated.go b/foo.generated.go", "*.generated.go", true},
		{"glob no match", "a/main.go b/main.go", "*.generated.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesIgnorePattern(tt.header, tt.pattern); got != tt.want {
				t.Errorf("matchesIgnorePattern(%q, %q) = %v; want %v", tt.header, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchesIgnorePattern_Subdirectories(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		pattern string
		want    bool
	}{
		{"glob filename in subdirectory", "a/internal/git/foo.generated.go b/internal/git/foo.generated.go", "*.generated.go", true},
		{"glob directory prefix with wildcard", "a/docs/api/readme.md b/docs/api/readme.md", "docs/*.md", false},
		{"nested path match with wildcard", "a/pkg/utils/helper.go b/pkg/utils/helper.go", "pkg/*/*.go", true},
		{"nested path miss with single wildcard", "a/pkg/utils/sub/helper.go b/pkg/utils/sub/helper.go", "pkg/*/*.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesIgnorePattern(tt.header, tt.pattern); got != tt.want {
				t.Errorf("matchesIgnorePattern(%q, %q) = %v; want %v", tt.header, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchesIgnorePattern_Anchored(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		pattern string
		want    bool
	}{
		{"anchored glob matches root file", "a/config.yaml b/config.yaml", "/*.yaml", true},
		{"anchored glob does not match nested file", "a/pkg/config.yaml b/pkg/config.yaml", "/*.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesIgnorePattern(tt.header, tt.pattern); got != tt.want {
				t.Errorf("matchesIgnorePattern(%q, %q) = %v; want %v", tt.header, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestParseIgnoreLines(t *testing.T) {
	input := "*.generated.go\n\n# comment\n  vendor/  \n"
	got := parseIgnoreLines(strings.NewReader(input))
	want := []string{"*.generated.go", "vendor/"}

	if len(got) != len(want) {
		t.Fatalf("parseIgnoreLines() = %v; want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parseIgnoreLines()[%d] = %q; want %q", i, got[i], want[i])
		}
	}
}
