package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("model", "gemini-1.5-flash")
	viper.SetEnvPrefix("SCRIBE")
	viper.AutomaticEnv()

	homeDir, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(homeDir)
		viper.SetConfigName(".scribe")
		viper.SetConfigType("yaml")
	}

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
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
	}

	return &cfg, nil
}

func CreateDefaultConfigFile(apiKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".scribe.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	content := fmt.Sprintf("api_key: \"%s\"\nmodel: \"gemini-1.5-flash\"\n", apiKey)
	return os.WriteFile(configPath, []byte(content), 0600)
}
