package main

import (
	"bytes"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/alan-shabrandi/scribe/internal/secrets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

// resetKeyring gives one test an empty in-memory credential store. The mock is
// a package-level global in go-keyring, so it is reinstalled afterwards too.
func resetKeyring(t *testing.T) {
	t.Helper()

	keyring.MockInit()
	t.Cleanup(keyring.MockInit)
}

// breakKeyring makes every credential-store call fail, standing in for a
// machine with no reachable store.
func breakKeyring(t *testing.T, err error) {
	t.Helper()

	keyring.MockInitWithError(err)
	t.Cleanup(keyring.MockInit)
}

// withSetFlags sets the `config set` flag variables for one test and restores
// them afterwards. Tests invoke RunE directly, which bypasses flag parsing.
func withSetFlags(t *testing.T, plaintext bool, provider string) {
	t.Helper()

	origPlaintext, origProvider := setPlaintext, setProvider
	setPlaintext, setProvider = plaintext, provider
	t.Cleanup(func() {
		setPlaintext, setProvider = origPlaintext, origProvider
	})
}

// runE invokes a command's RunE with its output captured, returning stdout and
// stderr separately so tests can assert on both.
func runE(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	t.Helper()

	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	t.Cleanup(func() {
		cmd.SetOut(nil)
		cmd.SetErr(nil)
	})

	err := cmd.RunE(cmd, args)

	return out.String(), errOut.String(), err
}

// storedKey returns the key the credential store holds for provider, or "" if
// it holds none.
func storedKey(t *testing.T, provider string) string {
	t.Helper()

	key, err := secrets.Get(provider)
	if errors.Is(err, secrets.ErrNotFound) {
		return ""
	}
	if err != nil {
		t.Fatalf("reading the credential store: %v", err)
	}
	return key
}

// readConfig returns the parsed config file at path, or nil if there is none.
func readConfig(t *testing.T, path string) map[string]any {
	t.Helper()

	raw, err := loadRawConfig(path)
	if err != nil {
		t.Fatalf("reading config %s: %v", path, err)
	}
	return raw
}

const (
	testOpenAIKey = "sk-proj-abcdefghijkl1234"
	testClaudeKey = "sk-ant-api03-abcdefgh5678"
)

// TestSetAPIKeyStoresInCredentialStore is the heart of the feature: the key
// lands in the OS credential store and never in the config file.
func TestSetAPIKeyStoresInCredentialStore(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\nmodel: gpt-4o-mini\n")
	withSetFlags(t, false, "")

	out, _, err := runE(t, configSetCmd, "api_key", testOpenAIKey)
	if err != nil {
		t.Fatalf("config set api_key returned error: %v", err)
	}

	if got := storedKey(t, "openai"); got != testOpenAIKey {
		t.Errorf("credential store holds %q, want %q", got, testOpenAIKey)
	}
	if _, present := readConfig(t, path)["api_key"]; present {
		t.Error("config set api_key wrote the key into the config file as well")
	}
	if !strings.Contains(out, secrets.StoreName()) {
		t.Errorf("output does not say where the key went:\n%s", out)
	}
}

// TestSetAPIKeyNeverEchoesTheKey is a regression guard: `config set` used to
// print the raw key, which lands in scrollback, screenshares and CI logs.
func TestSetAPIKeyNeverEchoesTheKey(t *testing.T) {
	tests := []struct {
		name      string
		plaintext bool
	}{
		{name: "stored in the credential store", plaintext: false},
		{name: "stored in the config file", plaintext: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetKeyring(t)
			path := fakeHome(t)
			writeConfig(t, path, "provider: openai\n")
			withSetFlags(t, tt.plaintext, "")

			out, errOut, err := runE(t, configSetCmd, "api_key", testOpenAIKey)
			if err != nil {
				t.Fatalf("config set api_key returned error: %v", err)
			}

			for stream, text := range map[string]string{"stdout": out, "stderr": errOut} {
				if strings.Contains(text, testOpenAIKey) {
					t.Errorf("config set leaked the raw API key on %s:\n%s", stream, text)
				}
			}
			if want := maskSecret(testOpenAIKey); !strings.Contains(out, want) {
				t.Errorf("output does not show the masked key %q:\n%s", want, out)
			}
		})
	}
}

