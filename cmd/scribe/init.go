package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/alan-shabrandi/scribe/internal/secrets"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const configFileName = ".scribe.yaml"

// Labels for the "where should the key live" question, matched on below.
const (
	storeInKeyring = "OS credential store (recommended)"
	storeInFile    = "config file, in plaintext"
)

type ConfigFile struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key,omitempty"`
	Model    string `yaml:"model"`
	Style    string `yaml:"style"`
	// Written explicitly rather than with omitempty: a visible "auto_copy: false"
	// tells the reader the setting exists.
	AutoCopy bool `yaml:"auto_copy"`
}

var (
	initGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	initRed   = color.New(color.FgRed, color.Bold).SprintFunc()
	initCyan  = color.New(color.FgCyan, color.Bold).SprintFunc()
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration for scribe CLI",
	Long:  `Interactively guides you through setting up your LLM provider, API keys, default models, and commit message style in ~/.scribe.yaml. The API key itself goes into your operating system's credential store rather than the config file, unless you choose otherwise.`,
	Run:   runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Printf("%s Welcome to Scribe Initialization!\n\n", initCyan("🚀"))

	cfg, useKeyring, err := gatherUserConfig()
	if err != nil {
		fmt.Printf("\n%s Initialization cancelled or failed: %v\n", initRed("❌"), err)
		os.Exit(1)
	}

	// Moving the key into the credential store before saveConfig is what keeps
	// it out of the YAML: ConfigFile.APIKey is omitempty.
	if useKeyring && cfg.APIKey != "" {
		if err := secrets.Set(cfg.Provider, cfg.APIKey); err != nil {
			fmt.Printf("\n%s Could not store the key in %s: %v\n", initRed("❌"), secrets.StoreName(), err)
			fmt.Printf("%s Falling back to the config file so your setup still works. Run\n", initRed("⚠"))
			fmt.Printf("  'scribe config keyring migrate'\n")
			fmt.Printf("to move it out of plaintext once the store is reachable.\n")
		} else {
			fmt.Printf("\n%s API key stored in %s under %q/%q\n",
				initGreen("🔐"), secrets.StoreName(), secrets.Service, secrets.Account(cfg.Provider))
			cfg.APIKey = ""
		}
	}

	configPath, err := saveConfig(cfg)
	if err != nil {
		fmt.Printf("\n%s Error saving config: %v\n", initRed("❌"), err)
		os.Exit(1)
	}

	fmt.Printf("\n%s Configuration successfully saved to %s!\n", initGreen("🎉"), configPath)
	if cfg.APIKey != "" {
		fmt.Printf("%s That file holds your API key in plaintext — anything able to read it can read the key.\n", initRed("⚠"))
	}
	fmt.Println("You can now run 'scribe generate' or 'scribe' from any Git repository.")
}

// gatherUserConfig prompts for every setting and reports whether the API key
// should go into the OS credential store rather than into the config file. The
// key is returned on the ConfigFile either way; runInit decides where it lands.
func gatherUserConfig() (*ConfigFile, bool, error) {
	var provider string
	if err := survey.AskOne(&survey.Select{
		Message: "Choose your LLM Provider:",
		Options: []string{"gemini", "openai", "claude", "ollama"},
		Default: "gemini",
	}, &provider); err != nil {
		return nil, false, err
	}

	defaultModel := getDefaultModel(provider)

	var model string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter default model name:",
		Default: defaultModel,
	}, &model); err != nil {
		return nil, false, err
	}

	var apiKey string
	if provider != "ollama" {
		if err := survey.AskOne(&survey.Password{
			Message: fmt.Sprintf("Enter your API Key for %s:", capitalizeFirst(provider)),
		}, &apiKey); err != nil {
			return nil, false, err
		}
	}

	useKeyring, err := askKeyStorage(apiKey)
	if err != nil {
		return nil, false, err
	}

	var style string
	if err := survey.AskOne(&survey.Select{
		Message: "Choose default commit message style:",
		Options: []string{"conventional", "freeform"},
		Default: "conventional",
		Help:    "conventional: feat(scope): message | freeform: descriptive plain text",
	}, &style); err != nil {
		return nil, false, err
	}

	return &ConfigFile{
		Provider: provider,
		APIKey:   apiKey,
		Model:    model,
		Style:    style,
		AutoCopy: false,
	}, useKeyring, nil
}

// askKeyStorage asks where the API key should be kept. The question is only
// worth asking when there is a key to keep and the credential store can
// actually be reached — on a headless Linux box there is often no Secret
// Service, and offering it there would just produce a confusing failure.
func askKeyStorage(apiKey string) (bool, error) {
	if strings.TrimSpace(apiKey) == "" {
		return false, nil
	}

	if err := secrets.Available(); err != nil {
		fmt.Printf("\n%s %s is not reachable, so the key will go into the config file in plaintext.\n", initRed("⚠"), secrets.StoreName())
		fmt.Printf("  (%v)\n\n", err)
		return false, nil
	}

	var choice string
	if err := survey.AskOne(&survey.Select{
		Message: "Where should your API key be stored?",
		Options: []string{storeInKeyring, storeInFile},
		Default: storeInKeyring,
		Help:    fmt.Sprintf("The credential store is %s. The config file is readable by anything that can read ~/%s.", secrets.StoreName(), configFileName),
	}, &choice); err != nil {
		return false, err
	}

	return choice == storeInKeyring, nil
}

func saveConfig(cfg *ConfigFile) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locating home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configFileName)

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("generating YAML config: %w", err)
	}

	if err := os.WriteFile(configPath, yamlData, 0600); err != nil {
		return "", fmt.Errorf("writing config file to %s: %w", configPath, err)
	}

	return configPath, nil
}

func getDefaultModel(provider string) string {
	switch provider {
	case "gemini":
		return "gemini-1.5-flash"
	case "openai":
		return "gpt-4o-mini"
	case "claude":
		return "claude-3-5-sonnet-20241022"
	case "ollama":
		return "llama3"
	default:
		return ""
	}
}

func capitalizeFirst(s string) string {
	//Helper func used only in ollama sanitization
	// Takes s string as input to capitalize the strings first letter only, and lowercase the following strings if any.
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
