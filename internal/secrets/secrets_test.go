package secrets

import (
	"errors"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

// mockStore swaps go-keyring's platform backend for its in-memory mock, so a
// test never reads or writes the developer's real Keychain / Credential
// Manager and still passes on a headless CI box with no Secret Service.
func mockStore(t *testing.T) {
	t.Helper()

	keyring.MockInit()
	// The mock is a package-level global in go-keyring, so a fresh one has to
	// be installed after the test as well as before it.
	t.Cleanup(keyring.MockInit)
}

// mockStoreWithError makes every credential-store call fail, standing in for a
// machine where the store cannot be reached at all.
func mockStoreWithError(t *testing.T, err error) {
	t.Helper()

	keyring.MockInitWithError(err)
	t.Cleanup(keyring.MockInit)
}

func TestAccount(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{name: "plain provider", provider: "openai", want: "openai"},
		{name: "case is normalized", provider: "OpenAI", want: "openai"},
		{name: "surrounding space is trimmed", provider: "  gemini  ", want: "gemini"},
		{name: "anthropic is an alias for claude", provider: "anthropic", want: "claude"},
		{name: "claude is left alone", provider: "claude", want: "claude"},
		{name: "unknown provider passes through", provider: "mistral", want: "mistral"},
		{name: "empty provider", provider: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Account(tt.provider); got != tt.want {
				t.Errorf("Account(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}

func TestSetGetDeleteRoundTrip(t *testing.T) {
	mockStore(t)

	const apiKey = "sk-proj-roundtrip1234"

	if err := Set("openai", apiKey); err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	got, err := Get("openai")
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if got != apiKey {
		t.Errorf("Get() = %q, want %q", got, apiKey)
	}

	if err := Delete("openai"); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	if _, err := Get("openai"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
	}
}

func TestKeysAreScopedPerProvider(t *testing.T) {
	mockStore(t)

	if err := Set("openai", "sk-openai-key"); err != nil {
		t.Fatalf("Set(openai) returned error: %v", err)
	}
	if err := Set("gemini", "AIza-gemini-key"); err != nil {
		t.Fatalf("Set(gemini) returned error: %v", err)
	}

	for provider, want := range map[string]string{
		"openai": "sk-openai-key",
		"gemini": "AIza-gemini-key",
	} {
		got, err := Get(provider)
		if err != nil {
			t.Fatalf("Get(%q) returned error: %v", provider, err)
		}
		if got != want {
			t.Errorf("Get(%q) = %q, want %q", provider, got, want)
		}
	}

	// Switching providers must not be able to send one provider's key to
	// another's endpoint.
	if _, err := Get("claude"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(claude) error = %v, want ErrNotFound", err)
	}
}

// TestAliasedProviderSharesOneEntry guards the reason aliases exist: a user who
// stored a key as "anthropic" and then set provider to "claude" should not have
// to re-enter it.
func TestAliasedProviderSharesOneEntry(t *testing.T) {
	mockStore(t)

	const apiKey = "sk-ant-api03-alias1234"
	if err := Set("anthropic", apiKey); err != nil {
		t.Fatalf("Set(anthropic) returned error: %v", err)
	}

	got, err := Get("claude")
	if err != nil {
		t.Fatalf("Get(claude) returned error: %v", err)
	}
	if got != apiKey {
		t.Errorf("Get(claude) = %q, want the key stored under anthropic", got)
	}
}

func TestSetRejectsIncompleteInput(t *testing.T) {
	mockStore(t)

	tests := []struct {
		name     string
		provider string
		apiKey   string
	}{
		{name: "no provider", provider: "", apiKey: "sk-key"},
		{name: "empty key", provider: "openai", apiKey: ""},
		{name: "whitespace-only key", provider: "openai", apiKey: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Set(tt.provider, tt.apiKey); err == nil {
				t.Errorf("Set(%q, %q) = nil, want error", tt.provider, tt.apiKey)
			}
		})
	}
}

func TestGetWithoutProviderIsNotFound(t *testing.T) {
	mockStore(t)

	if _, err := Get(""); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(\"\") error = %v, want ErrNotFound", err)
	}
}

func TestDeleteMissingKeyIsNotFound(t *testing.T) {
	mockStore(t)

	if err := Delete("openai"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete() on an empty store error = %v, want ErrNotFound", err)
	}
}

// TestUnreachableStoreIsNotConfusedWithMissingKey matters because the two are
// handled differently: a missing key falls through to the next place a key
// might live, while an unreachable store is a problem to report.
func TestUnreachableStoreIsNotConfusedWithMissingKey(t *testing.T) {
	storeErr := errors.New("dbus: no session bus")
	mockStoreWithError(t, storeErr)

	_, err := Get("openai")
	if err == nil {
		t.Fatal("Get() on an unreachable store = nil error, want error")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("Get() on an unreachable store reported ErrNotFound: %v", err)
	}
	if !errors.Is(err, storeErr) {
		t.Errorf("Get() error %v does not wrap the underlying store error", err)
	}
}

func TestAvailable(t *testing.T) {
	t.Run("A reachable but empty store is available", func(t *testing.T) {
		mockStore(t)

		if err := Available(); err != nil {
			t.Errorf("Available() on an empty store = %v, want nil", err)
		}
	})

	t.Run("A store that cannot be reached is not available", func(t *testing.T) {
		mockStoreWithError(t, errors.New("dbus: no session bus"))

		if err := Available(); err == nil {
			t.Error("Available() on an unreachable store = nil, want error")
		}
	})

	t.Run("Availability does not depend on a key being stored", func(t *testing.T) {
		mockStore(t)

		if err := Set("openai", "sk-key"); err != nil {
			t.Fatalf("Set() returned error: %v", err)
		}
		if err := Available(); err != nil {
			t.Errorf("Available() = %v, want nil", err)
		}
		// The probe must not have left anything behind.
		if _, err := Get(probeAccount); !errors.Is(err, ErrNotFound) {
			t.Errorf("Available() stored something under the probe account: %v", err)
		}
	})
}

// TestErrorsDoNotLeakTheKey is a regression guard: these errors reach the
// terminal, and a key in an error message ends up in scrollback and CI logs.
func TestErrorsDoNotLeakTheKey(t *testing.T) {
	const apiKey = "sk-proj-supersecret1234"
	mockStoreWithError(t, errors.New("store exploded"))

	err := Set("openai", apiKey)
	if err == nil {
		t.Fatal("Set() on an unreachable store = nil error, want error")
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Errorf("Set() error leaked the API key: %v", err)
	}
}