func TestSetAPIKeyPlaintextWritesConfigFile(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	withSetFlags(t, true, "")

	out, _, err := runE(t, configSetCmd, "api_key", testOpenAIKey)
	if err != nil {
		t.Fatalf("config set --plaintext returned error: %v", err)
	}

	if got, _ := readConfig(t, path)["api_key"].(string); got != testOpenAIKey {
		t.Errorf("config file holds api_key %q, want %q", got, testOpenAIKey)
	}
	if got := storedKey(t, "openai"); got != "" {
		t.Errorf("--plaintext still wrote %q to the credential store", got)
	}
	if !strings.Contains(out, "plaintext") {
		t.Errorf("--plaintext did not warn that the file is readable:\n%s", out)
	}
}

// TestSetAPIKeyClearsShadowingPlaintextKey matters because the config file
// outranks the credential store at load time: a leftover plaintext key would
// quietly keep the old key in use after the user thought they had replaced it.
func TestSetAPIKeyClearsShadowingPlaintextKey(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\napi_key: sk-old-plaintext-key\nmodel: gpt-4o-mini\n")
	withSetFlags(t, false, "")

	if _, _, err := runE(t, configSetCmd, "api_key", testOpenAIKey); err != nil {
		t.Fatalf("config set api_key returned error: %v", err)
	}

	raw := readConfig(t, path)
	if _, present := raw["api_key"]; present {
		t.Errorf("the shadowing plaintext api_key survived in the config file: %v", raw["api_key"])
	}
	// The rest of the config must survive the rewrite.
	if got := raw["model"]; got != "gpt-4o-mini" {
		t.Errorf("model = %v, want %q: clearing api_key dropped other settings", got, "gpt-4o-mini")
	}
	if got := storedKey(t, "openai"); got != testOpenAIKey {
		t.Errorf("credential store holds %q, want %q", got, testOpenAIKey)
	}
}

func TestSetAPIKeyUsesProviderFlag(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	withSetFlags(t, false, "claude")

	if _, _, err := runE(t, configSetCmd, "api_key", testClaudeKey); err != nil {
		t.Fatalf("config set --provider claude returned error: %v", err)
	}

	if got := storedKey(t, "claude"); got != testClaudeKey {
		t.Errorf("claude entry = %q, want %q", got, testClaudeKey)
	}
	if got := storedKey(t, "openai"); got != "" {
		t.Errorf("--provider claude also wrote the openai entry: %q", got)
	}
}

func TestSetAPIKeyRejectsProvidersThatNeedNoKey(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: ollama\n")
	withSetFlags(t, false, "")

	if _, _, err := runE(t, configSetCmd, "api_key", testOpenAIKey); err == nil {
		t.Error("config set api_key with provider ollama = nil error, want error")
	}
}

func TestSetAPIKeyRejectsEmptyValue(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	withSetFlags(t, false, "")

	if _, _, err := runE(t, configSetCmd, "api_key", "   "); err == nil {
		t.Error("config set api_key with a blank value = nil error, want error")
	}
}

// TestSetAPIKeyOnUnreachableStoreDoesNotSilentlyDowngrade checks that a failed
// store write reports the failure instead of quietly writing plaintext: the
// user asked for the secure path and must choose the fallback themselves.
func TestSetAPIKeyOnUnreachableStoreDoesNotSilentlyDowngrade(t *testing.T) {
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	withSetFlags(t, false, "")
	breakKeyring(t, errors.New("dbus: no session bus"))

	_, errOut, err := runE(t, configSetCmd, "api_key", testOpenAIKey)
	if err == nil {
		t.Fatal("config set api_key against an unreachable store = nil error, want error")
	}

	if _, present := readConfig(t, path)["api_key"]; present {
		t.Error("a failed credential-store write fell back to plaintext without being asked")
	}
	if !strings.Contains(errOut, "--plaintext") {
		t.Errorf("the failure does not mention the --plaintext escape hatch:\n%s", errOut)
	}
}

