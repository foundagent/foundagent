package cli

import (
	"strings"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

// getWorktreeCompletions returns available worktree names for completion
// Returns empty list gracefully if not in a workspace
func getWorktreeCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Try to discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		// Not in a workspace - return empty gracefully
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Load state to get worktrees
	state, err := ws.LoadState()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Collect all worktree names
	var worktrees []string
	for _, repo := range state.Repositories {
		for _, wt := range repo.Worktrees {
			// Filter by prefix if provided
			if toComplete == "" || strings.HasPrefix(wt, toComplete) {
				worktrees = append(worktrees, wt)
			}
		}
	}

	return worktrees, cobra.ShellCompDirectiveNoFileComp
}

// getRepoCompletions returns available repository names for completion
// Returns empty list gracefully if not in a workspace
func getRepoCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Try to discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		// Not in a workspace - return empty gracefully
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Load state to get repos
	state, err := ws.LoadState()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Collect all repo names
	var repos []string
	for name := range state.Repositories {
		// Filter by prefix if provided
		if toComplete == "" || strings.HasPrefix(name, toComplete) {
			repos = append(repos, name)
		}
	}

	return repos, cobra.ShellCompDirectiveNoFileComp
}

// getBranchCompletions returns available branch names for completion
// Returns empty list gracefully if not in a workspace
func getBranchCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Try to discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		// Not in a workspace - return empty gracefully
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Load state to get branches from worktrees
	state, err := ws.LoadState()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Collect unique branch names
	branchSet := make(map[string]bool)
	for _, repo := range state.Repositories {
		for _, wt := range repo.Worktrees {
			// Worktree names often encode branch names
			// For now, just return the worktree names
			branchSet[wt] = true
		}
	}

	// Convert to slice and filter
	var branches []string
	for branch := range branchSet {
		if toComplete == "" || strings.HasPrefix(branch, toComplete) {
			branches = append(branches, branch)
		}
	}

	return branches, cobra.ShellCompDirectiveNoFileComp
}
