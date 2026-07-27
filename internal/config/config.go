package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
	Style    string `mapstructure:"style"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("provider", "gemini")
	viper.SetDefault("model", "gemini-1.5-flash")
	viper.SetDefault("style", "conventional")

	viper.SetEnvPrefix("SCRIBE")
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

	if cfg.APIKey == "" {
		cfg.APIKey = fallbackAPIKey(cfg.Provider)
	}

	return &cfg, nil
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
