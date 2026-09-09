package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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
			return nil
		}

		// Never print the real key: masking is a display concern, so it happens
		// here rather than in loadRawConfig, which other callers rely on.
		if k, ok := raw["api_key"].(string); ok && k != "" {
			raw["api_key"] = maskSecret(k)
		}

		out, err := yaml.Marshal(raw)
		if err != nil {
			return fmt.Errorf("render config %s: %w", configPath, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "using config file at path: %s\n\n%s", configPath, out)

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		configPath, err := configFilePath()
		if err != nil {
			return fmt.Errorf("failed to find home directory: %w", err)
		}

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

		if key == "api_key" {
			value = strings.TrimSpace(value)
			if value == "" {
				return fmt.Errorf("API key cannot be empty")
			}
			if warning := apiKeyFormatWarning(v.GetString("provider"), value); warning != "" {
				fmt.Println(warning)
			}
		}

		v.Set(key, value)

		if fileExists {
			if err := v.WriteConfig(); err != nil {
				return fmt.Errorf("failed to update config file: %w", err)
			}
		} else {
			if err := v.WriteConfigAs(configPath); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		}

		fmt.Printf("✔ Updated %s = %s\n", key, value)
		return nil
	},
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
