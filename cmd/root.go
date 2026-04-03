package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/internal/hook"
	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/tui"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "committed [message-file] [source] [sha1]",
	Short: "Interactive conventional-commit assistant",
	Long:  "A Go TUI that helps format conventional commit messages and can integrate with git hooks.",
	Args:  cobra.RangeArgs(1, 3),
	Run:   runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	// Build hook args matching hook.ParseHookArgs expectations: [program, message-file, source?, sha1?]
	hookArgs := make([]string, len(os.Args))
	copy(hookArgs, os.Args)

	hookCtx, err := hook.ParseHookArgs(hookArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing hook arguments: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	existingContent, err := hook.ReadMessageFile(hookCtx.MessageFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading commit message: %v\n", err)
		os.Exit(1)
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

	model := tui.NewModel(cfg, existingMsg)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	tuiModel, ok := finalModel.(tui.Model)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unexpected model type\n")
		os.Exit(1)
	}

	resultMsg, err := tuiModel.Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Form not completed: %v\n", err)
		os.Exit(1)
	}

	formatted := resultMsg.String()
	err = hook.WriteMessageFile(hookCtx.MessageFilePath, formatted)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing commit message: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func init() {
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Show version")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println("committed", version)
			os.Exit(0)
		}
	}
}
