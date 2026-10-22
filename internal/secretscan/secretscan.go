// Package secretscan looks for credentials in a git diff before Scribe sends
// that diff to an LLM provider. It is a pre-flight check, not a vault scanner:
// a handful of well-known token shapes are matched line by line, so a key that
// was staged by mistake is caught on this machine instead of being shipped to
// a third party along with the rest of the change.
package secretscan

import (
	"regexp"
	"strconv"
	"strings"
)

// maxDisplayLen caps how much of a diff line a Finding carries, so a match in
// a minified bundle does not flood the terminal.
const maxDisplayLen = 160

// Finding is one suspected secret in a diff.
type Finding struct {
	RuleID      string
	Description string
	// File is the path from the "diff --git a/X b/X" header.
	File string
	// Line is the line number in the new file, or in the old file for a
	// removed ("-") line.
	Line int
	// Text is the diff line, prefix included, with every match redacted.
	Text string
}

// rule pairs a pattern with what it detects. A pattern with a named "secret"
// group has only that group checked for placeholders and redacted; otherwise
// the whole match is. redact is false where the match itself is not secret,
// e.g. the BEGIN line of a private key.
type rule struct {
	id          string
	description string
	pattern     *regexp.Regexp
	redact      bool
}

// rules are ordered most specific first: when two rules match overlapping
// text, the earlier one is the one reported.
var rules = []rule{
	{
		id:          "private-key",
		description: "Private key",
		pattern:     regexp.MustCompile(`-----BEGIN (?:(?:RSA|DSA|EC|OPENSSH|PGP|ENCRYPTED) )?PRIVATE KEY(?: BLOCK)?-----`),
	},
	{
		id:          "aws-access-key-id",
		description: "AWS access key ID",
		pattern:     regexp.MustCompile(`\b(?P<secret>(?:AKIA|ASIA|ABIA|ACCA)[0-9A-Z]{16})\b`),
		redact:      true,
	},
	{
		id:          "aws-secret-access-key",
		description: "AWS secret access key",
		// Base64 ends in "/", "+" or "=", which \b cannot border, so the end of
		// the key is matched as "anything that is not base64".
		pattern: regexp.MustCompile(`(?i)aws.{0,20}?(?:secret|key).{0,20}?[:=]\s*['"]?(?P<secret>[A-Za-z0-9/+=]{40})(?:[^A-Za-z0-9/+=]|$)`),
		redact:  true,
	},
	{
		id:          "github-token",
		description: "GitHub token",
		pattern:     regexp.MustCompile(`\b(?P<secret>gh[pousr]_[A-Za-z0-9]{36,255}|github_pat_[A-Za-z0-9_]{82})\b`),
		redact:      true,
	},
	{
		id:          "slack-token",
		description: "Slack token",
		pattern:     regexp.MustCompile(`\b(?P<secret>xox[abposr]-[A-Za-z0-9-]{10,})`),
		redact:      true,
	},
	{
		id:          "slack-webhook",
		description: "Slack webhook URL",
		pattern:     regexp.MustCompile(`(?P<secret>https://hooks\.slack\.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[A-Za-z0-9]+)`),
		redact:      true,
	},
	{
		id:          "db-connection-string",
		description: "Database connection string with password",
		// The user name may be empty, as in redis://:password@host.
		pattern: regexp.MustCompile(`(?i)\b(?:postgres(?:ql)?|mysql|mariadb|mongodb(?:\+srv)?|rediss?|amqps?|mssql|sqlserver)://[^\s:/@]*:(?P<secret>[^\s@/]+)@`),
		redact:  true,
	},
	{
		id:          "generic-api-key",
		description: "Generic API key or token",
		// Only a quoted value of 20+ characters assigned to a key-like name
		// counts; bare words in prose are far too noisy to flag.
		pattern: regexp.MustCompile(`(?i)(?:api[_-]?key|secret[_-]?key|access[_-]?token|auth[_-]?token|client[_-]?secret)\b['"]?\s*(?::=|[:=])\s*['"](?P<secret>[A-Za-z0-9_\-./+=]{20,})['"]`),
		redact:  true,
	},
}

var hunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

// Scan returns every suspected secret in diff, in the order they appear. It
// checks added, removed and context lines alike, since all of them are sent to
// the provider.
func Scan(diff string) []Finding {
	var findings []Finding

	var file string
	var oldLine, newLine int
	inHunk := false

	for raw := range strings.SplitSeq(diff, "\n") {
		line := strings.TrimSuffix(raw, "\r")

		if header, ok := strings.CutPrefix(line, "diff --git "); ok {
			file = newPath(header)
			inHunk = false
			continue
		}
		if m := hunkHeader.FindStringSubmatch(line); m != nil {
			oldLine, _ = strconv.Atoi(m[1])
			newLine, _ = strconv.Atoi(m[2])
			inHunk = true
			continue
		}
		// Outside a hunk are only the index/mode/---/+++ headers.
		if !inHunk || line == "" {
			continue
		}

		var lineNo int
		switch line[0] {
		case '+':
			lineNo = newLine
			newLine++
		case '-':
			lineNo = oldLine
			oldLine++
		case '\\': // "\ No newline at end of file"
			continue
		default:
			lineNo = newLine
			oldLine++
			newLine++
		}

		findings = append(findings, scanLine(file, lineNo, line)...)
	}

	return findings
}