func TestSetNonSecretValueStillWritesConfigFile(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	withSetFlags(t, false, "")

	out, _, err := runE(t, configSetCmd, "provider", "claude")
	if err != nil {
		t.Fatalf("config set provider returned error: %v", err)
	}

	if got := readConfig(t, path)["provider"]; got != "claude" {
		t.Errorf("provider = %v, want %q", got, "claude")
	}
	if !strings.Contains(out, "provider = claude") {
		t.Errorf("output does not confirm the change:\n%s", out)
	}
}

func TestConfigShowReportsStoredKeyMasked(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\nmodel: gpt-4o-mini\n")
	if err := secrets.Set("openai", testOpenAIKey); err != nil {
		t.Fatalf("seeding the credential store: %v", err)
	}

	out, err := runConfigShow(t)
	if err != nil {
		t.Fatalf("config show returned error: %v", err)
	}

	if strings.Contains(out, testOpenAIKey) {
		t.Errorf("config show leaked the stored key:\n%s", out)
	}
	if want := maskSecret(testOpenAIKey); !strings.Contains(out, want) {
		t.Errorf("config show does not report the stored key as %q:\n%s", want, out)
	}
}

func TestConfigShowWarnsAboutPlaintextKey(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\napi_key: "+testOpenAIKey+"\n")

	out, err := runConfigShow(t)
	if err != nil {
		t.Fatalf("config show returned error: %v", err)
	}

	if !strings.Contains(out, "keyring migrate") {
		t.Errorf("config show does not point at the migration command:\n%s", out)
	}
}

// TestConfigShowNamesTheKeyInUse covers the confusing case of a key in both
// places: listing them without saying which one wins would send someone
// debugging the wrong key.
func TestConfigShowNamesTheKeyInUse(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\napi_key: "+testOpenAIKey+"\n")
	if err := secrets.Set("openai", testClaudeKey); err != nil {
		t.Fatalf("seeding the credential store: %v", err)
	}

	out, err := runConfigShow(t)
	if err != nil {
		t.Fatalf("config show returned error: %v", err)
	}

	if !strings.Contains(out, "takes precedence") {
		t.Errorf("config show lists two keys without saying which is in use:\n%s", out)
	}
}

func TestKeyringMigrateMovesKeyOutOfConfigFile(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\napi_key: "+testOpenAIKey+"\nmodel: gpt-4o-mini\n")

	out, _, err := runE(t, keyringMigrateCmd)
	if err != nil {
		t.Fatalf("keyring migrate returned error: %v", err)
	}

	if got := storedKey(t, "openai"); got != testOpenAIKey {
		t.Errorf("credential store holds %q, want %q", got, testOpenAIKey)
	}
	raw := readConfig(t, path)
	if _, present := raw["api_key"]; present {
		t.Error("keyring migrate left the plaintext api_key in the config file")
	}
	if got := raw["model"]; got != "gpt-4o-mini" {
		t.Errorf("model = %v, want %q: migrate dropped other settings", got, "gpt-4o-mini")
	}
	if strings.Contains(out, testOpenAIKey) {
		t.Errorf("keyring migrate echoed the raw key:\n%s", out)
	}
}

func TestKeyringMigrateWithNothingToMigrate(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")

	out, _, err := runE(t, keyringMigrateCmd)
	if err != nil {
		t.Fatalf("keyring migrate with no key returned error: %v", err)
	}
	if !strings.Contains(out, "nothing to migrate") {
		t.Errorf("keyring migrate did not say there was nothing to do:\n%s", out)
	}
}

