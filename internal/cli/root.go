package cli

import (
	"fmt"

	"github.com/foundagent/foundagent/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fa",
	Short: "Foundagent - Git worktree workspace manager",
	Long: `Foundagent is a CLI tool for managing multi-repository workspaces
using git worktrees and VS Code integration.`,
}

var showVersion bool

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Show version information")

	// Override RunE to handle --version flag
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.String())
			return nil
		}
		return cmd.Help()
	}
}
