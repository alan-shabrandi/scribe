package main

import (
	"fmt"
	"os"

	"github.com/alan-shabrandi/scribe/internal/git"
	"github.com/spf13/cobra"
)

var generatedCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message from staged changes",
	Long:  `Analyzes the current git diff and sends it to the LLM to construct a commit message.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Fetching staged git changes...")
		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully fetched git diff:")
		fmt.Println("--------------------------------------------------")
		fmt.Println(diff)
		fmt.Println("--------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(generatedCmd)
}