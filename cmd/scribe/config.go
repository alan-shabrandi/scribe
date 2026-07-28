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

		configPath := filepath.Join(home, ".scribe.yaml")

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
		} else {
			if err := v.WriteConfigAs(configPath); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		}

		fmt.Printf("✔ Updated %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
