package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/jaredreisinger/committed/internal/core"
)

var rootCmd = &cobra.Command{
	Use:   "committed [message-file] [source] [sha1]",
	Short: "Interactive conventional-commit assistant",
	Long:  "A Go TUI that helps format conventional commit messages and can integrate with git hooks.",
	Args:  cobra.RangeArgs(1, 3),

	RunE: func(cmd *cobra.Command, args []string) error {
		return core.Run(os.Args)
	},
}

func init() {
	// get the version from the build info
	info, ok := debug.ReadBuildInfo()
	if ok {
		rootCmd.Version = info.Main.Version
	}

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
