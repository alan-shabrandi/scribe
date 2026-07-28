package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/alan-shabrandi/scribe/internal/cache"
	"github.com/alan-shabrandi/scribe/internal/config"
	"github.com/alan-shabrandi/scribe/internal/git"
	"github.com/alan-shabrandi/scribe/internal/llm"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	LargeDiffThreshold = 6000
	spinnerCharSet     = 14
	spinnerSpeed       = 100 * time.Millisecond
)

var (
	red    = color.New(color.FgRed, color.Bold).SprintFunc()
	green  = color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan, color.Bold).SprintFunc()
)

var (
	hookMsgFilePath string
	jsonOutput      bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message from staged changes",
	Long:  `Analyzes current git diff, interacts with LLM, and allows user to confirm or edit the generated message.`,
	Run:   runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&hookMsgFilePath, "hook-mode", "", "Write generated message directly to Git hook msg file")
	generateCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output generated commit messages in JSON format")
	rootCmd.AddCommand(generateCmd)
}

func printInfo(format string, a ...interface{}) {
	if !jsonOutput {
		fmt.Printf(format, a...)
	}
}

func printErr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func runGenerate(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		printErr("%s Configuration Error: %v\n", red("❌"), err)
		os.Exit(1)
	}

	if cfg.APIKey == "" && cfg.Provider != "ollama" {
		printErr("%s Error: API key not set.\n", red("❌"))
		os.Exit(1)
	}

	printInfo("%s Fetching staged git changes...\n", cyan("🔍"))
	diff, err := git.GetStagedDiff()
	if err != nil {
		printErr("%s Error: %v\n", red("❌"), err)
		os.Exit(1)
	}

	if len(diff) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			printInfo("%s No staged changes detected. Use 'git add' first.\n", yellow("⚠️"))
		}
		os.Exit(0)
	}

	branch, _ := git.GetCurrentBranch()
	ticketID := git.ExtractTicketID(branch)
	if ticketID != "" {
		printInfo("%s Detected Context: Branch '%s' (Ticket: %s)\n", cyan("📌"), branch, green(ticketID))
	}

	diffHash := cache.ComputeHash(diff)

	candidates, foundInCache := cache.GetCachedCandidates(diffHash)
	if foundInCache {
		printInfo("%s Found valid cached responses for unchanged diff!\n", green("⚡"))
	} else {
		candidates = fetchCandidatesFromLLM(cmd, cfg, diff, ticketID)
		if len(candidates) > 0 {
			_ = cache.SaveCandidates(diffHash, candidates)
		}
	}

	if len(candidates) == 0 {
		printErr("%s No commit message candidates generated.\n", red("❌"))
		os.Exit(1)
	}

	if jsonOutput {
		output, err := json.Marshal(candidates)
		if err != nil {
			printErr("%s Error parsing JSON: %v\n", red("❌"), err)
			os.Exit(1)
		}
		fmt.Println(string(output))
		return
	}

	handleUserSelection(candidates)
}

func fetchCandidatesFromLLM(cmd *cobra.Command, cfg *config.Config, diff, ticketID string) []string {
	provider, err := llm.NewProvider(cfg.Provider, cfg.APIKey, cfg.Model)
	if err != nil {
		printErr("%s Config Error: %v\n", red("❌"), err)
		os.Exit(1)
	}

	ctx := cmd.Context()
	diffChunks := git.SplitDiffByFile(diff)

	var candidates []string

	if len(diff) > LargeDiffThreshold && len(diffChunks) > 1 {
		printInfo("%s Large diff detected (%d chars across %d files). Activating multi-step chunk summarization...\n", yellow("⚡"), len(diff), len(diffChunks))

		var fileSummaries []string
		for i, chunk := range diffChunks {
			var s *spinner.Spinner
			if !jsonOutput {
				s = startSpinner(fmt.Sprintf("Summarizing file change %d/%d...", i+1, len(diffChunks)), "yellow")
			}

			summary, err := provider.SummarizeFile(ctx, chunk)

			if !jsonOutput && s != nil {
				s.Stop()
			}

			if err != nil {
				printInfo("%s Warning: failed summarizing file %d: %v\n", yellow("⚠️"), i+1, err)
				continue
			}
			fileSummaries = append(fileSummaries, summary)
		}

		summaryPrompt := llm.BuildSummaryBasedPrompt(fileSummaries, cfg.Style, ticketID)

		var s *spinner.Spinner
		if !jsonOutput {
			s = startSpinner("Generating commit message options...", "cyan")
		}

		candidates, err = provider.GenerateMultipleCommitMessages(ctx, summaryPrompt, cfg.Style, ticketID)

		if !jsonOutput && s != nil {
			s.Stop()
		}
	} else {
		var s *spinner.Spinner
		if !jsonOutput {
			s = startSpinner(fmt.Sprintf("Generating candidate commit messages via '%s'...", yellow(cfg.Provider)), "cyan")
		}

		candidates, err = provider.GenerateMultipleCommitMessages(ctx, diff, cfg.Style, ticketID)

		if !jsonOutput && s != nil {
			s.Stop()
		}
	}

	if err != nil {
		printErr("%s Error generating options: %v\n", red("❌"), err)
		os.Exit(1)
	}

	return candidates
}

