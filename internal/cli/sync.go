package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [branch]",
	Short: "Sync workspace with remotes",
	Long: `Sync all repos in the workspace with their remotes.

By default, syncs by fetching from all remotes in parallel. Use --pull to also 
fast-forward merge the current (or specified) branch, or --push to push local commits.

Examples:
  # Fetch all repos
  fa sync

  # Fetch and pull current branch
  fa sync --pull

  # Fetch and pull specific branch
  fa sync feature-123 --pull

  # Push all repos with unpushed commits
  fa sync --push

  # Stash uncommitted changes before pull
  fa sync --pull --stash`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSync,
}

var (
	syncPull    bool
	syncPush    bool
	syncStash   bool
	syncJSON    bool
	syncVerbose bool
)

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().BoolVar(&syncPull, "pull", false, "Fetch and pull (fast-forward merge)")
	syncCmd.Flags().BoolVar(&syncPush, "push", false, "Push local commits to remotes")
	syncCmd.Flags().BoolVar(&syncStash, "stash", false, "Stash uncommitted changes before pull")
	syncCmd.Flags().BoolVar(&syncJSON, "json", false, "Output results as JSON")
	syncCmd.Flags().BoolVarP(&syncVerbose, "verbose", "v", false, "Show detailed progress")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Detect workspace
	ws, err := workspace.Discover("")
	if err != nil {
		return err
	}

	// Determine target branch (from args or current)
	var targetBranch string
	if len(args) > 0 {
		targetBranch = args[0]
	}

	// Validate flags
	if syncPush && syncPull {
		return fmt.Errorf("cannot use --push and --pull together")
	}

	if syncStash && !syncPull {
		return fmt.Errorf("--stash requires --pull")
	}

	// Execute sync based on flags
	if syncPush {
		return runSyncPush(ws)
	} else if syncPull {
		// Use current branch if not specified
		if targetBranch == "" {
			state, err := ws.LoadState()
			if err != nil {
				return err
			}
			targetBranch = state.CurrentBranch
			if targetBranch == "" {
				targetBranch = "main" // Default fallback
			}
		}
		return runSyncPull(ws, targetBranch)
	} else {
		// Default: fetch only
		return runSyncFetch(ws)
	}
}

func runSyncFetch(ws *workspace.Workspace) error {
	if syncVerbose {
		fmt.Println("Fetching from all remotes...")
	}

	results, err := ws.SyncAllRepos(syncVerbose)
	if err != nil {
		return err
	}

	// Handle empty workspace
	if len(results) == 0 {
		if syncJSON {
			return outputSyncJSON(results)
		}
		fmt.Println("No repositories configured in workspace")
		fmt.Println("Add repositories with: fa add <url>")
		return nil
	}

	// Display results
	if syncJSON {
		return outputSyncJSON(results)
	}

	// Human output
	fmt.Println(workspace.FormatSyncResults(results, "fetch"))
	
	summary := workspace.CalculateSummary(results)
	fmt.Printf("\nSummary: %d synced, %d failed\n", summary.Synced+summary.Updated, summary.Failed)

	// Exit with error if any failed
	if summary.Failed > 0 {
		return fmt.Errorf("sync completed with %d failures", summary.Failed)
	}

	return nil
}

func runSyncPull(ws *workspace.Workspace, branch string) error {
	if syncVerbose {
		fmt.Printf("Syncing branch '%s' with --pull...\n", branch)
	}

	results, err := ws.PullAllWorktrees(branch, syncStash, syncVerbose)
	if err != nil {
		return err
	}

	// Handle empty workspace
	if len(results) == 0 {
		if syncJSON {
			return outputSyncJSON(results)
		}
		fmt.Println("No repositories configured in workspace")
		fmt.Println("Add repositories with: fa add <url>")
		return nil
	}

	// Display results
	if syncJSON {
		return outputSyncJSON(results)
	}

	// Human output
	fmt.Println(workspace.FormatSyncResults(results, "pull"))
	
	summary := workspace.CalculateSummary(results)
	fmt.Printf("\nSummary: %d updated, %d skipped, %d failed\n", summary.Updated, summary.Skipped, summary.Failed)

	if summary.Failed > 0 {
		return fmt.Errorf("sync completed with %d failures", summary.Failed)
	}

	return nil
}

func runSyncPush(ws *workspace.Workspace) error {
	if syncVerbose {
		fmt.Println("Pushing local commits...")
	}

	results, err := ws.PushAllRepos(syncVerbose)
	if err != nil {
		return err
	}

	// Handle empty workspace
	if len(results) == 0 {
		if syncJSON {
			return outputSyncJSON(results)
		}
		fmt.Println("No repositories configured in workspace")
		fmt.Println("Add repositories with: fa add <url>")
		return nil
	}

	// Display results
	if syncJSON {
		return outputSyncJSON(results)
	}

	// Human output
	fmt.Println(workspace.FormatSyncResults(results, "push"))
	
	summary := workspace.CalculateSummary(results)
	
	// Check if nothing to push
	if summary.Pushed == 0 && summary.Failed == 0 {
		fmt.Println("\nNothing to push")
		return nil
	}
	
	fmt.Printf("\nSummary: %d pushed, %d failed\n", summary.Pushed, summary.Failed)

	if summary.Failed > 0 {
		return fmt.Errorf("sync completed with %d failures", summary.Failed)
	}

	return nil
}

func outputSyncJSON(results []workspace.SyncResult) error {
	summary := workspace.CalculateSummary(results)
	
	output := struct {
		Repos   []workspace.SyncResult  `json:"repos"`
		Summary workspace.SyncSummary   `json:"summary"`
	}{
		Repos:   results,
		Summary: summary,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
