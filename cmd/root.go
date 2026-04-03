package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/internal/hook"
	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/tui"
)

var rootCmd = &cobra.Command{
	Use:   "committed [message-file] [source] [sha1]",
	Short: "Interactive conventional-commit assistant",
	Long:  "A Go TUI that helps format conventional commit messages and can integrate with git hooks.",
	Args:  cobra.RangeArgs(1, 3),
	RunE:  runRoot,
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

func runRoot(cmd *cobra.Command, args []string) error {
	// Build hook args matching hook.ParseHookArgs expectations: [program, message-file, source?, sha1?]
	hookArgs := make([]string, len(os.Args))
	copy(hookArgs, os.Args)

	hookCtx, err := hook.ParseHookArgs(hookArgs)
	if err != nil {
		return errors.WithMessage(err, "error parsing hook arguments")
	}

	cfg, err := config.LoadConfig(".")
	if err != nil {
		return errors.WithMessage(err, "error loading config")
	}

	existingContent, err := hook.ReadMessageFile(hookCtx.MessageFilePath)
	if err != nil {
		return errors.WithMessage(err, "error reading commit message")
	}

	var existingMsg *commit.Message
	if existingContent != "" {
		parsed, err := commit.ParseMessage(existingContent)
		if err != nil {
			existingMsg = &commit.Message{Description: "", Body: existingContent}
		} else {
			existingMsg = parsed
		}
	}

	// this seems like too much UI knowledge in the CLI

	model := tui.NewModel(cfg, existingMsg)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return errors.WithMessage(err, "error running TUI")
	}

	tuiModel, ok := finalModel.(tui.Model)
	if !ok {
		return errors.WithMessage(err, "unexpected model type")
	}

	resultMsg, err := tuiModel.Result()
	if err != nil {
		return errors.WithMessage(err, "form not completed")
	}

	formatted := resultMsg.String()
	err = hook.WriteMessageFile(hookCtx.MessageFilePath, formatted)
	if err != nil {
		return errors.WithMessage(err, "error writing commit message")
	}

	return nil
}
