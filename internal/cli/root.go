package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fa",
	Short: "Foundagent - Git worktree workspace manager",
	Long: `Foundagent is a CLI tool for managing multi-repository workspaces
using git worktrees and VS Code integration.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be added here
}
