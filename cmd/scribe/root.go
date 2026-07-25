package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "scribe",
	Short: "scribe is an AI-powered CLI for generating git commit messages",
	Long:  `A smart CLI tool written in Go that analyzes git staged changes and generates standardized commit messages using LLMs.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}