package main

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/alan-shabrandi/scribe/internal/config"
	"github.com/alan-shabrandi/scribe/internal/git"
	"github.com/alan-shabrandi/scribe/internal/llm"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message from staged changes",
	Long:  `Analyzes current git diff, interacts with LLM, and allows user to confirm or edit the generated message.`,
	Run: func(cmd *cobra.Command, args []string) {
		red := color.New(color.FgRed, color.Bold).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("%s Configuration Error: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if cfg.APIKey == "" {
			fmt.Printf("%s Error: API key not set.\n", red("❌"))
			fmt.Println("Please add it to ~/.scribe.yaml or set SCRIBE_API_KEY environment variable.")
			os.Exit(1)
		}

		fmt.Printf("%s Fetching staged git changes...\n", cyan("🔍"))
		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Printf("%s Error: %v\n", red("❌"), err)
			os.Exit(1)
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Generating commit message using model '%s'...", yellow(cfg.Model))
		s.Color("cyan", "bold")
		s.Start()

		client := llm.NewClient(cfg.APIKey, cfg.Model)
		commitMessage, err := client.GenerateCommitMessage(diff)

		s.Stop()

		if err != nil {
			fmt.Printf("%s Error generating message: %v\n", red("❌"), err)
			os.Exit(1)
		}

		fmt.Printf("\n%s Proposed Commit Message:\n", green("✨"))
		fmt.Println(color.HiBlackString("--------------------------------------------------"))
		fmt.Println(color.CyanString(commitMessage))
		fmt.Println(color.HiBlackString("--------------------------------------------------\n"))

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
			fmt.Printf("%s Interaction cancelled.\n", red("❌"))
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
				fmt.Printf("%s Edit cancelled.\n", red("❌"))
				os.Exit(1)
			}
			commitAndFinish(editedMessage)

		case OptionCancel:
			fmt.Printf("%s Commit cancelled by user.\n", yellow("🚫"))
			os.Exit(0)
		}
	},
}

func commitAndFinish(msg string) {
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	red := color.New(color.FgRed, color.Bold).SprintFunc()

	fmt.Printf("%s Executing git commit...\n", color.CyanString("🚀"))
	if err := git.ExecuteCommit(msg); err != nil {
		fmt.Printf("%s Failed to commit: %v\n", red("❌"), err)
		os.Exit(1)
	}
	fmt.Printf("%s Successfully committed staged changes!\n", green("🎉"))
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