func handleUserSelection(candidates []string) {
	const (
		OptionEditManual = "✏️  Edit custom message in system editor"
		OptionCancel     = "🚫 Cancel"
	)

	selectOptions := append(candidates, OptionEditManual, OptionCancel)
	var selectedOption string

	prompt := &survey.Select{
		Message: "Select a commit message option:",
		Options: selectOptions,
		Default: selectOptions[0],
	}

	var tty *os.File
	var err error
	if runtime.GOOS == "windows" {
		tty, err = os.Open("CONIN$")
	} else {
		tty, err = os.Open("/dev/tty")
	}

	var stdioOpt survey.AskOpt
	if err == nil {
		defer tty.Close()
		stdioOpt = survey.WithStdio(tty, os.Stdout, os.Stderr)
	} else {
		stdioOpt = survey.WithStdio(os.Stdin, os.Stdout, os.Stderr)
	}
	// -----------------------------------------------

	if err := survey.AskOne(prompt, &selectedOption, stdioOpt); err != nil {
		fmt.Printf("%s Interaction cancelled.\n", red("❌"))
		os.Exit(1)
	}

	switch selectedOption {
	case OptionCancel:
		fmt.Printf("%s Commit cancelled by user.\n", yellow("🚫"))
		os.Exit(0)

	case OptionEditManual:
		editedMessage, err := git.OpenInEditor(candidates[0])
		if err != nil {
			fmt.Printf("%s Edit failed: %v\n", red("❌"), err)
			os.Exit(1)
		}
		commitAndFinish(editedMessage)

	default:
		var finalAction string
		actionPrompt := &survey.Select{
			Message: fmt.Sprintf("Selected: \"%s\"\nWhat would you like to do?", green(selectedOption)),
			Options: []string{"Accept & Commit", "Edit in system editor", "Cancel"},
			Default: "Accept & Commit",
		}

		if err := survey.AskOne(actionPrompt, &finalAction, stdioOpt); err != nil {
			os.Exit(1)
		}

		switch finalAction {
		case "Accept & Commit":
			commitAndFinish(selectedOption)

		case "Edit in system editor":
			editedMessage, err := git.OpenInEditor(selectedOption)
			if err != nil {
				fmt.Printf("%s Edit failed: %v\n", red("❌"), err)
				os.Exit(1)
			}
			commitAndFinish(editedMessage)

		default:
			fmt.Printf("%s Commit cancelled.\n", yellow("🚫"))
		}
	}
}

func commitAndFinish(msg string) {
	fmt.Printf("%s Executing git commit...\n", cyan("🚀"))
	if err := git.ExecuteCommit(msg); err != nil {
		fmt.Printf("%s Failed to commit: %v\n", red("❌"), err)
		os.Exit(1)
	}

	fmt.Printf("%s Successfully committed staged changes!\n", green("🎉"))
}

func startSpinner(suffix, colorName string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[spinnerCharSet], spinnerSpeed)
	s.Suffix = " " + suffix
	s.Color(colorName, "bold")
	s.Start()
	return s
}
