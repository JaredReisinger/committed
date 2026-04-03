package cmd

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/jaredreisinger/committed/internal/core"
)

const (
	minArgs = 1 // always need filename
	maxArgs = 3 // max: filename source sha
)

var (
	dryRun bool

	rootCmd = &cobra.Command{
		Use:   "committed [message-file] [source] [sha1]",
		Short: "Interactive conventional-commit assistant",
		Long: `
A text user interface that helps format conventional commit messages and can
integrate with git hooks.
`,
		Args: ifElse(&dryRun, cobra.MaximumNArgs(maxArgs), cobra.RangeArgs(minArgs, maxArgs)),

		RunE: func(cmd *cobra.Command, args []string) error {
			// If we get to actually running the command, we no longer want errors
			// to result in the usage message. Also,
			cmd.SilenceUsage = true
			// cmd.SilenceErrors = true

			return core.Run(args, dryRun)
		},
	}
)

func init() {
	// get the version from the build info
	info, ok := debug.ReadBuildInfo()
	if ok {
		rootCmd.Version = info.Main.Version
	}

	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "test the UI without saving anything")
}

func ifElse(test *bool, trueArgs cobra.PositionalArgs, falseArgs cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if test != nil && *test {
			return trueArgs(cmd, args)
		}
		return falseArgs(cmd, args)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// if rootCmd.SilenceErrors {
		// 	fmt.Fprintln(os.Stderr, err)
		// }
		os.Exit(1)
	}
}
