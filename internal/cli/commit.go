package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit [message]",
	Short: "Commit staged changes across all repos",
	Long: `Create coordinated commits across all repos in the workspace.

The commit message can be provided as a positional argument or via the -m flag.
Only repos with staged changes (in the current branch worktrees) receive commits.

Examples:
  # Commit with message as positional argument
  fa commit "Add user preferences feature"

  # Commit with -m flag (git muscle memory)
  fa commit -m "Add user preferences feature"

  # Stage all tracked changes and commit
  fa commit -a "Quick fix"

  # Commit specific repos only
  fa commit --repo api --repo lib "API changes"

  # Preview what would be committed
  fa commit --dry-run "Test commit"

  # Amend previous commits
  fa commit --amend "Updated message"

  # JSON output for automation
  fa commit --json "Automated commit"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCommit,
}

var (
	commitMessage       string
	commitAll           bool
	commitAmend         bool
	commitDryRun        bool
	commitRepos         []string
	commitJSON          bool
	commitVerbose       bool
	commitAllowDetached bool
)

func init() {
	rootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message")
	commitCmd.Flags().BoolVarP(&commitAll, "all", "a", false, "Stage all tracked modifications")
	commitCmd.Flags().BoolVar(&commitAmend, "amend", false, "Amend the previous commit")
	commitCmd.Flags().BoolVar(&commitDryRun, "dry-run", false, "Preview without committing")
	commitCmd.Flags().StringArrayVar(&commitRepos, "repo", nil, "Limit to specific repos (can be repeated)")
	commitCmd.Flags().BoolVar(&commitJSON, "json", false, "Output as JSON")
	commitCmd.Flags().BoolVarP(&commitVerbose, "verbose", "v", false, "Show detailed progress")
	commitCmd.Flags().BoolVar(&commitAllowDetached, "allow-detached", false, "Allow commits in detached HEAD")
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		return err
	}

	// Determine message: positional arg takes precedence over -m flag
	message := getCommitMessage(args)

	// Validate message (required unless --amend)
	if message == "" && !commitAmend {
		return fmt.Errorf("commit message cannot be empty\nProvide a message as argument or use -m flag")
	}

	// Build options
	opts := workspace.CommitOptions{
		Message:       message,
		All:           commitAll,
		Amend:         commitAmend,
		DryRun:        commitDryRun,
		Repos:         commitRepos,
		Verbose:       commitVerbose,
		AllowDetached: commitAllowDetached,
	}

	// Execute commit
	results, err := ws.CommitAllRepos(opts)
	if err != nil {
		return err
	}

	return handleCommitResults(results, message)
}

func getCommitMessage(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return commitMessage
}

func handleCommitResults(results []workspace.CommitResult, message string) error {
	// Handle empty workspace
	if len(results) == 0 {
		if commitJSON {
			return outputCommitJSON(results, message)
		}
		fmt.Println("No repositories configured in workspace")
		fmt.Println("Add repositories with: fa add <url>")
		return nil
	}

	// Calculate summary
	summary := workspace.CalculateCommitSummary(results)

	// Output results
	if commitJSON {
		return outputCommitJSON(results, message)
	}

	printCommitHumanOutput(results, summary)
	return checkCommitFailures(summary)
}

func printCommitHumanOutput(results []workspace.CommitResult, summary workspace.CommitSummary) {
	if commitDryRun {
		fmt.Println("DRY RUN - No commits will be created")
		fmt.Println()
	}

	if commitVerbose {
		fmt.Printf("Committing across %d repositories...\n\n", len(results))
	}

	fmt.Print(workspace.FormatCommitResults(results))

	// Summary line
	fmt.Println()
	if commitDryRun {
		fmt.Printf("Summary: %d repos would be committed, %d skipped\n", summary.Committed, summary.Skipped)
	} else {
		fmt.Printf("Summary: %d committed, %d skipped, %d failed\n", summary.Committed, summary.Skipped, summary.Failed)
	}

	// Check for "nothing to commit" across all repos
	if summary.Committed == 0 && summary.Failed == 0 && !commitDryRun {
		fmt.Println("\nNothing to commit")
		fmt.Println("Hint: Stage changes first or use -a to stage all tracked modifications")
	}
}

func checkCommitFailures(summary workspace.CommitSummary) error {
	if summary.Failed > 0 {
		return fmt.Errorf("commit completed with %d failures", summary.Failed)
	}
	return nil
}

func outputCommitJSON(results []workspace.CommitResult, message string) error {
	summary := workspace.CalculateCommitSummary(results)

	// Convert error to string for JSON
	reposOutput := make([]map[string]interface{}, len(results))
	for i, r := range results {
		repo := map[string]interface{}{
			"name":          r.RepoName,
			"status":        r.Status,
			"commit_sha":    nil,
			"files_changed": r.FilesChanged,
			"insertions":    r.Insertions,
			"deletions":     r.Deletions,
			"error":         nil,
		}
		if r.CommitSHA != "" {
			repo["commit_sha"] = r.CommitSHA
		}
		if r.ErrorMessage != "" {
			repo["error"] = r.ErrorMessage
		}
		reposOutput[i] = repo
	}

	output := map[string]interface{}{
		"repos":   reposOutput,
		"summary": summary,
		"message": message,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
