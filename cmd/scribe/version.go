package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the scribe version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Scribe CLI %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Go version: %s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
