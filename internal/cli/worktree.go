package cli

import (
	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:     "worktree",
	Aliases: []string{"wt"},
	Short:   "Manage git worktrees across all repositories",
	Long: `Manage git worktrees across all repositories in the workspace.

Worktrees allow you to have multiple branches checked out simultaneously
in different directories. This command creates, lists, switches, and removes
worktrees atomically across all repos in your workspace.`,
	Example: `  # Create a new worktree across all repos
  fa wt create feature-123

  # Create from a specific branch
  fa wt create hotfix-1 --from release-2.0

  # List all worktrees
  fa wt list

  # Switch to a different worktree
  fa wt switch feature-123

  # Remove a worktree
  fa wt remove feature-123`,
}

func init() {
	rootCmd.AddCommand(worktreeCmd)
}
