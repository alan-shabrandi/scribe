package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/alan-shabrandi/scribe/internal/config"
	"github.com/alan-shabrandi/scribe/internal/git"
	"github.com/alan-shabrandi/scribe/internal/llm"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message from staged changes",
	Long:  `Analyzes current git diff, interacts with LLM, and allows user to confirm or edit the generated message.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("❌ Configuration Error: %v\n", err)
			os.Exit(1)
		}

		if cfg.APIKey == "" {
			fmt.Println("❌ Error: API key not set.")
			fmt.Println("Please add it to ~/.scribe.yaml or set SCRIBE_API_KEY environment variable.")
			os.Exit(1)
		}

		fmt.Println("🔍 Fetching staged git changes...")
		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🤖 Generating commit message using model '%s'...\n", cfg.Model)
		client := llm.NewClient(cfg.APIKey, cfg.Model)
		commitMessage, err := client.GenerateCommitMessage(diff)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✨ Proposed Commit Message:")
		fmt.Println("--------------------------------------------------")
		fmt.Println(commitMessage)
		fmt.Println("--------------------------------------------------\n")

		// Interactive menu options
		const (
			OptionAccept = "Accept & Commit"
			OptionEdit   = "Edit message"
			OptionCancel = "Cancel"
		)

		var selectedAction string
		prompt := &survey.Select{
			Message: "Would you like to use this commit message?",
			Options: []string{OptionAccept, OptionEdit, OptionCancel},
			Default: OptionAccept,
		}

		if err := survey.AskOne(prompt, &selectedAction); err != nil {
			fmt.Println("❌ Interaction cancelled.")
			os.Exit(1)
		}

		switch selectedAction {
		case OptionAccept:
			commitAndFinish(commitMessage)

		case OptionEdit:
			var editedMessage string
			editPrompt := &survey.Input{
				Message: "Edit commit message:",
				Default: commitMessage,
			}
			if err := survey.AskOne(editPrompt, &editedMessage); err != nil {
				fmt.Println("❌ Edit cancelled.")
				os.Exit(1)
			}
			commitAndFinish(editedMessage)

		case OptionCancel:
			fmt.Println("🚫 Commit cancelled by user.")
			os.Exit(0)
		}
	},
}

func commitAndFinish(msg string) {
	fmt.Println("🚀 Executing git commit...")
	if err := git.ExecuteCommit(msg); err != nil {
		fmt.Printf("❌ Failed to commit: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("🎉 Successfully committed staged changes!")
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
