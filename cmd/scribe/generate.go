package main

import (
	"fmt"
	"os"

	"github.com/alan-shabrandi/scribe/internal/config"
	"github.com/alan-shabrandi/scribe/internal/git"
	"github.com/alan-shabrandi/scribe/internal/llm"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message from staged changes",
	Long:  `Analyzes the current git diff and sends it to the LLM to construct a commit message.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Configuration Error: %v\n", err)
			os.Exit(1)
		}

		if cfg.APIKey == "" {
			fmt.Println("Error: API key not set.")
			fmt.Println("Please add it to ~/.scribe.yaml or set SCRIBE_API_KEY environment variable.")
			os.Exit(1)
		}

		fmt.Println("Fetching staged git changes...")
		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Generating commit message using model '%s'...\n", cfg.Model)
		client := llm.NewClient(cfg.APIKey, cfg.Model)
		commitMessage, err := client.GenerateCommitMessage(diff)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n Proposed Commit Message:")
		fmt.Println("--------------------------------------------------")
		fmt.Println(commitMessage)
		fmt.Println("--------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
