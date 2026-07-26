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

type ConfigFile struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key,omitempty"`
	Model    string `yaml:"model"`
	Style    string `yaml:"style"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration for scribe CLI",
	Long:  `Interactively guides you through setting up your LLM provider, API keys, default models, and commit message style in ~/.scribe.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		red := color.New(color.FgRed, color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()

		fmt.Printf("%s Welcome to Scribe Initialization!\n\n", cyan("🚀"))

		var provider string
		providerPrompt := &survey.Select{
			Message: "Choose your LLM Provider:",
			Options: []string{"gemini", "openai", "ollama"},
			Default: "gemini",
		}
		if err := survey.AskOne(providerPrompt, &provider); err != nil {
			fmt.Printf("%s Initialization cancelled.\n", red("❌"))
			os.Exit(1)
		}

		var defaultModel string
		switch provider {
		case "gemini":
			defaultModel = "gemini-1.5-flash"
		case "openai":
			defaultModel = "gpt-4o-mini"
		case "ollama":
			defaultModel = "llama3"
		}

		var model string
		modelPrompt := &survey.Input{
			Message: "Enter default model name:",
			Default: defaultModel,
		}
		if err := survey.AskOne(modelPrompt, &model); err != nil {
			os.Exit(1)
		}

		var apiKey string
		if provider != "ollama" {
			apiKeyPrompt := &survey.Password{
				Message: fmt.Sprintf("Enter your API Key for %s:", strings.Title(provider)),
			}
			if err := survey.AskOne(apiKeyPrompt, &apiKey); err != nil {
				os.Exit(1)
			}
		}

		var style string
		stylePrompt := &survey.Select{
			Message: "Choose default commit message style:",
			Options: []string{"conventional", "freeform"},
			Default: "conventional",
			Help:    "conventional: feat(scope): message | freeform: descriptive plain text",
		}
		if err := survey.AskOne(stylePrompt, &style); err != nil {
			os.Exit(1)
		}

		cfg := ConfigFile{
			Provider: provider,
			APIKey:   apiKey,
			Model:    model,
			Style:    style,
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("%s Error locating home directory: %v\n", red("❌"), err)
			os.Exit(1)
		}

		configPath := filepath.Join(homeDir, ".scribe.yaml")

		yamlData, err := yaml.Marshal(&cfg)
		if err != nil {
			fmt.Printf("%s Error generating YAML config: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, yamlData, 0600); err != nil {
			fmt.Printf("%s Error writing config file to %s: %v\n", red("❌"), configPath, err)
			os.Exit(1)
		}

		fmt.Printf("\n%s Configuration successfully saved to %s!\n", green("🎉"), configPath)
		fmt.Println("You can now run 'scribe generate' or 'scribe' from any Git repository.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
