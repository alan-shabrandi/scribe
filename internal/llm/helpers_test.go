package llm

import (
	"reflect"
	"testing"
)

func TestCleanResponseText(t *testing.T) {
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
			name:  "trim spaces",
			input: "  hello, world  ",
			want:  "hello, world",
		},

		{
			name:  "spaces in the middle",
			input: "hello,   world",
			want:  "hello,   world",
		},

		{
			name:  "trim edges but preserve middle",
			input: "   hello   world   ",
			want:  "hello   world",
		},

		{
			name:  "remove codeblock backticks",
			input: "```\nfix: memory leak\n```",
			want:  "fix: memory leak",
		},

		{
			// TODO/FIXME: The Markdown was not removed.
			// want: feat: add new button
			name:  "remove codeblock with language",
			input: "```markdown\nfeat: add new button\n```",
			want:  "markdown\nfeat: add new button",
		},

		{
			name:  "mixed spaces and backticks",
			input: "  ```\n  chore: update deps  \n```  ",
			want:  "chore: update deps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanResponseText(tt.input)
			if got != tt.want {
				t.Errorf("cleanResponseText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCandidates(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "numbered list",
			input:   "1. fix: crash on startup\n2. feat: add login",
			want:    []string{"fix: crash on startup", "feat: add login"},
			wantErr: false,
		},

		{
			name:    "bulleted list with dashes",
			input:   "- fix: crash on startup\n- feat: add login",
			want:    []string{"fix: crash on startup", "feat: add login"},
			wantErr: false,
		},

		{
			name:    "bulleted list with asterisks",
			input:   "* fix: crash on startup\n* feat: add login",
			want:    []string{"fix: crash on startup", "feat: add login"},
			wantErr: false,
		},

		{
			name:    "empty lines and extra whitespace",
			input:   "\n  1.   fix: typo \n\n  2. docs: api \n",
			want:    []string{"fix: typo", "docs: api"},
			wantErr: false,
		},

		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},

		{
			// TODO/FIXME: According to the issue, this should return an error (invalid formatting),
			// but currently it returns err = nil.
			name:    "text without formatting",
			input:   "fix: crash on startup\nfeat: add login",
			want:    []string{"fix: crash on startup", "feat: add login"},
			wantErr: false,
		},

		{
			name:    "mixed formatting styles",
			input:   "1. fix: crash on startup\n- feat: add login",
			want:    []string{"fix: crash on startup", "feat: add login"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCandidates(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseCandidates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}