type match struct {
	rule       *rule
	start, end int // span checked for overlap
	secretFrom int // span that gets redacted
	secretTo   int
}

// scanLine matches every rule against one diff line, drops placeholder values
// and matches overlapping a more specific rule's, and returns one Finding per
// match left over.
func scanLine(file string, lineNo int, line string) []Finding {
	content := line[1:]
	var matches []match

	for i := range rules {
		r := &rules[i]
		secretIdx := r.pattern.SubexpIndex("secret")

		for _, loc := range r.pattern.FindAllStringSubmatchIndex(content, -1) {
			m := match{rule: r, start: loc[0], end: loc[1], secretFrom: loc[0], secretTo: loc[1]}
			if secretIdx > 0 {
				m.secretFrom, m.secretTo = loc[2*secretIdx], loc[2*secretIdx+1]
			}
			if r.redact && isPlaceholder(content[m.secretFrom:m.secretTo]) {
				continue
			}
			if overlapsAny(m, matches) {
				continue
			}
			matches = append(matches, m)
		}
	}

	if len(matches) == 0 {
		return nil
	}

	text := line[:1] + redactLine(content, matches)
	findings := make([]Finding, 0, len(matches))
	for _, m := range matches {
		findings = append(findings, Finding{
			RuleID:      m.rule.id,
			Description: m.rule.description,
			File:        file,
			Line:        lineNo,
			Text:        text,
		})
	}
	return findings
}

func overlapsAny(m match, existing []match) bool {
	for _, e := range existing {
		if m.start < e.end && e.start < m.end {
			return true
		}
	}
	return false
}

// redactLine masks each match's secret span in content and trims the result
// to maxDisplayLen, keeping the first match in view.
func redactLine(content string, matches []match) string {
	var spans []match
	for _, m := range matches {
		if m.rule.redact {
			spans = append(spans, m)
		}
	}
	// Matches come out grouped by rule, not by position.
	for i := 1; i < len(spans); i++ {
		for j := i; j > 0 && spans[j].secretFrom < spans[j-1].secretFrom; j-- {
			spans[j], spans[j-1] = spans[j-1], spans[j]
		}
	}

	var sb strings.Builder
	prev := 0
	for _, s := range spans {
		sb.WriteString(content[prev:s.secretFrom])
		sb.WriteString(Redact(content[s.secretFrom:s.secretTo]))
		prev = s.secretTo
	}
	sb.WriteString(content[prev:])
	out := sb.String()

	// Redaction preserves length, so the first match's offset still holds.
	first := matches[0].start
	for _, m := range matches {
		first = min(first, m.start)
	}
	return window(out, first)
}

// window cuts s down to maxDisplayLen bytes, starting a little before focus so
// the match stays visible. Bytes, not runes: the cut points are nudged off any
// UTF-8 continuation byte.
func window(s string, focus int) string {
	if len(s) <= maxDisplayLen {
		return s
	}
	start := max(0, focus-40)
	end := min(len(s), start+maxDisplayLen)
	for start > 0 && start < len(s) && isContinuation(s[start]) {
		start--
	}
	for end < len(s) && isContinuation(s[end]) {
		end++
	}

	out := s[start:end]
	if start > 0 {
		out = "…" + out
	}
	if end < len(s) {
		out += "…"
	}
	return out
}

func isContinuation(b byte) bool { return b&0xC0 == 0x80 }

// Redact masks a secret for display, keeping a few characters at each end so
// the user can recognise which value was flagged. Short values keep fewer, so
// the mask never gives most of a password away.
func Redact(s string) string {
	keep := min(4, len(s)/4)
	return s[:keep] + strings.Repeat("*", len(s)-2*keep) + s[len(s)-keep:]
}

// placeholderWords mark documentation values rather than real credentials.
var placeholderWords = []string{"example", "your_", "your-", "yourkey", "changeme", "placeholder", "redacted", "dummy"}

// placeholderValues are complete values that appear in sample connection
// strings far more often than as real passwords.
var placeholderValues = map[string]bool{"password": true, "pass": true, "secret": true}

// isPlaceholder reports whether a matched value is obviously not a real
// secret: an environment or template reference, a <hint>, a run of x's or
// stars, or a well-known sample value.
func isPlaceholder(v string) bool {
	lower := strings.ToLower(v)

	switch {
	case strings.HasPrefix(v, "$"), strings.Contains(v, "${"), strings.Contains(v, "{{"):
		return true
	case len(v) > 1 && strings.HasPrefix(v, "%") && strings.HasSuffix(v, "%"):
		return true
	case strings.HasPrefix(v, "<") && strings.HasSuffix(v, ">"):
		return true
	case placeholderValues[lower]:
		return true
	case strings.Trim(lower, "x*.-_") == "":
		return true
	}

	for _, w := range placeholderWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

// newPath returns the "b/" path from the remainder of a "diff --git a/X b/X"
// header, falling back to the raw header if it does not have that shape.
func newPath(header string) string {
	if _, after, ok := strings.Cut(header, " b/"); ok {
		return after
	}
	return header
}
