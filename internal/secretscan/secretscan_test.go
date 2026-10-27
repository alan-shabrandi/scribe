package secretscan

import (
	"strings"
	"testing"
)

// Every fake secret below is assembled from pieces. Written out whole, they
// would trip this very scanner when the test file is committed with Scribe,
// and GitHub push protection would refuse the push.
var (
	fakeAWSKeyID     = "AKIA" + "Z7QW3RTY5UIO2PLK"
	fakeAWSSecret    = "q8Hn2" + "Lp0vXr7TzYk4Wm9Bc1Ds6Fg3Jh5Ka8Nb2V/" // 40 chars, ending in "/"
	fakeGitHubToken  = "ghp" + "_" + strings.Repeat("aB3d", 9)
	fakeGitHubPAT    = "github" + "_pat_" + strings.Repeat("A1b2C", 16) + "zz"
	fakeSlackToken   = "xox" + "b-1234567890-0987654321-AbCdEfGhIjKlMnOp"
	fakeSlackWebhook = "https://hooks.slack.com/services/" + "T0123ABCD/B0456EFGH/" + "zYxWvUtSrQpOnMlK"
	fakePrivateKey   = "-----BEGIN " + "RSA PRIVATE KEY-----"
	fakeGeneric      = "f3a9" + "c1d7e5b2a8046c3f9e1d"
)

// diffOf wraps lines in a single-file, single-hunk diff starting at line 1.
func diffOf(file string, lines ...string) string {
	return "diff --git a/" + file + " b/" + file + "\n" +
		"index 1111111..2222222 100644\n" +
		"--- a/" + file + "\n" +
		"+++ b/" + file + "\n" +
		"@@ -1,1 +1,1 @@\n" +
		strings.Join(lines, "\n") + "\n"
}

func TestScanDetectsEachRule(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		rule   string
		secret string // must not survive into Finding.Text; "" when not redacted
	}{
		{"aws access key id", "+AWS_ACCESS_KEY_ID=" + fakeAWSKeyID, "aws-access-key-id", fakeAWSKeyID},
		{"aws secret key", `+aws_secret_access_key = "` + fakeAWSSecret + `"`, "aws-secret-access-key", fakeAWSSecret},
		{"aws secret key at end of line", "+AWS_SECRET_ACCESS_KEY=" + fakeAWSSecret, "aws-secret-access-key", fakeAWSSecret},
		{"github classic token", `+token := "` + fakeGitHubToken + `"`, "github-token", fakeGitHubToken},
		{"github fine-grained pat", "+GH_TOKEN=" + fakeGitHubPAT, "github-token", fakeGitHubPAT},
		{"slack token", "+SLACK_BOT_TOKEN=" + fakeSlackToken, "slack-token", fakeSlackToken},
		{"slack webhook", "+url: " + fakeSlackWebhook, "slack-webhook", fakeSlackWebhook},
		{"private key", "+" + fakePrivateKey, "private-key", ""},
		{"postgres url", `+dsn := "postgres://app:` + `hunter2pw@db:5432/app"`, "db-connection-string", "hunter2pw"},
		{"mongodb+srv url", "+MONGO_URI=mongodb+srv://admin:" + "S3cr3tPa55@cluster0.mongodb.net/db", "db-connection-string", "S3cr3tPa55"},
		{"redis url without user", "+REDIS_URL=redis://:" + "Zq81mx0Lp@cache:6379", "db-connection-string", "Zq81mx0Lp"},
		{"generic yaml api key", `+api_key: "` + fakeGeneric + `"`, "generic-api-key", fakeGeneric},
		{"generic go assignment", `+clientSecret := "` + fakeGeneric + `"`, "generic-api-key", fakeGeneric},
		{"generic json", `+  "access_token": "` + fakeGeneric + `",`, "generic-api-key", fakeGeneric},
		{"removed line still scanned", "-AWS_ACCESS_KEY_ID=" + fakeAWSKeyID, "aws-access-key-id", fakeAWSKeyID},
		{"context line still scanned", " AWS_ACCESS_KEY_ID=" + fakeAWSKeyID, "aws-access-key-id", fakeAWSKeyID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := Scan(diffOf("conf.txt", tt.line))
			if len(findings) != 1 {
				t.Fatalf("Scan found %d findings, want 1: %+v", len(findings), findings)
			}

			f := findings[0]
			if f.RuleID != tt.rule {
				t.Errorf("RuleID = %q, want %q", f.RuleID, tt.rule)
			}
			if f.File != "conf.txt" || f.Line != 1 {
				t.Errorf("location = %s:%d, want conf.txt:1", f.File, f.Line)
			}
			if tt.secret != "" && strings.Contains(f.Text, tt.secret) {
				t.Errorf("Text %q leaks the secret", f.Text)
			}
			if tt.secret != "" && !strings.Contains(f.Text, Redact(tt.secret)) {
				t.Errorf("Text %q does not contain the redacted secret %q", f.Text, Redact(tt.secret))
			}
			if f.Text[0] != tt.line[0] {
				t.Errorf("Text %q lost the diff prefix %q", f.Text, tt.line[:1])
			}
		})
	}
}

