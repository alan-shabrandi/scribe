package main

import (
	"fmt"
	"os"

	"github.com/alan-shabrandi/scribe/internal/git"
	"github.com/alan-shabrandi/scribe/internal/llm"
	"github.com/spf13/cobra"
)

var generatedCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message from staged changes",
	Long:  `Analyzes the current git diff and sends it to the LLM to construct a commit message.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			fmt.Println("Error: GEMINI_API_KEY environment variable is not set.")
			fmt.Println("Please set it using: export GEMINI_API_KEY='your-api-key'")
			os.Exit(1)
		}
		fmt.Println("Fetching staged git changes...")
		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Generating commit message using Gemini API...")
		client := llm.NewClient(apiKey)
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
	rootCmd.AddCommand(generatedCmd)
}