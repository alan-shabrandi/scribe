// Package secrets stores Scribe's API keys in the operating system's native
// credential store — the macOS Keychain, the Windows Credential Manager, or
// the Linux Secret Service (GNOME Keyring, KWallet) — so a key never has to
// sit in plaintext in ~/.scribe.yaml or in a shell profile.
//
// Keys are stored per provider, so keys for OpenAI, Claude and Gemini can
// coexist and switching the `provider` setting picks up the matching key
// without re-entering it.
package secrets

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/zalando/go-keyring"
)

// Service is the name Scribe's entries carry in the credential store. It is
// what you search for when inspecting entries by hand, e.g. in Keychain
// Access, in the Windows Credential Manager, or with `secret-tool`.
const Service = "scribe"

// probeAccount is a name Scribe never stores a real key under. Available uses
// it to tell an empty credential store from an unreachable one.
const probeAccount = "__availability_probe__"

// ErrNotFound reports that the credential store is reachable but holds no key
// for the provider. It is the expected result on a machine that has not stored
// one yet, so callers treat it as "no key" rather than as a failure.
var ErrNotFound = errors.New("no API key in the OS credential store")

// Indirection for tests: a unit test must never read or write the developer's
// real credential store.
var (
	keyringSet    = keyring.Set
	keyringGet    = keyring.Get
	keyringDelete = keyring.Delete
)

// Providers lists the providers that use an API key, in the order Scribe
// reports them. Ollama runs locally and needs no key, so it is absent.
var Providers = []string{"openai", "claude", "gemini"}

// providerAliases folds interchangeable provider spellings onto a single
// account name, so a key stored while the provider was "anthropic" is still
// found after switching the setting to "claude".
var providerAliases = map[string]string{
	"anthropic": "claude",
}

// Account returns the credential-store account name used for a provider. It is
// the provider name, lowercased and de-aliased, and it is exported so
// user-facing messages can name the exact entry Scribe touched.
func Account(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if canonical, ok := providerAliases[normalized]; ok {
		return canonical
	}
	return normalized
}

// StoreName names the platform's credential store, for use in messages. It
// describes where a key would go rather than promising the store is reachable;
// Available answers that.
func StoreName() string {
	switch runtime.GOOS {
	case "darwin":
		return "the macOS Keychain"
	case "windows":
		return "the Windows Credential Manager"
	case "linux":
		return "the Secret Service keyring"
	default:
		return "the OS credential store"
	}
}

// Set stores apiKey for provider, replacing any key already stored for it.
func Set(provider, apiKey string) error {
	account := Account(provider)
	if account == "" {
		return errors.New("cannot store an API key without a provider")
	}
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("cannot store an empty API key")
	}

	if err := keyringSet(Service, account, apiKey); err != nil {
		return fmt.Errorf("store the %s API key in %s: %w", account, StoreName(), err)
	}
	return nil
}

// Get returns the stored key for provider. It returns an error wrapping
// ErrNotFound when the store holds no key for the provider, and a different
// error when the store itself could not be reached.
func Get(provider string) (string, error) {
	account := Account(provider)
	if account == "" {
		return "", fmt.Errorf("%w: no provider given", ErrNotFound)
	}

	apiKey, err := keyringGet(Service, account)
	switch {
	case errors.Is(err, keyring.ErrNotFound):
		return "", fmt.Errorf("%w for provider %q", ErrNotFound, account)
	case err != nil:
		return "", fmt.Errorf("read the %s API key from %s: %w", account, StoreName(), err)
	}
	return apiKey, nil
}

// Delete removes the stored key for provider. Deleting a key that was never
// stored returns an error wrapping ErrNotFound.
func Delete(provider string) error {
	account := Account(provider)
	if account == "" {
		return fmt.Errorf("%w: no provider given", ErrNotFound)
	}

	err := keyringDelete(Service, account)
	switch {
	case errors.Is(err, keyring.ErrNotFound):
		return fmt.Errorf("%w for provider %q", ErrNotFound, account)
	case err != nil:
		return fmt.Errorf("delete the %s API key from %s: %w", account, StoreName(), err)
	}
	return nil
}

// Available reports whether the credential store can be reached, returning nil
// when it can. A store that answers "not found" for the probe account is
// working and merely empty, which is the common case; anything else — no
// Secret Service on a headless Linux box, a platform go-keyring has no backend
// for — comes back as an error explaining why Scribe cannot use it.
func Available() error {
	if _, err := keyringGet(Service, probeAccount); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("%s is unavailable: %w", StoreName(), err)
	}
	return nil
}
