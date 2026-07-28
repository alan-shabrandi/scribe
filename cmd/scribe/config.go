// ./cmd/scribe/config.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Scribe configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to find home directory: %w", err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".scribe")
		viper.SetConfigType("yaml")

		_ = viper.ReadInConfig()

		viper.Set(key, value)

		configPath := filepath.Join(home, ".scribe.yaml")
		if err := viper.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write config file at %s: %w", configPath, err)
		}

		fmt.Printf("✔ Updated %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
