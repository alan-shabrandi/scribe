package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alan-shabrandi/scribe/internal/config"
	"github.com/alan-shabrandi/scribe/internal/secrets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Flags for `config set`. Both only affect the api_key key, where the question
// of *where* a value is stored actually arises.
var (
	setPlaintext bool
	setProvider  string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Scribe configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows current config if any",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := configFilePath()
		if err != nil {
			return err
		}

		// loadRawConfig reports a missing file as a nil map, not an error.
		raw, err := loadRawConfig(configPath)
		if err != nil {
			return err
		}
		if raw == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "No config file at %s. Run 'scribe init' to create one.\n", configPath)
			reportStoredKey(cmd, effectiveProvider(nil), false)
			return nil
		}

		// Never print the real key: masking is a display concern, so it happens
		// here rather than in loadRawConfig, which other callers rely on.
		plaintextKey := false
		if k, ok := raw["api_key"].(string); ok && k != "" {
			plaintextKey = true
			raw["api_key"] = maskSecret(k)
		}

		out, err := yaml.Marshal(raw)
		if err != nil {
			return fmt.Errorf("render config %s: %w", configPath, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "using config file at path: %s\n\n%s", configPath, out)

		if plaintextKey {
			fmt.Fprintf(cmd.OutOrStdout(), "\n%s api_key is stored in plaintext in this file. Run 'scribe config keyring migrate' to move it into %s.\n",
				yellow("⚠️"), secrets.StoreName())
		}
		reportStoredKey(cmd, effectiveProvider(raw), plaintextKey)

		return nil
	},
}

// reportStoredKey prints whether the credential store holds a key for
// provider, so `config show` accounts for the key the config file does not
// mention. shadowedByFile says a plaintext api_key is also configured, which
// outranks the store and would otherwise make this line misleading. An
// unreachable store is reported rather than hidden, since otherwise a stored
// key would look as though it had vanished.
func reportStoredKey(cmd *cobra.Command, provider string, shadowedByFile bool) {
	if secrets.Account(provider) == "ollama" {
		return
	}

	out := cmd.OutOrStdout()
	key, err := secrets.Get(provider)

	switch {
	case err == nil:
		fmt.Fprintf(out, "\n%s The %s API key is stored in %s: %s\n",
			green("🔐"), secrets.Account(provider), secrets.StoreName(), maskSecret(key))
		if shadowedByFile {
			fmt.Fprintf(out, "   The plaintext api_key above takes precedence, so this one is unused.\n")
		}
	case errors.Is(err, secrets.ErrNotFound):
		fmt.Fprintf(out, "\nNo %s API key in %s.\n", secrets.Account(provider), secrets.StoreName())
	default:
		fmt.Fprintf(out, "\n%s Could not read %s: %v\n", yellow("⚠️"), secrets.StoreName(), err)
	}
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: "Set a configuration value in ~/.scribe.yaml.\n\n" +
		"The api_key value is handled differently: it goes into the operating\n" +
		"system's native credential store (the macOS Keychain, the Windows\n" +
		"Credential Manager, or the Linux Secret Service), keyed by provider, so\n" +
		"that it never lands in a plaintext file. Pass --plaintext to write it\n" +
		"into ~/.scribe.yaml instead.",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if key == "api_key" {
			return setAPIKey(cmd, setProvider, value, setPlaintext)
		}

		configPath, err := configFilePath()
		if err != nil {
			return err
		}

		if err := writeConfigValue(configPath, key, value); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✔ Updated %s = %s\n", key, value)
		return nil
	},
}

// setAPIKey stores an API key for provider (the configured provider when
// empty), in the OS credential store unless plaintext asks for ~/.scribe.yaml.
// It never echoes the key back unmasked, since the terminal it would land in
// is scrollback, screenshares and CI logs.
func setAPIKey(cmd *cobra.Command, provider, value string, plaintext bool) error {
	out := cmd.OutOrStdout()

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	configPath, err := configFilePath()
	if err != nil {
		return err
	}
	raw, err := loadRawConfig(configPath)
	if err != nil {
		return err
	}

	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = effectiveProvider(raw)
	}
	if secrets.Account(provider) == "ollama" {
		return fmt.Errorf("ollama runs locally and needs no API key")
	}

	if warning := apiKeyFormatWarning(provider, value); warning != "" {
		fmt.Fprintln(out, warning)
	}

	if plaintext {
		if err := writeConfigValue(configPath, "api_key", value); err != nil {
			return err
		}

		fmt.Fprintf(out, "✔ Stored the %s API key in %s: %s\n", secrets.Account(provider), configPath, maskSecret(value))
		fmt.Fprintf(out, "%s That file is plaintext — anything able to read it can read the key.\n", yellow("⚠️"))

		// The file outranks the credential store at load time, so an existing
		// stored key is now dead weight and the user should know.
		if _, err := secrets.Get(provider); err == nil {
			fmt.Fprintf(out, "%s The %s API key is also in %s, which the config file now takes precedence over.\n",
				yellow("⚠️"), secrets.Account(provider), secrets.StoreName())
		}
		return nil
	}

	if err := secrets.Set(provider, value); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s Re-run with --plaintext to store the key in %s instead, accepting that it will be readable there.\n",
			yellow("⚠️"), configPath)
		return err
	}

	fmt.Fprintf(out, "✔ Stored the %s API key in %s: %s\n", secrets.Account(provider), secrets.StoreName(), maskSecret(value))

	// A plaintext api_key outranks the credential store when the config loads,
	// so leaving one behind would quietly keep the old key in use.
	removed, err := removePlaintextAPIKey(configPath)
	if err != nil {
		return fmt.Errorf("stored the key, but could not clear the plaintext api_key from %s: %w", configPath, err)
	}
	if removed {
		fmt.Fprintf(out, "✔ Removed the now-shadowing plaintext api_key from %s\n", configPath)
	}

	return nil
}

