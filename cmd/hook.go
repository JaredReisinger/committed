package cmd

import (
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook [message-file] [source] [sha1]",
	Short: "Run as prepare-commit-msg hook",
	Args:  cobra.RangeArgs(1, 3),
	Run:   runHook,
}
