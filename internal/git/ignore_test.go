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