// writeConfigValue sets one key in the config file, creating the file if it
// does not exist yet.
func writeConfigValue(configPath, key, value string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	fileExists := false
	if _, err := os.Stat(configPath); err == nil {
		fileExists = true
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read existing config: %w", err)
		}
	}

	v.Set(key, value)

	if fileExists {
		if err := v.WriteConfig(); err != nil {
			return fmt.Errorf("failed to update config file: %w", err)
		}
		return nil
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	// Viper creates the file world-readable; a config that may hold a key
	// should match what `scribe init` writes.
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("failed to restrict permissions on %s: %w", configPath, err)
	}
	return nil
}

// removePlaintextAPIKey drops api_key from the config file and reports whether
// there was one to drop. The file is rewritten from its parsed form, so it
// comes back without comments and with its keys sorted.
func removePlaintextAPIKey(configPath string) (bool, error) {
	raw, err := loadRawConfig(configPath)
	if err != nil {
		return false, err
	}
	if _, present := raw["api_key"]; !present {
		return false, nil
	}

	delete(raw, "api_key")

	data, err := yaml.Marshal(raw)
	if err != nil {
		return false, fmt.Errorf("render config %s: %w", configPath, err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return false, fmt.Errorf("write config %s: %w", configPath, err)
	}
	return true, nil
}

// effectiveProvider returns the provider `scribe generate` would use for the
// given parsed config file, so a key is filed under the provider that will
// actually look for it. It mirrors config.LoadConfig: SCRIBE_PROVIDER first,
// then the config file, then the built-in default.
func effectiveProvider(raw map[string]any) string {
	if env := strings.TrimSpace(os.Getenv("SCRIBE_PROVIDER")); env != "" {
		return env
	}
	if p, ok := raw["provider"].(string); ok && strings.TrimSpace(p) != "" {
		return strings.TrimSpace(p)
	}
	return config.DefaultProvider
}

// apiKeyPrefixes maps known providers to the prefix their API keys are
// documented to use, so a mistyped or wrong-provider key can be flagged.
var apiKeyPrefixes = map[string]string{
	"openai":    "sk-",
	"claude":    "sk-ant-",
	"anthropic": "sk-ant-",
	"gemini":    "AIza",
}

// apiKeyFormatWarning returns a friendly warning if value doesn't look like a
// key for provider, or "" if the provider is unknown/unset or the key matches.
func apiKeyFormatWarning(provider, value string) string {
	prefix, ok := apiKeyPrefixes[strings.ToLower(strings.TrimSpace(provider))]
	if !ok || strings.HasPrefix(value, prefix) {
		return ""
	}
	return fmt.Sprintf("⚠ Warning: %s API keys usually start with %q — double-check this value", provider, prefix)
}

// maskSecret redacts a secret for display. It keeps the documented provider
// prefix (so you can still tell an OpenAI key from an Anthropic one) and the
// last few characters (so you can confirm which key is configured), and stars
// out everything in between. Current helper function for the above config show command.
func maskSecret(s string) string {
	const keep = 4

	if len(s) <= keep {
		return strings.Repeat("*", len(s))
	}

	// Longest matching known prefix wins, e.g. "sk-ant-" over "sk-".
	prefix := ""
	for _, p := range apiKeyPrefixes {
		if strings.HasPrefix(s, p) && len(p) > len(prefix) {
			prefix = p
		}
	}
	if len(prefix)+keep >= len(s) {
		prefix = ""
	}

	return prefix + strings.Repeat("*", len(s)-len(prefix)-keep) + s[len(s)-keep:]
}

func init() {
	configSetCmd.Flags().BoolVar(&setPlaintext, "plaintext", false, "Store api_key in ~/.scribe.yaml instead of the OS credential store")
	configSetCmd.Flags().StringVar(&setProvider, "provider", "", "Provider the api_key belongs to (default: the configured provider)")

	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
}

func loadRawConfig(path string) (map[string]any, error) {
	//helper function to read and parse config file, returns nil map and nil err if file not existing
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return raw, nil
}

// helper function to return path to a users config file
func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to find home directory: %w", err)
	}
	return filepath.Join(home, configFileName), nil
}
