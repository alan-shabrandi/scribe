package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generatedCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message from staged changes",
	Long:  `Analyzes the current git diff and sends it to the LLM to construct a commit message.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("'generate' command triggered! Git diff reading and LLM integrations will be added in upcoming steps.")
	},
}

func init() {
	rootCmd.AddCommand(generatedCmd)
}