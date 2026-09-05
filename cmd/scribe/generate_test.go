package main

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
)

// resetCopyFlag restores the --copy flag and its backing global after a test.
// generateCmd is package-level state shared by every test in this file, so both
// the parsed value and pflag's "was it typed" bit have to be put back.
func resetCopyFlag(t *testing.T) {
	t.Helper()

	f := generateCmd.Flags().Lookup("copy")
	if f == nil {
		t.Fatal("generateCmd has no --copy flag registered")
	}

	prevValue, prevChanged := copyToClipboard, f.Changed
	t.Cleanup(func() {
		copyToClipboard = prevValue
		f.Changed = prevChanged
		_ = f.Value.Set(strconv.FormatBool(prevValue))
	})

	copyToClipboard = false
	f.Changed = false
	_ = f.Value.Set("false")
}

func TestCopyFlagRegistration(t *testing.T) {
	f := generateCmd.Flags().Lookup("copy")
	if f == nil {
		t.Fatal("generateCmd has no --copy flag registered")
	}

	if f.Shorthand != "c" {
		t.Errorf("--copy shorthand = %q, want %q", f.Shorthand, "c")
	}
	if f.DefValue != "false" {
		t.Errorf("--copy default = %q, want %q", f.DefValue, "false")
	}
}

func TestCopyFlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        bool
		wantChanged bool
		wantErr     bool
	}{
		{
			name:        "shorthand -c",
			args:        []string{"-c"},
			want:        true,
			wantChanged: true,
		},
		{
			name:        "long --copy",
			args:        []string{"--copy"},
			want:        true,
			wantChanged: true,
		},
		{
			name:        "explicit --copy=true",
			args:        []string{"--copy=true"},
			want:        true,
			wantChanged: true,
		},
		{
			name:        "explicit --copy=false is still a typed flag",
			args:        []string{"--copy=false"},
			want:        false,
			wantChanged: true,
		},
		{
			name:        "flag absent",
			args:        []string{},
			want:        false,
			wantChanged: false,
		},
		{
			// pflag reads a single dash as clustered shorthands, so "-copy"
			// is -c -o -p -y and fails on the unknown 'o'.
			name:    "single-dash -copy is not supported",
			args:    []string{"-copy"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCopyFlag(t)

			err := generateCmd.ParseFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFlags(%v) = nil error, want an error", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFlags(%v) returned error: %v", tt.args, err)
			}

			if copyToClipboard != tt.want {
				t.Errorf("ParseFlags(%v) set copyToClipboard = %v, want %v", tt.args, copyToClipboard, tt.want)
			}
			if got := generateCmd.Flags().Changed("copy"); got != tt.wantChanged {
				t.Errorf("ParseFlags(%v) Changed(\"copy\") = %v, want %v", tt.args, got, tt.wantChanged)
			}
		})
	}
}

func TestResolveCopyMode(t *testing.T) {
	tests := []struct {
		name        string
		flagChanged bool
		flagValue   bool
		autoCopy    bool
		want        bool
	}{
		{
			name:     "no flag and auto_copy off commits",
			autoCopy: false,
			want:     false,
		},
		{
			name:     "no flag and auto_copy on copies",
			autoCopy: true,
			want:     true,
		},
		{
			name:        "-c wins over auto_copy off",
			flagChanged: true,
			flagValue:   true,
			autoCopy:    false,
			want:        true,
		},
		{
			name:        "--copy=false wins over auto_copy on",
			flagChanged: true,
			flagValue:   false,
			autoCopy:    true,
			want:        false,
		},
		{
			name:        "-c agrees with auto_copy on",
			flagChanged: true,
			flagValue:   true,
			autoCopy:    true,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveCopyMode(tt.flagChanged, tt.flagValue, tt.autoCopy)
			if got != tt.want {
				t.Errorf("resolveCopyMode(%v, %v, %v) = %v, want %v",
					tt.flagChanged, tt.flagValue, tt.autoCopy, got, tt.want)
			}
		})
	}
}

func TestAcceptActionLabel(t *testing.T) {
	tests := []struct {
		name     string
		copyMode bool
		want     string
	}{
		{name: "commit mode", copyMode: false, want: "Accept & Commit"},
		{name: "copy mode", copyMode: true, want: "Accept & Copy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := acceptActionLabel(tt.copyMode)
			if got != tt.want {
				t.Errorf("acceptActionLabel(%v) = %q, want %q", tt.copyMode, got, tt.want)
			}
		})
	}
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what it
// wrote. printInfo and copyAndFinish both use fmt.Printf, which resolves
// os.Stdout at call time, so swapping the package variable is enough.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}

	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			sb.Write(buf[:n])
			if err != nil {
				break
			}
		}
		done <- sb.String()
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	out := <-done
	if err := r.Close(); err != nil {
		t.Fatalf("closing pipe reader: %v", err)
	}
	return out
}

// printInfo carries the whole contract that --json output stays machine
// readable: every progress message in generate.go goes through it.
func TestPrintInfoRespectsJSONOutput(t *testing.T) {
	prev := jsonOutput
	t.Cleanup(func() { jsonOutput = prev })

	jsonOutput = false
	if got := captureStdout(t, func() { printInfo("hello %s\n", "world") }); got != "hello world\n" {
		t.Errorf("printInfo with jsonOutput=false wrote %q, want %q", got, "hello world\n")
	}

	jsonOutput = true
	if got := captureStdout(t, func() { printInfo("hello %s\n", "world") }); got != "" {
		t.Errorf("printInfo with jsonOutput=true wrote %q, want no output", got)
	}
}

// clipboardAvailable reports whether this machine can actually service a
// clipboard write. Linux needs xclip or xsel installed, and headless CI
// generally has neither, so the copy tests skip rather than fail there.
func clipboardAvailable(t *testing.T) bool {
	t.Helper()
	return clipboard.WriteAll("") == nil
}

func TestCopyAndFinishWritesClipboard(t *testing.T) {
	if !clipboardAvailable(t) {
		t.Skip("no clipboard available on this machine")
	}

	const msg = "feat(generate): add --copy flag"

	out := captureStdout(t, func() { copyAndFinish(msg) })

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("reading clipboard back: %v", err)
	}
	if got != msg {
		t.Errorf("clipboard = %q, want %q", got, msg)
	}

	// The confirmation is the only signal the user gets that anything
	// happened, and that no commit was made.
	if !strings.Contains(out, msg) {
		t.Errorf("confirmation %q does not echo the message", out)
	}
	if !strings.Contains(out, "Nothing was committed") {
		t.Errorf("confirmation %q does not say the commit was skipped", out)
	}
}

// finishWithMessage is the fork between copying and committing. Only the copy
// branch is exercised here: the commit branch shells out to git, which belongs
// in an end-to-end test rather than a unit test.
func TestFinishWithMessageCopiesWhenCopyModeOn(t *testing.T) {
	if !clipboardAvailable(t) {
		t.Skip("no clipboard available on this machine")
	}

	resetCopyFlag(t)
	copyToClipboard = true

	const msg = "docs(readme): document auto_copy"
	captureStdout(t, func() { finishWithMessage(msg) })

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("reading clipboard back: %v", err)
	}
	if got != msg {
		t.Errorf("clipboard = %q, want %q", got, msg)
	}
}
