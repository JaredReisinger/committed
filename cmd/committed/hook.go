package main

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

var hookCmd = &cobra.Command{
	Use:   "hook [message-file] [source] [sha1]",
	Short: "Run as prepare-commit-msg hook",
	Args:  cobra.RangeArgs(1, 3),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse hook arguments
		hookCtx, err := hook.ParseHookArgs(os.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing hook arguments: %v\n", err)
			os.Exit(1)
		}

		// Load configuration
		cfg, err := config.LoadConfig(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Read existing commit message
		existingContent, err := hook.ReadMessageFile(hookCtx.MessageFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading commit message: %v\n", err)
			os.Exit(1)
		}

		// Parse existing message if present
		var existingMsg *commit.Message
		if existingContent != "" {
			parsed, err := commit.ParseMessage(existingContent)
			if err != nil {
				// If parsing fails, treat as plain text in details field
				existingMsg = &commit.Message{
					Description: "",
					Body:        existingContent,
				}
			} else {
				existingMsg = parsed
			}
		}

		// Run TUI
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

		// Get the result
		resultMsg, err := tuiModel.Result()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Form not completed: %v\n", err)
			os.Exit(1)
		}

		// Format and write the message
		formatted := resultMsg.String()
		err = hook.WriteMessageFile(hookCtx.MessageFilePath, formatted)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing commit message: %v\n", err)
			os.Exit(1)
		}

		// Success
		os.Exit(0)
	},
}
