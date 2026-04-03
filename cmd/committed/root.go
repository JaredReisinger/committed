package committed

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "committed",
	Short: "Interactive conventional-commit assistant",
	Long:  "A Go TUI that helps format conventional commit messages and can integrate with git hooks.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(hookCmd)
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Show version")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println("committed", version)
			os.Exit(0)
		}
	}
}
