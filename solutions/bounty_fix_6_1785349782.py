Here is the solution to create the `internal/llm/helpers_test.go` file with comprehensive, table-driven unit tests for `cleanResponseText` and `parseCandidates`.

### Code Implementation

Create the file `internal/llm/helpers_test.go`:

```go
package llm

import (
	"reflect"
	"testing"
)

func TestCleanResponseText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  \n",
			expected: "",
		},
		{
			name:     "trim surrounding whitespace",
			input:    "  hello world \n\t",
			expected: "hello world",
		},
		{
			name:     "markdown code block with json tag",
			input:    "```json\n{\"result\": true}\n```",
			expected: "{\"result\": true}",
		},
		{
			name:     "markdown code block without language tag",
			input:    "```\nplain text code block\n```",
			expected: "plain text code block",
		},
		{
			name:     "code block surrounded by whitespace",
			input:    "   ```python\nprint('hello')\n```   \n",
			expected: "print('hello')",
		},
		{
			name:     "multiline content inside code block",
			input:    "```yaml\nkey1: value1\nkey2: value2\n```",
			expected: "key1: value1\nkey2: value2",
		},
		{
			name:     "plain multiline text without code block",
			input:    "\n  Line 1\nLine 2  \n",
			expected: "Line 1\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanResponseText(tt.input)
			if got != tt.expected {
				t.Errorf("cleanResponseText(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseCandidates(t *testing.T) {
	tests := []struct {
		name        string
		candidates  []Candidate
		expected    []string
		expectErr   bool
	}{
		{
			name:       "nil candidates",
			candidates: nil,
			expected:   nil,
			expectErr:  false,
		},
		{
			name:       "empty candidates slice",
			candidates: []Candidate{},
			expected:   []string{},
			expectErr:  false,
		},
		{
			name: "single candidate with valid text",
			candidates: []Candidate{
				{Text: "Response content 1"},
			},
			expected:  []string{"Response content 1"},
			expectErr: false,
		},
		{
			name: "multiple candidates",
			candidates: []Candidate{
				{Text: "Candidate A"},
				{Text: "Candidate B"},
			},
			expected:  []string{"Candidate A", "Candidate B"},
			expectErr: false,
		},
		{
			name: "candidate requiring text cleaning",
			candidates: []Candidate{
				{Text: "```json\n{\"status\": \"ok\"}\n```"},
			},
			expected:  []string{"{\"status\": \"ok\"}"},
			expectErr: false,
		},
		{
			name: "empty content candidate",
			candidates: []Candidate{
				{Text: ""},
			},
			expected:  []string{""},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCandidates(tt.candidates)
			if (err != nil) != tt.expectErr {
				t.Fatalf("parseCandidates() error = %v, expectErr %v", err, tt.expectErr)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseCandidates() = %v, want %v", got, tt.expected)
			}
		})
	}
}
```

---

### Technical Overview

1. **Table-Driven Design**:
   - Built using Go's standard `t.Run()` pattern for sub-tests.
   - Allows straightforward addition of future edge cases without duplicating assertions.

2. **`cleanResponseText()` Test Coverage**:
   - **Whitespace Handling**: Validates trimming of leading/trailing spaces, newlines, and tabs.
   - **Code Fence Stripping**: Verifies extraction of internal content from triple backtick code fences (` ``` `), both with and without language specifiers (e.g., ` ```json `, ` ```python `).
   - **Multi-line Integrity**: Ensures formatting inside code blocks or plain multiline text remains preserved after cleanup.

3. **`parseCandidates()` Test Coverage**:
   - **Nil & Empty Collections**: Ensures standard slice operations handle `nil` and empty candidate slices safely without panicking.
   - **Text Extraction & Cleaning Integration**: Confirms candidates with wrapped Markdown responses are cleaned correctly during parsing.
   - **Multiple Candidate Parsing**: Verifies all candidates in a batch are processed in sequence.