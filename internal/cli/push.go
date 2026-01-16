package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push unpushed commits across all repos",
	Long: `Push all repos in the workspace that have unpushed commits.

Only repos with commits ahead of their upstream (in the current branch worktrees) 
are pushed. Repos already up-to-date are skipped.

Examples:
  # Push all repos with unpushed commits
  fa push

  # Push specific repos only
  fa push --repo api --repo lib

  # Preview what would be pushed
  fa push --dry-run

  # JSON output for automation
  fa push --json

  # Force push (dangerous, requires confirmation)
  fa push --force`,
	Args: cobra.NoArgs,
	RunE: runPush,
}

var (
	pushDryRun  bool
	pushRepos   []string
	pushJSON    bool
	pushVerbose bool
	pushForce   bool
)

func init() {
	rootCmd.AddCommand(pushCmd)

	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false, "Preview without pushing")
	pushCmd.Flags().StringArrayVar(&pushRepos, "repo", nil, "Limit to specific repos (can be repeated)")
	pushCmd.Flags().BoolVar(&pushJSON, "json", false, "Output as JSON")
	pushCmd.Flags().BoolVarP(&pushVerbose, "verbose", "v", false, "Show detailed progress")
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "Force push (dangerous)")
}

func runPush(cmd *cobra.Command, args []string) error {
	// Discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		return err
	}

	// Force push requires confirmation (or fails in JSON mode)
	if err := confirmForcePush(); err != nil {
		return err
	}

	// Build options
	opts := workspace.PushOptions{
		DryRun:  pushDryRun,
		Repos:   pushRepos,
		Verbose: pushVerbose,
		Force:   pushForce,
	}

	// Execute push
	results, err := ws.PushAllReposNew(opts)
	if err != nil {
		return err
	}

	return handlePushResults(results)
}

// confirmForcePush handles force push confirmation. Returns nil if confirmed or not needed.
func confirmForcePush() error {
	if !pushForce || pushDryRun {
		return nil
	}
	if pushJSON {
		return fmt.Errorf("force push requires confirmation - cannot use with --json")
	}
	fmt.Print("WARNING: Force push will overwrite remote history. Continue? [y/N] ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		fmt.Println("Aborted.")
		return errAborted
	}
	if response != "y" && response != "Y" {
		fmt.Println("Aborted.")
		return errAborted
	}
	return nil
}

// errAborted is a sentinel error for user-aborted operations
var errAborted = fmt.Errorf("")

func handlePushResults(results []workspace.PushResult) error {
	// Handle empty workspace
	if len(results) == 0 {
		if pushJSON {
			return outputPushJSON(results)
		}
		fmt.Println("No repositories configured in workspace")
		fmt.Println("Add repositories with: fa add <url>")
		return nil
	}

	// Calculate summary
	summary := workspace.CalculatePushSummary(results)

	// Output results
	if pushJSON {
		return outputPushJSON(results)
	}

	printPushHumanOutput(results, summary)
	return checkPushFailures(summary)
}

func printPushHumanOutput(results []workspace.PushResult, summary workspace.PushSummary) {
	if pushDryRun {
		fmt.Println("DRY RUN - No pushes will be executed")
		fmt.Println()
	}

	if pushVerbose {
		fmt.Printf("Pushing %d repositories...\n\n", len(results))
	}

	fmt.Print(workspace.FormatPushResults(results))

	// Summary line
	fmt.Println()
	if pushDryRun {
		fmt.Printf("Summary: %d repos would be pushed, %d already up-to-date\n", summary.Pushed, summary.Skipped)
	} else {
		fmt.Printf("Summary: %d pushed, %d skipped, %d failed\n", summary.Pushed, summary.Skipped, summary.Failed)
	}

	// Check for "nothing to push"
	if summary.Pushed == 0 && summary.Failed == 0 && !pushDryRun {
		fmt.Println("\nNothing to push")
	}
}

func checkPushFailures(summary workspace.PushSummary) error {
	if summary.Failed > 0 {
		fmt.Println("\nHint: Run 'fa sync --pull' to fetch and merge remote changes, then retry push")
		return fmt.Errorf("push completed with %d failures", summary.Failed)
	}
	return nil
}

func outputPushJSON(results []workspace.PushResult) error {
	summary := workspace.CalculatePushSummary(results)

	// Convert for JSON output
	reposOutput := make([]map[string]interface{}, len(results))
	for i, r := range results {
		repo := map[string]interface{}{
			"name":           r.RepoName,
			"status":         r.Status,
			"refs_pushed":    r.RefsPushed,
			"commits_pushed": r.CommitsPushed,
			"error":          nil,
		}
		if len(r.RefsPushed) == 0 {
			repo["refs_pushed"] = []string{}
		}
		if r.ErrorMessage != "" {
			repo["error"] = r.ErrorMessage
		}
		reposOutput[i] = repo
	}

	output := map[string]interface{}{
		"repos":   reposOutput,
		"summary": summary,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
