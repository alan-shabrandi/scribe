package git

import "testing"

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
