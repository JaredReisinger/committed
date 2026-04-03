package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook [message-file] [source] [sha1]",
	Short: "Run as prepare-commit-msg hook",
	Args:  cobra.RangeArgs(1, 3),
	Run: func(cmd *cobra.Command, args []string) {
		messageFile := args[0]
		source := ""
		sha1 := ""
		if len(args) > 1 {
			source = args[1]
		}
		if len(args) > 2 {
			sha1 = args[2]
		}

		fmt.Fprintf(os.Stderr, "prepare-commit-msg hook invoked with file=%q source=%q sha1=%q\n", messageFile, source, sha1)

		// TODO: integrate with config, parser, TUI, and writer.
		os.Exit(0)
	},
}
