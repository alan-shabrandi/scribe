package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/alan-shabrandi/scribe/internal/secrets"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	// authProvider names the provider `auth set-key` and `auth remove-key`
	// act on; empty means the configured provider.
	authProvider string
	// authAssumeYes skips the confirmation prompt on `auth remove-key`.
	authAssumeYes bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage provider API keys in the OS credential store",
	Long: "Store, inspect and remove provider API keys in the operating system's\n" +
		"native credential store (the macOS Keychain, the Windows Credential\n" +
		"Manager, or the Linux Secret Service). Keys are stored per provider.",
}

var authSetKeyCmd = &cobra.Command{
	Use:   "set-key",
	Short: "Prompt for an API key without echoing it and store it in the credential store",
	Long: "Prompt for an API key without echoing it to the terminal, then store it\n" +
		"in the OS credential store under --provider (default: the configured\n" +
		"provider). Unlike 'scribe config set api_key', the key never appears on\n" +
		"the command line, so it stays out of shell history.\n\n" +
		"When stdin is not a terminal the key is read from its first line, so it\n" +
		"can be piped in from a secret manager or script:\n\n" +
		"  echo \"$OPENAI_API_KEY\" | scribe auth set-key --provider openai",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := providerArg(providerFlagArgs(authProvider))
		if err != nil {
			return err
		}
		// Reject before prompting, so nobody types a key just to have it refused.
		if secrets.Account(provider) == "ollama" {
			return fmt.Errorf("ollama runs locally and needs no API key")
		}

		apiKey, err := readAPIKey(cmd, provider)
		if err != nil {
			return err
		}
		return setAPIKey(cmd, provider, apiKey, false)
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show which providers have API keys in the credential store",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printKeyringStatus(cmd)
	},
}

var authRemoveKeyCmd = &cobra.Command{
	Use:   "remove-key",
	Short: "Delete a provider's API key from the credential store",
	Long: "Delete the stored API key for --provider (default: the configured\n" +
		"provider) from the OS credential store. A key in ~/.scribe.yaml or in an\n" +
		"environment variable is left alone.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := providerArg(providerFlagArgs(authProvider))
		if err != nil {
			return err
		}
		return deleteStoredKey(cmd, provider, authAssumeYes)
	},
}

// providerFlagArgs adapts a --provider flag value to the optional positional
// argument providerArg expects.
func providerFlagArgs(flag string) []string {
	if flag == "" {
		return nil
	}
	return []string{flag}
}

// readAPIKey reads a key without echoing it: from a hidden prompt when stdin
// is a terminal, otherwise from the first line of stdin.
func readAPIKey(cmd *cobra.Command, provider string) (string, error) {
	in := cmd.InOrStdin()

	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		var apiKey string
		prompt := &survey.Password{
			Message: fmt.Sprintf("Enter your API key for %s:", capitalizeFirst(secrets.Account(provider))),
		}
		if err := survey.AskOne(prompt, &apiKey); err != nil {
			return "", err
		}
		return apiKey, nil
	}

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read API key from stdin: %w", err)
	}
	return strings.TrimSpace(line), nil
}

func init() {
	for _, cmd := range []*cobra.Command{authSetKeyCmd, authRemoveKeyCmd} {
		cmd.Flags().StringVar(&authProvider, "provider", "", "Provider the API key belongs to (default: the configured provider)")
	}
	authRemoveKeyCmd.Flags().BoolVarP(&authAssumeYes, "yes", "y", false, "Skip the confirmation prompt")

	authCmd.AddCommand(authSetKeyCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authRemoveKeyCmd)
	rootCmd.AddCommand(authCmd)
}
