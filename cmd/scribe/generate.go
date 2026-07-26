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

const LargeDiffThreshold = 6000

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

		provider, err := llm.NewProvider(cfg.Provider, cfg.APIKey, cfg.Model)
		if err != nil {
			fmt.Printf("%s Config Error: %v\n", red("❌"), err)
			os.Exit(1)
		}

		diffChunks := git.SplitDiffByFile(diff)

		var finalCommitMessage string

		if len(diff) > LargeDiffThreshold && len(diffChunks) > 1 {

			fmt.Printf(
				"%s Large diff detected (%d chars across %d files). Activating multi-step chunk summarization...\n",
				yellow("⚡"),
				len(diff),
				len(diffChunks),
			)

			var fileSummaries []string

			for i, chunk := range diffChunks {

				s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
				s.Suffix = fmt.Sprintf(
					" Summarizing file change %d/%d...",
					i+1,
					len(diffChunks),
				)
				s.Color("yellow", "bold")
				s.Start()

				summary, err := provider.SummarizeFile(chunk)

				s.Stop()

				if err != nil {
					fmt.Printf(
						"%s Warning: failed to summarize file chunk %d: %v\n",
						yellow("⚠️"),
						i+1,
						err,
					)
					continue
				}

				fileSummaries = append(fileSummaries, summary)
			}

			if len(fileSummaries) == 0 {
				fmt.Printf(
					"%s No file summaries generated. Falling back to standard generation...\n",
					yellow("⚠️"),
				)

				s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
				s.Suffix = fmt.Sprintf(
					" Generating commit message via provider '%s' [%s]...",
					yellow(cfg.Provider),
					yellow(cfg.Model),
				)
				s.Color("cyan", "bold")
				s.Start()

				finalCommitMessage, err = provider.GenerateCommitMessage(
					diff,
					cfg.Style,
				)

				s.Stop()

			} else {

				s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
				s.Suffix = " Synthesizing file summaries into final commit message..."
				s.Color("cyan", "bold")
				s.Start()

				summaryPrompt := llm.BuildSummaryBasedPrompt(
					fileSummaries,
					cfg.Style,
				)

				finalCommitMessage, err = provider.GenerateCommitMessage(
					summaryPrompt,
					cfg.Style,
				)

				s.Stop()
			}

		} else {

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = fmt.Sprintf(
				" Generating commit message via provider '%s' [%s]...",
				yellow(cfg.Provider),
				yellow(cfg.Model),
			)
			s.Color("cyan", "bold")
			s.Start()

			finalCommitMessage, err = provider.GenerateCommitMessage(
				diff,
				cfg.Style,
			)

			s.Stop()
		}

		if err != nil {
			fmt.Printf(
				"%s Error generating message: %v\n",
				red("❌"),
				err,
			)
			os.Exit(1)
		}

		fmt.Printf("\n%s Proposed Commit Message:\n", green("✨"))
		fmt.Println(color.HiBlackString("--------------------------------------------------"))
		fmt.Println(color.CyanString(finalCommitMessage))
		fmt.Println(color.HiBlackString("--------------------------------------------------\n"))

		const (
			OptionAccept = "Accept & Commit"
			OptionEdit   = "Edit message in system editor"
			OptionCancel = "Cancel"
		)

		var selectedAction string

		prompt := &survey.Select{
			Message: "Would you like to use this commit message?",
			Options: []string{
				OptionAccept,
				OptionEdit,
				OptionCancel,
			},
			Default: OptionAccept,
		}

		if err := survey.AskOne(prompt, &selectedAction); err != nil {
			fmt.Printf("%s Interaction cancelled.\n", red("❌"))
			os.Exit(1)
		}

		switch selectedAction {

		case OptionAccept:
			commitAndFinish(finalCommitMessage)

		case OptionEdit:

			editedMessage, err := git.OpenInEditor(finalCommitMessage)
			if err != nil {
				fmt.Printf(
					"%s Edit failed: %v\n",
					red("❌"),
					err,
				)
				os.Exit(1)
			}

			commitAndFinish(editedMessage)

		case OptionCancel:

			fmt.Printf(
				"%s Commit cancelled by user.\n",
				yellow("🚫"),
			)
			os.Exit(0)
		}
	},
}

func commitAndFinish(msg string) {

	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	red := color.New(color.FgRed, color.Bold).SprintFunc()

	fmt.Printf(
		"%s Executing git commit...\n",
		color.CyanString("🚀"),
	)

	if err := git.ExecuteCommit(msg); err != nil {
		fmt.Printf(
			"%s Failed to commit: %v\n",
			red("❌"),
			err,
		)
		os.Exit(1)
	}

	fmt.Printf(
		"%s Successfully committed staged changes!\n",
		green("🎉"),
	)
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