func TestKeyringDelete(t *testing.T) {
	origAssumeYes := keyringAssumeYes
	keyringAssumeYes = true
	t.Cleanup(func() { keyringAssumeYes = origAssumeYes })

	t.Run("Removes the stored key", func(t *testing.T) {
		resetKeyring(t)
		fakeHome(t)
		if err := secrets.Set("openai", testOpenAIKey); err != nil {
			t.Fatalf("seeding the credential store: %v", err)
		}

		if _, _, err := runE(t, keyringDeleteCmd, "openai"); err != nil {
			t.Fatalf("keyring delete returned error: %v", err)
		}

		if got := storedKey(t, "openai"); got != "" {
			t.Errorf("the key survived deletion: %q", got)
		}
	})

	t.Run("Deleting a key that was never stored is not an error", func(t *testing.T) {
		resetKeyring(t)
		fakeHome(t)

		out, _, err := runE(t, keyringDeleteCmd, "openai")
		if err != nil {
			t.Fatalf("keyring delete on an empty store returned error: %v", err)
		}
		if !strings.Contains(out, "nothing to delete") {
			t.Errorf("keyring delete did not say there was nothing to do:\n%s", out)
		}
	})

	t.Run("Only the named provider is touched", func(t *testing.T) {
		resetKeyring(t)
		fakeHome(t)
		for provider, key := range map[string]string{"openai": testOpenAIKey, "claude": testClaudeKey} {
			if err := secrets.Set(provider, key); err != nil {
				t.Fatalf("seeding the credential store: %v", err)
			}
		}

		if _, _, err := runE(t, keyringDeleteCmd, "openai"); err != nil {
			t.Fatalf("keyring delete returned error: %v", err)
		}

		if got := storedKey(t, "claude"); got != testClaudeKey {
			t.Errorf("claude entry = %q, want it untouched (%q)", got, testClaudeKey)
		}
	})
}

func TestKeyringStatusListsStoredKeys(t *testing.T) {
	resetKeyring(t)
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	clearConfigEnv(t)
	if err := secrets.Set("openai", testOpenAIKey); err != nil {
		t.Fatalf("seeding the credential store: %v", err)
	}

	out, _, err := runE(t, keyringStatusCmd)
	if err != nil {
		t.Fatalf("keyring status returned error: %v", err)
	}

	if strings.Contains(out, testOpenAIKey) {
		t.Errorf("keyring status leaked the stored key:\n%s", out)
	}
	for _, want := range []string{maskSecret(testOpenAIKey), "openai", "claude", "gemini", "available"} {
		if !strings.Contains(out, want) {
			t.Errorf("keyring status output missing %q:\n%s", want, out)
		}
	}
}

func TestKeyringStatusOnUnreachableStore(t *testing.T) {
	path := fakeHome(t)
	writeConfig(t, path, "provider: openai\n")
	clearConfigEnv(t)
	breakKeyring(t, errors.New("dbus: no session bus"))

	out, _, err := runE(t, keyringStatusCmd)
	if err != nil {
		t.Fatalf("keyring status on an unreachable store returned error: %v", err)
	}
	if !strings.Contains(out, "unavailable") {
		t.Errorf("keyring status did not report the store as unavailable:\n%s", out)
	}
	if !strings.Contains(out, "--plaintext") {
		t.Errorf("keyring status did not suggest a fallback:\n%s", out)
	}
}

// clearConfigEnv isolates commands that call config.LoadConfig from the
// developer's environment and from viper's package-level globals.
func clearConfigEnv(t *testing.T) {
	t.Helper()

	viper.Reset()
	t.Cleanup(viper.Reset)

	for _, key := range []string{
		"SCRIBE_PROVIDER", "SCRIBE_API_KEY", "SCRIBE_MODEL", "SCRIBE_STYLE",
		"OPENAI_API_KEY", "GEMINI_API_KEY",
	} {
		t.Setenv(key, "")
	}
}

