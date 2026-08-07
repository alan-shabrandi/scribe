package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const configFileName = ".scribe.yaml"

type ConfigFile struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key,omitempty"`
	Model    string `yaml:"model"`
	Style    string `yaml:"style"`
}

var (
	initGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	initRed   = color.New(color.FgRed, color.Bold).SprintFunc()
	initCyan  = color.New(color.FgCyan, color.Bold).SprintFunc()
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration for scribe CLI",
	Long:  `Interactively guides you through setting up your LLM provider, API keys, default models, and commit message style in ~/.scribe.yaml.`,
	Run:   runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Printf("%s Welcome to Scribe Initialization!\n\n", initCyan("🚀"))

	cfg, err := gatherUserConfig()
	if err != nil {
		fmt.Printf("\n%s Initialization cancelled or failed: %v\n", initRed("❌"), err)
		os.Exit(1)
	}

	configPath, err := saveConfig(cfg)
	if err != nil {
		fmt.Printf("\n%s Error saving config: %v\n", initRed("❌"), err)
		os.Exit(1)
	}

	fmt.Printf("\n%s Configuration successfully saved to %s!\n", initGreen("🎉"), configPath)
	fmt.Println("You can now run 'scribe generate' or 'scribe' from any Git repository.")
}

func gatherUserConfig() (*ConfigFile, error) {
	var provider string
	if err := survey.AskOne(&survey.Select{
		Message: "Choose your LLM Provider:",
		Options: []string{"gemini", "openai", "ollama"},
		Default: "gemini",
	}, &provider); err != nil {
		return nil, err
	}

	defaultModel := getDefaultModel(provider)

	var model string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter default model name:",
		Default: defaultModel,
	}, &model); err != nil {
		return nil, err
	}

	var apiKey string
	if provider != "ollama" {
		if err := survey.AskOne(&survey.Password{
			Message: fmt.Sprintf("Enter your API Key for %s:", capitalizeFirst(provider)),
		}, &apiKey); err != nil {
			return nil, err
		}
	}

	var style string
	if err := survey.AskOne(&survey.Select{
		Message: "Choose default commit message style:",
		Options: []string{"conventional", "freeform"},
		Default: "conventional",
		Help:    "conventional: feat(scope): message | freeform: descriptive plain text",
	}, &style); err != nil {
		return nil, err
	}

	return &ConfigFile{
		Provider: provider,
		APIKey:   apiKey,
		Model:    model,
		Style:    style,
	}, nil
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