func TestScanIgnoresPlaceholdersAndCleanLines(t *testing.T) {
	lines := []string{
		"+AWS_ACCESS_KEY_ID=AKIAIOSFODNN7" + "EXAMPLE",
		`+dsn := "postgres://app:${DB_PASSWORD}@db:5432/app"`,
		"+DATABASE_URL=postgres://user:password@localhost/dev",
		"+DATABASE_URL=mysql://root:<your-password>@localhost/dev",
		"+DATABASE_URL=postgres://localhost/dev",
		`+api_key: "your_api_key_goes_here_123"`,
		`+api_key: "xxxxxxxxxxxxxxxxxxxxxxxx"`,
		`+api_key = os.Getenv("SCRIBE_API_KEY")`,
		`+apiKeyFile := "/home/user/.config/keyfile.txt"`,
		"+// Rotate the GitHub token and the Slack webhook every 90 days.",
		"+ghp_tooShort",
	}

	if findings := Scan(diffOf("README.md", lines...)); len(findings) != 0 {
		t.Errorf("Scan flagged placeholders or clean lines: %+v", findings)
	}
}

func TestScanTracksFilesAndLineNumbers(t *testing.T) {
	diff := "diff --git a/app/config.go b/app/config.go\n" +
		"index 1111111..2222222 100644\n" +
		"--- a/app/config.go\n" +
		"+++ b/app/config.go\n" +
		"@@ -10,4 +10,5 @@ func load() {\n" +
		" \tctx := context.Background()\n" + // old 10, new 10
		"-\told := \"" + fakeGitHubToken + "\"\n" + // old 11
		"+\tkey := \"" + fakeAWSKeyID + "\"\n" + // new 11
		"+\tother := 1\n" + // new 12
		" \treturn ctx\n" + // old 12, new 13
		"@@ -40,2 +41,3 @@\n" +
		" }\n" + // new 41
		"+var hook = \"" + fakeSlackWebhook + "\"\n" + // new 42
		"\\ No newline at end of file\n" +
		"diff --git a/deploy/id_rsa b/deploy/id_rsa\n" +
		"new file mode 100600\n" +
		"--- /dev/null\n" +
		"+++ b/deploy/id_rsa\n" +
		"@@ -0,0 +1,2 @@\n" +
		"+" + fakePrivateKey + "\r\n" + // new 1, CRLF
		"+MIIEow\r\n"

	want := []struct {
		file string
		line int
		rule string
	}{
		{"app/config.go", 11, "github-token"},
		{"app/config.go", 11, "aws-access-key-id"},
		{"app/config.go", 42, "slack-webhook"},
		{"deploy/id_rsa", 1, "private-key"},
	}

	got := Scan(diff)
	if len(got) != len(want) {
		t.Fatalf("Scan found %d findings, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].File != w.file || got[i].Line != w.line || got[i].RuleID != w.rule {
			t.Errorf("finding %d = %s:%d [%s], want %s:%d [%s]",
				i, got[i].File, got[i].Line, got[i].RuleID, w.file, w.line, w.rule)
		}
	}
	if strings.HasSuffix(got[3].Text, "\r") {
		t.Errorf("Text %q kept the carriage return", got[3].Text)
	}
}

// A token assigned to a key-like name matches both its own rule and the
// generic one; only the more specific rule should be reported.
func TestScanReportsOverlappingMatchOnce(t *testing.T) {
	findings := Scan(diffOf("conf.yaml", `+api_key: "`+fakeGitHubToken+`"`))
	if len(findings) != 1 || findings[0].RuleID != "github-token" {
		t.Fatalf("Scan = %+v, want a single github-token finding", findings)
	}
}

func TestScanReportsEachSecretOnALine(t *testing.T) {
	line := "+export AWS_ACCESS_KEY_ID=" + fakeAWSKeyID + " SLACK=" + fakeSlackToken
	findings := Scan(diffOf(".env", line))
	if len(findings) != 2 {
		t.Fatalf("Scan found %d findings, want 2: %+v", len(findings), findings)
	}
	for _, s := range []string{fakeAWSKeyID, fakeSlackToken} {
		if strings.Contains(findings[0].Text, s) {
			t.Errorf("Text %q leaks %q", findings[0].Text, s)
		}
	}
}

func TestScanTrimsLongLines(t *testing.T) {
	line := "+" + strings.Repeat("a", 500) + " key=" + fakeAWSKeyID + " " + strings.Repeat("b", 500)
	findings := Scan(diffOf("bundle.min.js", line))
	if len(findings) != 1 {
		t.Fatalf("Scan found %d findings, want 1", len(findings))
	}

	text := findings[0].Text
	if len(text) > maxDisplayLen+10 {
		t.Errorf("Text is %d bytes, want at most about %d", len(text), maxDisplayLen)
	}
	if !strings.Contains(text, Redact(fakeAWSKeyID)) {
		t.Errorf("trimmed Text %q lost the match", text)
	}
}

func TestRedact(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "***"},
		{"hunter2pw", "hu*****pw"},
		{"0123456789abcdefghij", "0123************ghij"},
	}

	for _, tt := range tests {
		if got := Redact(tt.in); got != tt.want {
			t.Errorf("Redact(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
