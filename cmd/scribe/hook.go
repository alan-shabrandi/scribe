package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	gitDirName   = ".git"
	hooksDirName = "hooks"
	hookFileName = "prepare-commit-msg"
)

const hookContent = `#!/bin/sh
# Scribe Git Hook
# Automatically trigger scribe generate if commit message source is empty or message file

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2

# Only auto-generate if user didn't provide a -m message or merge template
if [ -z "$COMMIT_SOURCE" ] || [ "$COMMIT_SOURCE" = "message" ] || [ "$COMMIT_SOURCE" = "template" ]; then
    if [ ! -t 0 ]; then
        if ( : < /dev/tty ) 2>/dev/null; then
            exec < /dev/tty
        else
            exit 0
        fi
    fi

    scribe generate --hook-mode "$COMMIT_MSG_FILE"
fi
`

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage Git hooks integration for automatic commit generation",
	Long:  `Install or uninstall Scribe as a Git prepare-commit-msg hook in your current repository.`,
}

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install prepare-commit-msg hook in current Git repository",
	Run: func(cmd *cobra.Command, args []string) {
		hookPath, err := getHookPath()
		if err != nil {
			fmt.Printf("%s %v\n", red("❌"), err)
			os.Exit(1)
		}

		hooksDir := filepath.Dir(hookPath)
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			fmt.Printf("%s Failed to create hooks directory: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			fmt.Printf("%s Failed to install Git hook: %v\n", red("❌"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Successfully installed Scribe prepare-commit-msg hook in %s!\n", green("🎉"), hookPath)
		fmt.Println("Now running 'git commit' will automatically invoke Scribe.")
	},
}

var hookUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall prepare-commit-msg hook from current Git repository",
	Run: func(cmd *cobra.Command, args []string) {
		hookPath, err := getHookPath()
		if err != nil {
			fmt.Printf("%s %v\n", red("❌"), err)
			os.Exit(1)
		}

		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			fmt.Printf("%s Hook is not currently installed in this repository.\n", yellow("⚠️"))
			return
		}

		if err := os.Remove(hookPath); err != nil {
			fmt.Printf("%s Failed to remove Git hook: %v\n", red("❌"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Successfully uninstalled Scribe prepare-commit-msg hook.\n", green("🗑️"))
	},
}

func init() {
	hookCmd.AddCommand(hookInstallCmd)
	hookCmd.AddCommand(hookUninstallCmd)
	rootCmd.AddCommand(hookCmd)
}

func getHookPath() (string, error) {
	if _, err := os.Stat(gitDirName); os.IsNotExist(err) {
		return "", fmt.Errorf("Error: Not a git repository (or .git directory not found in the current directory)")
	}
	return filepath.Join(gitDirName, hooksDirName, hookFileName), nil
}
