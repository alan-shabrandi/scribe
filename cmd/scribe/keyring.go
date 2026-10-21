package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/alan-shabrandi/scribe/internal/config"
	"github.com/alan-shabrandi/scribe/internal/secrets"
	"github.com/spf13/cobra"
)

var (
	// keyringAssumeYes skips the confirmation prompt on `keyring delete`, for
	// scripts and for machines with no terminal to prompt on.
	keyringAssumeYes bool
	// migrateProvider names the provider a migrated key belongs to.
	migrateProvider string
)

var keyringCmd = &cobra.Command{
	Use:   "keyring",
	Short: "Inspect and manage API keys in the OS credential store",
	Long: "Scribe keeps API keys in the operating system's native credential store\n" +
		"(the macOS Keychain, the Windows Credential Manager, or the Linux Secret\n" +
		"Service) instead of in plaintext in ~/.scribe.yaml. Keys are stored per\n" +
		"provider, so keys for several providers can coexist.\n\n" +
		"Keys are written by 'scribe config set api_key' and by 'scribe init'.\n" +
		"These subcommands inspect, move and remove them.",
}

var keyringStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether the credential store is reachable and which keys it holds",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printKeyringStatus(cmd)
	},
}

// printKeyringStatus reports whether the credential store is reachable, which
// providers have a key in it, and which key `scribe generate` would use.
func printKeyringStatus(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Credential store: %s\n", secrets.StoreName())
	fmt.Fprintf(out, "Entry name:       %s\n", secrets.Service)

	if err := secrets.Available(); err != nil {
		fmt.Fprintf(out, "Availability:     %s unavailable\n\n", red("✖"))
		fmt.Fprintf(out, "%v\n\n", err)
		fmt.Fprintf(out, "On a headless Linux box there is usually no Secret Service running. Use\n")
		fmt.Fprintf(out, "an environment variable (SCRIBE_API_KEY) or 'scribe config set api_key\n")
		fmt.Fprintf(out, "--plaintext' there instead.\n")
		return nil
	}
	fmt.Fprintf(out, "Availability:     %s available\n\n", green("✔"))

	fmt.Fprintln(out, "Stored keys:")
	for _, provider := range secrets.Providers {
		key, err := secrets.Get(provider)
		switch {
		case err == nil:
			fmt.Fprintf(out, "  %-8s %s\n", provider, maskSecret(key))
		case errors.Is(err, secrets.ErrNotFound):
			fmt.Fprintf(out, "  %-8s %s\n", provider, "(none)")
		default:
			fmt.Fprintf(out, "  %-8s %s %v\n", provider, yellow("⚠️"), err)
		}
	}

	// LoadConfig is the only thing that knows the full precedence order, so
	// ask it rather than re-deriving which key actually wins.
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "\nActive provider:  %s\n", cfg.Provider)
	fmt.Fprintf(out, "Key in use:       %s\n", cfg.APIKeySource)

	return nil
}

var keyringDeleteCmd = &cobra.Command{
	Use:   "delete [provider]",
	Short: "Remove a stored API key from the OS credential store",
	Long: "Remove the stored API key for a provider, defaulting to the configured\n" +
		"provider. This only touches the credential store: a key in ~/.scribe.yaml\n" +
		"or in an environment variable is left alone.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := providerArg(args)
		if err != nil {
			return err
		}
		return deleteStoredKey(cmd, provider, keyringAssumeYes)
	},
}

// deleteStoredKey removes provider's key from the credential store, asking
// for confirmation first unless assumeYes is set.
func deleteStoredKey(cmd *cobra.Command, provider string, assumeYes bool) error {
	out := cmd.OutOrStdout()

	// Checking first turns "nothing was stored" into a clear message rather
	// than an error on an action that had nothing to do.
	if _, err := secrets.Get(provider); errors.Is(err, secrets.ErrNotFound) {
		fmt.Fprintf(out, "No %s API key in %s — nothing to delete.\n", secrets.Account(provider), secrets.StoreName())
		return nil
	} else if err != nil {
		return err
	}

	if !assumeYes {
		confirmed := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Delete the %s API key from %s?", secrets.Account(provider), secrets.StoreName()),
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirmed); err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintf(out, "%s Left the key in place.\n", yellow("🚫"))
			return nil
		}
	}

	if err := secrets.Delete(provider); err != nil {
		return err
	}

	fmt.Fprintf(out, "✔ Deleted the %s API key from %s\n", secrets.Account(provider), secrets.StoreName())
	return nil
}

var keyringMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Move a plaintext api_key out of ~/.scribe.yaml and into the credential store",
	Long: "Move the plaintext api_key in ~/.scribe.yaml into the OS credential store\n" +
		"and remove it from the file. The key is filed under the configured\n" +
		"provider, or under --provider if given.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		configPath, err := configFilePath()
		if err != nil {
			return err
		}
		raw, err := loadRawConfig(configPath)
		if err != nil {
			return err
		}

		apiKey, _ := raw["api_key"].(string)
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			fmt.Fprintf(out, "No plaintext api_key in %s — nothing to migrate.\n", configPath)
			reportStoredKey(cmd, effectiveProvider(raw), false)
			return nil
		}

		provider := strings.TrimSpace(migrateProvider)
		if provider == "" {
			provider = effectiveProvider(raw)
		}

		if err := secrets.Set(provider, apiKey); err != nil {
			return err
		}
		fmt.Fprintf(out, "✔ Copied the %s API key into %s: %s\n", secrets.Account(provider), secrets.StoreName(), maskSecret(apiKey))

		// Only now is it safe to drop the plaintext copy: the key exists in two
		// places for the length of one function call rather than zero.
		if _, err := removePlaintextAPIKey(configPath); err != nil {
			return fmt.Errorf("copied the key into %s, but could not remove the plaintext api_key from %s: %w", secrets.StoreName(), configPath, err)
		}
		fmt.Fprintf(out, "✔ Removed the plaintext api_key from %s\n", configPath)

		fmt.Fprintf(out, "\n%s The old key is still valid and was readable in plaintext — consider\n", yellow("⚠️"))
		fmt.Fprintf(out, "rotating it with your provider if that file was ever shared or backed up.\n")

		return nil
	},
}

// providerArg resolves the optional provider argument these subcommands take,
// falling back to the provider the config file names.
func providerArg(args []string) (string, error) {
	if len(args) == 1 {
		provider := strings.TrimSpace(args[0])
		if provider == "" {
			return "", fmt.Errorf("provider cannot be empty")
		}
		return provider, nil
	}

	configPath, err := configFilePath()
	if err != nil {
		return "", err
	}
	raw, err := loadRawConfig(configPath)
	if err != nil {
		return "", err
	}
	return effectiveProvider(raw), nil
}

func init() {
	keyringDeleteCmd.Flags().BoolVarP(&keyringAssumeYes, "yes", "y", false, "Skip the confirmation prompt")
	keyringMigrateCmd.Flags().StringVar(&migrateProvider, "provider", "", "Provider the api_key belongs to (default: the configured provider)")

	keyringCmd.AddCommand(keyringStatusCmd)
	keyringCmd.AddCommand(keyringDeleteCmd)
	keyringCmd.AddCommand(keyringMigrateCmd)
	configCmd.AddCommand(keyringCmd)
}