func TestEffectiveProvider(t *testing.T) {
	t.Run("The config file names the provider", func(t *testing.T) {
		t.Setenv("SCRIBE_PROVIDER", "")

		if got := effectiveProvider(map[string]any{"provider": "claude"}); got != "claude" {
			t.Errorf("effectiveProvider() = %q, want %q", got, "claude")
		}
	})

	t.Run("A missing provider falls back to the default", func(t *testing.T) {
		t.Setenv("SCRIBE_PROVIDER", "")

		if got := effectiveProvider(nil); got != "gemini" {
			t.Errorf("effectiveProvider(nil) = %q, want the default %q", got, "gemini")
		}
	})

	t.Run("SCRIBE_PROVIDER wins, as it does at generate time", func(t *testing.T) {
		t.Setenv("SCRIBE_PROVIDER", "openai")

		if got := effectiveProvider(map[string]any{"provider": "claude"}); got != "openai" {
			t.Errorf("effectiveProvider() = %q, want %q from SCRIBE_PROVIDER", got, "openai")
		}
	})
}

func TestRemovePlaintextAPIKey(t *testing.T) {
	t.Run("Reports nothing removed when there is no key", func(t *testing.T) {
		path := fakeHome(t)
		writeConfig(t, path, "provider: openai\n")

		removed, err := removePlaintextAPIKey(path)
		if err != nil {
			t.Fatalf("removePlaintextAPIKey() returned error: %v", err)
		}
		if removed {
			t.Error("removePlaintextAPIKey() reported a removal with no api_key present")
		}
	})

	t.Run("Reports nothing removed when there is no config file", func(t *testing.T) {
		path := fakeHome(t)

		removed, err := removePlaintextAPIKey(path)
		if err != nil {
			t.Fatalf("removePlaintextAPIKey() on a missing file returned error: %v", err)
		}
		if removed {
			t.Error("removePlaintextAPIKey() reported a removal on a missing file")
		}
	})

	t.Run("The rewritten file stays owner-only", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("file modes are not meaningful on Windows")
		}
		path := fakeHome(t)
		writeConfig(t, path, "provider: openai\napi_key: "+testOpenAIKey+"\n")

		if _, err := removePlaintextAPIKey(path); err != nil {
			t.Fatalf("removePlaintextAPIKey() returned error: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if mode := info.Mode().Perm(); mode&0o077 != 0 {
			t.Errorf("config file mode = %o, want no group or world access", mode)
		}
	})
}

// TestAskKeyStorageSkipsThePrompt covers the two cases where `scribe init` must
// not ask where to put the key: there is no key, or there is nowhere secure to
// put it. Both have to resolve without a terminal, since survey would block.
func TestAskKeyStorageSkipsThePrompt(t *testing.T) {
	t.Run("No key was entered", func(t *testing.T) {
		resetKeyring(t)

		useKeyring, err := askKeyStorage("")
		if err != nil {
			t.Fatalf("askKeyStorage() returned error: %v", err)
		}
		if useKeyring {
			t.Error("askKeyStorage() chose the credential store with no key to store")
		}
	})

	t.Run("The credential store cannot be reached", func(t *testing.T) {
		breakKeyring(t, errors.New("dbus: no session bus"))

		useKeyring, err := askKeyStorage(testOpenAIKey)
		if err != nil {
			t.Fatalf("askKeyStorage() returned error: %v", err)
		}
		if useKeyring {
			t.Error("askKeyStorage() chose a credential store that cannot be reached")
		}
	})
}

func TestKeyringFlagsAreRegistered(t *testing.T) {
	tests := []struct {
		cmd  *cobra.Command
		flag string
	}{
		{cmd: configSetCmd, flag: "plaintext"},
		{cmd: configSetCmd, flag: "provider"},
		{cmd: keyringMigrateCmd, flag: "provider"},
		{cmd: keyringDeleteCmd, flag: "yes"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd.Name()+" --"+tt.flag, func(t *testing.T) {
			if tt.cmd.Flags().Lookup(tt.flag) == nil {
				t.Errorf("%s has no --%s flag", tt.cmd.Name(), tt.flag)
			}
		})
	}
}

func TestKeyringCommandIsWiredUnderConfig(t *testing.T) {
	for _, cmd := range configCmd.Commands() {
		if cmd.Name() == "keyring" {
			return
		}
	}
	t.Error("'scribe config keyring' is not registered under 'scribe config'")
}
