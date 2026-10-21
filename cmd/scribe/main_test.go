package main

import (
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestMain swaps go-keyring's platform backend for its in-memory mock for the
// whole package. Without it, any test that touches a command dealing with API
// keys would read and write the developer's real Keychain or Credential
// Manager, and would fail outright on a headless CI box with no Secret
// Service. Individual tests call resetKeyring to start from an empty store.
func TestMain(m *testing.M) {
	keyring.MockInit()
	os.Exit(m.Run())
}
