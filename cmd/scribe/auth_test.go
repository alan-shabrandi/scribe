package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// withAuthFlags sets the `auth` flag variables for one test and restores them
// afterwards. Tests invoke RunE directly, which bypasses flag parsing.
func withAuthFlags(t *testing.T, provider string, assumeYes bool) {
	t.Helper()

	origProvider, origAssumeYes := authProvider, authAssumeYes
	authProvider, authAssumeYes = provider, assumeYes
	t.Cleanup(func() {
		authProvider, authAssumeYes = origProvider, origAssumeYes
	})
}

// withStdin feeds input to cmd as stdin for one test. A strings.Reader is not
// a terminal, so set-key reads the key from it rather than prompting.
func withStdin(t *testing.T, cmd *cobra.Command, input string) {
	t.Helper()

	cmd.SetIn(strings.NewReader(input))
	t.Cleanup(func() { cmd.SetIn(nil) })
}

func TestAuthSetKeyStoresPipedKey(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	clearConfigEnv(t)
	withAuthFlags(t, "", false)
	withStdin(t, authSetKeyCmd, testOpenAIKey+"\n")

	out, _, err := runE(t, authSetKeyCmd)
	if err != nil {
		t.Fatalf("auth set-key returned error: %v", err)
	}

	if got := storedKey(t, "openai"); got != testOpenAIKey {
		t.Errorf("credential store holds %q, want %q", got, testOpenAIKey)
	}
	if strings.Contains(out, testOpenAIKey) {
		t.Errorf("auth set-key echoed the raw key:\n%s", out)
	}
	if _, present := readConfig(t, path)["api_key"]; present {
		t.Error("auth set-key wrote the key into the config file")
	}
}

func TestAuthSetKeyUsesProviderFlag(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	clearConfigEnv(t)
	withAuthFlags(t, "claude", false)
	withStdin(t, authSetKeyCmd, testClaudeKey)

	if _, _, err := runE(t, authSetKeyCmd); err != nil {
		t.Fatalf("auth set-key returned error: %v", err)
	}

	if got := storedKey(t, "claude"); got != testClaudeKey {
		t.Errorf("claude entry = %q, want %q", got, testClaudeKey)
	}
	if got := storedKey(t, "openai"); got != "" {
		t.Errorf("the key was also filed under the configured provider: %q", got)
	}
}

func TestAuthSetKeyRejectsEmptyInput(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	clearConfigEnv(t)
	withAuthFlags(t, "", false)
	withStdin(t, authSetKeyCmd, "   \n")

	if _, _, err := runE(t, authSetKeyCmd); err == nil {
		t.Fatal("auth set-key accepted an empty key")
	}
	if got := storedKey(t, "openai"); got != "" {
		t.Errorf("an empty key was stored: %q", got)
	}
}

func TestAuthSetKeyRejectsOllama(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: ollama\n")
	clearConfigEnv(t)
	withAuthFlags(t, "", false)
	withStdin(t, authSetKeyCmd, "anything\n")

	_, _, err := runE(t, authSetKeyCmd)
	if err == nil || !strings.Contains(err.Error(), "needs no API key") {
		t.Fatalf("auth set-key for ollama returned %v, want a 'needs no API key' error", err)
	}
}

func TestAuthStatusListsStoredKeys(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	clearConfigEnv(t)
	withAuthFlags(t, "claude", false)
	withStdin(t, authSetKeyCmd, testClaudeKey)
	if _, _, err := runE(t, authSetKeyCmd); err != nil {
		t.Fatalf("seeding via auth set-key: %v", err)
	}

	out, _, err := runE(t, authStatusCmd)
	if err != nil {
		t.Fatalf("auth status returned error: %v", err)
	}
	if strings.Contains(out, testClaudeKey) {
		t.Errorf("auth status leaked the stored key:\n%s", out)
	}
	for _, want := range []string{maskSecret(testClaudeKey), "openai", "claude", "gemini"} {
		if !strings.Contains(out, want) {
			t.Errorf("auth status output missing %q:\n%s", want, out)
		}
	}
}

func TestAuthRemoveKey(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	clearConfigEnv(t)
	for provider, key := range map[string]string{"openai": testOpenAIKey, "claude": testClaudeKey} {
		withAuthFlags(t, provider, true)
		withStdin(t, authSetKeyCmd, key)
		if _, _, err := runE(t, authSetKeyCmd); err != nil {
			t.Fatalf("seeding %s via auth set-key: %v", provider, err)
		}
	}

	withAuthFlags(t, "openai", true)
	if _, _, err := runE(t, authRemoveKeyCmd); err != nil {
		t.Fatalf("auth remove-key returned error: %v", err)
	}

	if got := storedKey(t, "openai"); got != "" {
		t.Errorf("the openai key survived removal: %q", got)
	}
	if got := storedKey(t, "claude"); got != testClaudeKey {
		t.Errorf("claude entry = %q, want it untouched", got)
	}

	out, _, err := runE(t, authRemoveKeyCmd)
	if err != nil {
		t.Fatalf("removing an already-removed key returned error: %v", err)
	}
	if !strings.Contains(out, "nothing to delete") {
		t.Errorf("auth remove-key did not say there was nothing to do:\n%s", out)
	}
}

func TestAuthCommandsAreWired(t *testing.T) {
	var auth *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "auth" {
			auth = cmd
		}
	}
	if auth == nil {
		t.Fatal("'scribe auth' is not registered on the root command")
	}

	subcommands := map[string]bool{}
	for _, cmd := range auth.Commands() {
		subcommands[cmd.Name()] = true
	}
	for _, name := range []string{"set-key", "status", "remove-key"} {
		if !subcommands[name] {
			t.Errorf("'scribe auth %s' is not registered", name)
		}
	}

	for _, tt := range []struct {
		cmd  *cobra.Command
		flag string
	}{
		{cmd: authSetKeyCmd, flag: "provider"},
		{cmd: authRemoveKeyCmd, flag: "provider"},
		{cmd: authRemoveKeyCmd, flag: "yes"},
	} {
		if tt.cmd.Flags().Lookup(tt.flag) == nil {
			t.Errorf("%s has no --%s flag", tt.cmd.Name(), tt.flag)
		}
	}
}
