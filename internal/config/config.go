package config

import (
	"fmt"
	"os"

	"github.com/alan-shabrandi/scribe/internal/secrets"
	"github.com/spf13/viper"
)

const (
	// envPrefix is the prefix for Scribe's own environment overrides, e.g.
	// SCRIBE_PROVIDER or SCRIBE_API_KEY.
	envPrefix = "SCRIBE"

	// DefaultProvider is the provider used when nothing is configured. It is
	// exported because commands that write config need the same default to
	// decide which provider a key belongs to.
	DefaultProvider = "gemini"
	DefaultModel    = "gemini-1.5-flash"
	DefaultStyle    = "conventional"
)

// KeySource names where the API key in use came from. Commands report it so a
// user can tell which of the several places is in effect without Scribe ever
// echoing the key itself.
type KeySource string

const (
	// KeySourceNone means no API key was found anywhere.
	KeySourceNone KeySource = "none"
	// KeySourceEnv is the SCRIBE_API_KEY override.
	KeySourceEnv KeySource = "SCRIBE_API_KEY environment variable"
	// KeySourceConfigFile is a plaintext api_key in ~/.scribe.yaml.
	KeySourceConfigFile KeySource = "config file (plaintext)"
	// KeySourceKeyring is the OS credential store, the recommended location.
	KeySourceKeyring KeySource = "OS credential store"
	// KeySourceProviderEnv is the provider's own conventional variable, e.g.
	// OPENAI_API_KEY.
	KeySourceProviderEnv KeySource = "provider environment variable"
)

type Config struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
	Style    string `mapstructure:"style"`
	// AutoCopy makes `scribe generate` behave as if --copy were passed. The -c
	// flag still wins when it is given explicitly.
	AutoCopy bool `mapstructure:"auto_copy"`
	// APIKeySource records which location APIKey was resolved from. It is
	// derived at load time, not read from the config file.
	APIKeySource KeySource `mapstructure:"-"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("provider", DefaultProvider)
	viper.SetDefault("model", DefaultModel)
	viper.SetDefault("style", DefaultStyle)
	viper.SetDefault("auto_copy", false)
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to locate home directory: %w", err)
	}

	viper.AddConfigPath(homeDir)
	viper.SetConfigName(".scribe")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	resolveAPIKey(&cfg)

	return &cfg, nil
}

// resolveAPIKey fills in cfg.APIKey and records where it came from. Highest
// precedence first:
//
//  1. SCRIBE_API_KEY, a deliberate per-shell override. Viper has already
//     folded it into cfg.APIKey by this point.
//  2. A plaintext api_key in ~/.scribe.yaml. It is the least safe place, but
//     it is visible and explicit, so it keeps winning over the credential
//     store: editing the file always has the effect the reader expects, and
//     upgrading Scribe never changes which key an existing setup sends.
//  3. The OS credential store, keyed by provider. This is where
//     `scribe config set api_key` writes, and the recommended location.
//  4. The provider's own conventional environment variable.
func resolveAPIKey(cfg *Config) {
	if cfg.APIKey != "" {
		if envKey := os.Getenv(envPrefix + "_API_KEY"); envKey != "" && envKey == cfg.APIKey {
			cfg.APIKeySource = KeySourceEnv
		} else {
			cfg.APIKeySource = KeySourceConfigFile
		}
		return
	}

	// Ollama runs locally and needs no key, so it is not worth waking the
	// credential store (and, on Linux, a D-Bus connection) for it.
	if cfg.Provider != "ollama" {
		// A missing key and an unreachable store are both just "nothing here":
		// the remaining fallback still deserves a chance, and
		// `scribe config keyring status` is where a broken store is diagnosed.
		if key, err := secrets.Get(cfg.Provider); err == nil && key != "" {
			cfg.APIKey = key
			cfg.APIKeySource = KeySourceKeyring
			return
		}
	}

	if key := fallbackAPIKey(cfg.Provider); key != "" {
		cfg.APIKey = key
		cfg.APIKeySource = KeySourceProviderEnv
		return
	}

	cfg.APIKeySource = KeySourceNone
}

func fallbackAPIKey(provider string) string {
	switch provider {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "gemini":
		return os.Getenv("GEMINI_API_KEY")
	default:
		return ""
	}
}
