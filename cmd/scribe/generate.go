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
			os.Exit(1)
		}

		fmt.Printf("%s Fetching staged git changes...\n", cyan("🔍"))

		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Printf("%s Error: %v\n", red("❌"), err)
			os.Exit(1)
		}

		branch, _ := git.GetCurrentBranch()
		ticketID := git.ExtractTicketID(branch)

		if ticketID != "" {
			fmt.Printf(
				"%s Detected Context: Branch '%s' (Ticket: %s)\n",
				cyan("📌"),
				branch,
				green(ticketID),
			)
		}

		provider, err := llm.NewProvider(
			cfg.Provider,
			cfg.APIKey,
			cfg.Model,
		)

		if err != nil {
			fmt.Printf("%s Config Error: %v\n", red("❌"), err)
			os.Exit(1)
		}

		diffChunks := git.SplitDiffByFile(diff)

		var candidates []string

		if len(diff) > LargeDiffThreshold && len(diffChunks) > 1 {

			fmt.Printf(
				"%s Large diff detected (%d chars across %d files). Activating multi-step chunk summarization...\n",
				yellow("⚡"),
				len(diff),
				len(diffChunks),
			)

			var fileSummaries []string

			for i, chunk := range diffChunks {

				s := spinner.New(
					spinner.CharSets[14],
					100*time.Millisecond,
				)

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
						"%s Warning: failed summarizing file %d: %v\n",
						yellow("⚠️"),
						i+1,
						err,
					)
					continue
				}

				fileSummaries = append(
					fileSummaries,
					summary,
				)
			}

			summaryPrompt := llm.BuildSummaryBasedPrompt(
				fileSummaries,
				cfg.Style,
				ticketID,
			)

			s := spinner.New(
				spinner.CharSets[14],
				100*time.Millisecond,
			)

			s.Suffix = " Generating commit message options..."
			s.Color("cyan", "bold")
			s.Start()

			candidates, err = provider.GenerateMultipleCommitMessages(
				summaryPrompt,
				cfg.Style,
				ticketID,
			)

			s.Stop()

		} else {

			s := spinner.New(
				spinner.CharSets[14],
				100*time.Millisecond,
			)

			s.Suffix = fmt.Sprintf(
				" Generating candidate commit messages via '%s'...",
				yellow(cfg.Provider),
			)

			s.Color("cyan", "bold")
			s.Start()

			candidates, err = provider.GenerateMultipleCommitMessages(
				diff,
				cfg.Style,
				ticketID,
			)

			s.Stop()
		}

		if err != nil {
			fmt.Printf(
				"%s Error generating options: %v\n",
				red("❌"),
				err,
			)
			os.Exit(1)
		}

		if len(candidates) == 0 {
			fmt.Printf(
				"%s No commit message candidates generated.\n",
				red("❌"),
			)
			os.Exit(1)
		}

		const (
			OptionEditManual = "✏️  Edit custom message in system editor"
			OptionCancel     = "🚫 Cancel"
		)

		selectOptions := append(
			candidates,
			OptionEditManual,
			OptionCancel,
		)

		var selectedOption string

		prompt := &survey.Select{
			Message: "Select a commit message option:",
			Options: selectOptions,
			Default: selectOptions[0],
		}

		if err := survey.AskOne(prompt, &selectedOption); err != nil {
			fmt.Printf("%s Interaction cancelled.\n", red("❌"))
			os.Exit(1)
		}

		switch selectedOption {

		case OptionCancel:

			fmt.Printf(
				"%s Commit cancelled by user.\n",
				yellow("🚫"),
			)

			os.Exit(0)

		case OptionEditManual:

			editedMessage, err := git.OpenInEditor(
				candidates[0],
			)

			if err != nil {
				fmt.Printf(
					"%s Edit failed: %v\n",
					red("❌"),
					err,
				)
				os.Exit(1)
			}

			commitAndFinish(editedMessage)

		default:

			var finalAction string

			actionPrompt := &survey.Select{
				Message: fmt.Sprintf(
					"Selected: \"%s\"\nWhat would you like to do?",
					green(selectedOption),
				),
				Options: []string{
					"Accept & Commit",
					"Edit in system editor",
					"Cancel",
				},
				Default: "Accept & Commit",
			}

			if err := survey.AskOne(
				actionPrompt,
				&finalAction,
			); err != nil {
				os.Exit(1)
			}

			switch finalAction {

			case "Accept & Commit":

				commitAndFinish(selectedOption)

			case "Edit in system editor":

				editedMessage, err := git.OpenInEditor(
					selectedOption,
				)

				if err != nil {
					fmt.Printf(
						"%s Edit failed: %v\n",
						red("❌"),
						err,
					)
					os.Exit(1)
				}

				commitAndFinish(editedMessage)

			default:

				fmt.Printf(
					"%s Commit cancelled.\n",
					yellow("🚫"),
				)
			}
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
