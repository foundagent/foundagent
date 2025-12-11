package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var repoRemoveCmd = &cobra.Command{
	Use:     "remove <repo>...",
	Aliases: []string{"rm"},
	Short:   "Remove repositories from workspace",
	ValidArgsFunction: getRepoCompletions,
	Long: `Remove one or more repositories from the workspace.

This command removes repositories completely: deletes bare clone, removes all worktrees,
updates config and workspace file. Blocks on uncommitted changes unless --force is used.

Examples:
  # Remove a repo
  fa remove api

  # Remove multiple repos
  fa remove api web

  # Force removal with uncommitted changes
  fa remove api --force

  # Remove from config but keep files
  fa remove api --config-only

  # JSON output
  fa remove api --json`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRepoRemove,
}

var (
	repoRemoveForce      bool
	repoRemoveConfigOnly bool
	repoRemoveJSON       bool
)

func init() {
	rootCmd.AddCommand(repoRemoveCmd)

	repoRemoveCmd.Flags().BoolVar(&repoRemoveForce, "force", false, "Force removal despite uncommitted changes")
	repoRemoveCmd.Flags().BoolVar(&repoRemoveConfigOnly, "config-only", false, "Remove from config but keep files")
	repoRemoveCmd.Flags().BoolVar(&repoRemoveJSON, "json", false, "Output results as JSON")
}

func runRepoRemove(cmd *cobra.Command, args []string) error {
	// Detect workspace
	ws, err := workspace.Discover("")
	if err != nil {
		return err
	}

	// Process each repo
	var results []workspace.RemovalResult
	for _, repoName := range args {
		result := ws.RemoveRepo(repoName, repoRemoveForce, repoRemoveConfigOnly)
		results = append(results, result)
	}

	// Output results
	if repoRemoveJSON {
		return outputRemovalJSON(results)
	}

	return outputRemovalHuman(results)
}

func outputRemovalJSON(results []workspace.RemovalResult) error {
	output := struct {
		Repos []workspace.RemovalResult `json:"repos"`
	}{
		Repos: results,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputRemovalHuman(results []workspace.RemovalResult) error {
	hasErrors := false

	for _, result := range results {
		if result.Error != "" {
			fmt.Fprintf(os.Stderr, "✗ %s: %s\n", result.RepoName, result.Error)
			hasErrors = true
		} else {
			fmt.Printf("✓ Removed %s\n", result.RepoName)
			if result.RemovedFromConfig {
				fmt.Println("  - Removed from config")
			}
			if result.BareCloneDeleted {
				fmt.Println("  - Deleted bare clone")
			}
			if result.WorktreesDeleted > 0 {
				fmt.Printf("  - Deleted %d worktree(s)\n", result.WorktreesDeleted)
			}
			if result.ConfigOnly {
				fmt.Println("  - Files kept (--config-only)")
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("some removals failed")
	}

	return nil
}
