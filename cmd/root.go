package cmd

import (
	"errors"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/jaredreisinger/asp"
	"github.com/jaredreisinger/committed/internal/core"
)

// See ../docs/to-hook-or-not-to-hook.md to see why this *shouldn't* really be
// set up as a git hook. Even so, we allow for the 3 possible hook args, which
// would be: `committed MESSAGE-FILE SOURCE SHA1`... we just subtly discourage
// this by not listing them in the help.

const (
	minArgs = 1 // always need filename
	maxArgs = 3 // max: filename source sha
)

var (
	rootCmd = &cobra.Command{
		Use:   "committed MESSAGE_FILE",
		Short: "Interactive conventional-commit assistant",
		Long: `
A text user interface that helps format conventional commit messages. It expects
to be set up as a repo's 'core.editor' so that Git will invoke it any time it
needs you to edit a message.
`,
		// Args: argCheck,
		RunE: run,
	}
)

type options struct {
	DryRun  bool   `asp.short:"n" asp.desc:"try the TUI without saving anything"`
	LogFile string `asp.short:"l" asp.desc:"file to write logging to"`
}

func init() {
	// get the version from the build info
	info, ok := debug.ReadBuildInfo()
	if ok {
		rootCmd.Version = info.Main.Version
	}

	err := asp.Attach(
		rootCmd,
		options{},
		asp.WithEnvPrefix("COMMITTED_"),
		asp.WithoutConfigFlag,
	)
	cobra.CheckErr(err)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// argCheck disabled because cobra.PersistentPreRunE is *not* called before the
// cobra.Command.Args check... so we can't depend on that to populate our
// options struct.
//
// func argCheck(cmd *cobra.Command, args []string) error {
// 	opts, err := asp.Get[options](cmd)
// 	if err != nil {
// 		return err
// 	}

// 	var cobraFn cobra.PositionalArgs

// 	if opts.DryRun {
// 		cobraFn = cobra.MaximumNArgs(maxArgs)
// 	} else {
// 		cobraFn = cobra.RangeArgs(minArgs, maxArgs)
// 	}

// 	return cobraFn(cmd, args)
// }

func run(cmd *cobra.Command, args []string) error {
	opts, err := asp.Get[options](cmd)
	if err != nil {
		return err
	}

	argCount := len(args)
	if !opts.DryRun && argCount <= 0 {
		return errors.New("MESSAGE_FILE argument required")
	}

	// If we get to actually running the command, we no longer want errors
	// to result in the usage message.
	cmd.SilenceUsage = true

	return core.Run(args, opts.DryRun, opts.LogFile)
}
