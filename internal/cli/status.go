package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	statusVerbose bool
	statusJSON    bool
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show workspace status overview",
	Long: `Show comprehensive workspace status including repos, worktrees, and dirty state.

This command provides a quick overview of your entire workspace: which repositories
are configured and cloned, which worktrees exist, and whether there's any uncommitted
work. It's designed to give you instant context when returning to work.

Examples:
  # Show workspace status
  fa status

  # Use short alias
  fa st

  # Show detailed status with file changes
  fa status -v

  # JSON output for AI agents
  fa status --json`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Flags
	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "Show detailed file-level status")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output in JSON format")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load workspace
	ws, err := workspace.Discover("")
	if err != nil {
		return err
	}

	// Collect workspace status
	status, err := ws.GetWorkspaceStatus(statusVerbose)
	if err != nil {
		return err
	}

	// Output based on format
	if statusJSON {
		return outputStatusJSON(status)
	}

	return outputStatusHuman(status, statusVerbose)
}

// outputStatusJSON outputs status in JSON format (US3)
func outputStatusJSON(status *workspace.WorkspaceStatus) error {
	output := map[string]interface{}{
		"workspace": map[string]interface{}{
			"name": status.WorkspaceName,
			"path": status.WorkspacePath,
		},
		"repos":     status.Repos,
		"worktrees": status.Worktrees,
		"summary": map[string]interface{}{
			"total_repos":             status.Summary.TotalRepos,
			"total_worktrees":         status.Summary.TotalWorktrees,
			"total_branches":          status.Summary.TotalBranches,
			"dirty_worktrees":         status.Summary.DirtyWorktrees,
			"has_uncommitted_changes": status.Summary.HasUncommittedChanges,
			"config_in_sync":          status.Summary.ConfigInSync,
			"repos_not_cloned":        status.Summary.ReposNotCloned,
			"repos_not_in_config":     status.Summary.ReposNotInConfig,
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// outputStatusHuman outputs status in human-readable format (US1, US2, US4, US5)
func outputStatusHuman(status *workspace.WorkspaceStatus, verbose bool) error {
	// Workspace header
	fmt.Printf("Workspace: %s\n", status.WorkspaceName)
	fmt.Printf("Path: %s\n\n", status.WorkspacePath)

	// Summary (US1)
	fmt.Printf("Summary:\n")
	fmt.Printf("  Repositories: %d\n", status.Summary.TotalRepos)
	fmt.Printf("  Worktrees: %d\n", status.Summary.TotalWorktrees)
	fmt.Printf("  Branches: %d\n", status.Summary.TotalBranches)

	// Uncommitted changes summary (US2)
	if status.Summary.HasUncommittedChanges {
		fmt.Printf("  Uncommitted changes: \033[33m%d worktree(s)\033[0m\n", status.Summary.DirtyWorktrees)
	} else if status.Summary.TotalWorktrees > 0 {
		fmt.Printf("  Status: \033[32m✓ All worktrees clean\033[0m\n")
	}

	// Config sync status (US4)
	if !status.Summary.ConfigInSync {
		if status.Summary.ReposNotCloned > 0 {
			fmt.Printf("  \033[33m⚠ %d repo(s) not cloned\033[0m\n", status.Summary.ReposNotCloned)
		}
	} else if status.Summary.TotalRepos > 0 {
		fmt.Printf("  Config: \033[32m✓ In sync\033[0m\n")
	}

	fmt.Println()

	// Repositories section (US1, US4)
	if len(status.Repos) > 0 {
		fmt.Println("Repositories:")
		for _, repo := range status.Repos {
			marker := "\033[32m✓\033[0m"
			statusText := ""
			if !repo.IsCloned {
				marker = "\033[33m✗\033[0m"
				statusText = " \033[33m[not cloned]\033[0m"
			}
			fmt.Printf("  %s %s%s\n", marker, repo.Name, statusText)
			if verbose {
				fmt.Printf("      URL: %s\n", repo.URL)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("No repositories configured")
		fmt.Println("Add repositories with: fa add <url> <name>")
		fmt.Println()
	}

	// Worktrees section (US1, US2, US5)
	if len(status.Worktrees) > 0 {
		fmt.Println("Worktrees:")

		// Group by branch
		branchMap := make(map[string][]workspace.WorktreeStatus)
		for _, wt := range status.Worktrees {
			branchMap[wt.Branch] = append(branchMap[wt.Branch], wt)
		}

		// Display each branch
		for branch, worktrees := range branchMap {
			fmt.Printf("\n  Branch: %s\n", branch)
			for _, wt := range worktrees {
				// Status indicator (US2)
				statusIndicator := ""
				switch wt.Status {
				case "modified":
					statusIndicator = " \033[33m[modified]\033[0m"
				case "untracked":
					statusIndicator = " \033[33m[untracked]\033[0m"
				case "conflict":
					statusIndicator = " \033[31m[conflict]\033[0m"
				}

				// Current marker (US1)
				currentMarker := " "
				if wt.IsCurrent {
					currentMarker = "*"
				}

				fmt.Printf("   %s %s%s\n", currentMarker, wt.Repo, statusIndicator)

				// Verbose mode - show files (US5)
				if verbose && len(wt.ModifiedFiles) > 0 {
					fmt.Printf("      Modified files:\n")
					for _, file := range wt.ModifiedFiles {
						fmt.Printf("        M %s\n", file)
					}
				}
				if verbose && len(wt.UntrackedFiles) > 0 {
					fmt.Printf("      Untracked files:\n")
					for _, file := range wt.UntrackedFiles {
						fmt.Printf("        ? %s\n", file)
					}
				}
			}
		}
		fmt.Println()
	} else if status.Summary.TotalRepos > 0 {
		fmt.Println("No worktrees found")
		fmt.Println("Create worktrees with: fa wt create <branch>")
		fmt.Println()
	}

	return nil
}
