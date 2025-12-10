package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/errors"
	"github.com/foundagent/foundagent/internal/git"
	"github.com/foundagent/foundagent/internal/output"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	removeForce        bool
	removeDeleteBranch bool
	removeJSON         bool
)

var removeCmd = &cobra.Command{
	Use:     "remove <branch>",
	Aliases: []string{"rm"},
	Short:   "Remove worktrees across all repositories",
	Long: `Remove worktrees for a branch across ALL repositories in the workspace.

This command removes worktrees atomically - if safety checks fail, no worktrees
are removed. The command blocks removal if:
  - Any worktree has uncommitted changes (unless --force)
  - You're currently inside a worktree being removed
  - Trying to remove default branch (unless --force)

After removal, the VS Code workspace file is updated to remove those folders.`,
	Example: `  # Remove worktrees for a branch
  fa wt remove feature-123

  # Force remove despite uncommitted changes
  fa wt remove feature-123 --force

  # Remove worktrees AND delete branches
  fa wt remove feature-123 --delete-branch

  # JSON output for automation
  fa wt remove feature-123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

type removeResult struct {
	RepoName     string `json:"repo_name"`
	Branch       string `json:"branch"`
	WorktreePath string `json:"worktree_path"`
	Status       string `json:"status"` // removed, skipped, failed
	Reason       string `json:"reason,omitempty"`
	Error        string `json:"error,omitempty"`
}

type removeOutput struct {
	Branch         string         `json:"branch"`
	TotalRemoved   int            `json:"total_removed"`
	TotalSkipped   int            `json:"total_skipped"`
	TotalFailed    int            `json:"total_failed"`
	BranchesDeleted bool          `json:"branches_deleted"`
	Results        []removeResult `json:"results"`
}

func init() {
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Force removal despite uncommitted changes")
	removeCmd.Flags().BoolVar(&removeDeleteBranch, "delete-branch", false, "Delete branches after removing worktrees")
	removeCmd.Flags().BoolVar(&removeJSON, "json", false, "Output result as JSON")
	worktreeCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	targetBranch := args[0]

	// Discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		if removeJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Load config to get repos
	cfg, err := config.Load(ws.Path)
	if err != nil {
		if removeJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Check if workspace has repos
	if len(cfg.Repos) == 0 {
		err := errors.New(
			errors.ErrCodeInvalidConfig,
			"No repositories in workspace",
			"Add repositories with 'fa add <url>'",
		)
		if removeJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Find all worktrees for this branch
	worktreesToRemove, err := findWorktreesForBranch(ws, cfg, targetBranch)
	if err != nil {
		if removeJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	if len(worktreesToRemove) == 0 {
		err := errors.New(
			errors.ErrCodeWorktreeNotFound,
			fmt.Sprintf("No worktrees found for branch '%s'", targetBranch),
			"Run 'fa wt list' to see available worktrees",
		)
		if removeJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Pre-validation: safety checks
	if err := preValidateRemoval(ws, cfg, worktreesToRemove, targetBranch); err != nil {
		if removeJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Remove worktrees
	results := removeWorktreesParallel(ws, cfg, worktreesToRemove)

	// Count results
	removed := 0
	skipped := 0
	failed := 0
	for _, r := range results {
		switch r.Status {
		case "removed":
			removed++
		case "skipped":
			skipped++
		case "failed":
			failed++
		}
	}

	// Delete branches if requested and all removals succeeded
	branchesDeleted := false
	if removeDeleteBranch && failed == 0 {
		if err := deleteBranchesForRepos(ws, cfg, targetBranch, worktreesToRemove); err != nil {
			if removeJSON {
				output.PrintError(err)
			} else {
				output.PrintErrorMessage("Warning: Failed to delete branches: %v", err)
			}
		} else {
			branchesDeleted = true
		}
	}

	// Update VS Code workspace
	if removed > 0 {
		if err := removeWorktreeFoldersFromVSCode(ws, worktreesToRemove); err != nil {
			// Non-fatal warning
			if !removeJSON {
				output.PrintMessage("Warning: Failed to update VS Code workspace: %v", err)
			}
		}
	}

	// Output results
	if removeJSON {
		return printRemoveJSON(removeOutput{
			Branch:          targetBranch,
			TotalRemoved:    removed,
			TotalSkipped:    skipped,
			TotalFailed:     failed,
			BranchesDeleted: branchesDeleted,
			Results:         results,
		})
	}

	// Human-readable output
	if removed > 0 {
		output.PrintMessage("\n✓ Removed worktrees for branch '%s' in %d repository(ies)", targetBranch, removed)
		for _, r := range results {
			if r.Status == "removed" {
				output.PrintMessage("  ✓ %s: %s", r.RepoName, r.WorktreePath)
			}
		}
	}

	if skipped > 0 {
		output.PrintMessage("\n⊘ Skipped %d repository(ies)", skipped)
		for _, r := range results {
			if r.Status == "skipped" {
				output.PrintMessage("  - %s: %s", r.RepoName, r.Reason)
			}
		}
	}

	if failed > 0 {
		output.PrintMessage("\n✗ Failed to remove %d worktree(s)", failed)
		for _, r := range results {
			if r.Status == "failed" {
				output.PrintMessage("  ✗ %s: %s", r.RepoName, r.Error)
			}
		}
		return fmt.Errorf("failed to remove %d worktree(s)", failed)
	}

	if branchesDeleted {
		output.PrintMessage("\n✓ Deleted branch '%s' from all repositories", targetBranch)
	}

	output.PrintMessage("")
	return nil
}

type worktreeToRemove struct {
	RepoName     string
	RepoConfig   config.RepoConfig
	Branch       string
	WorktreePath string
	BareRepoPath string
}

func findWorktreesForBranch(ws *workspace.Workspace, cfg *config.Config, branch string) ([]worktreeToRemove, error) {
	var worktrees []worktreeToRemove

	for _, repo := range cfg.Repos {
		worktreePath := ws.WorktreePath(repo.Name, branch)
		
		// Check if worktree exists
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			continue
		}

		worktrees = append(worktrees, worktreeToRemove{
			RepoName:     repo.Name,
			RepoConfig:   repo,
			Branch:       branch,
			WorktreePath: worktreePath,
			BareRepoPath: ws.BareRepoPath(repo.Name),
		})
	}

	return worktrees, nil
}

func preValidateRemoval(ws *workspace.Workspace, cfg *config.Config, worktrees []worktreeToRemove, branch string) error {
	var validationErrors []string

	// Check 1: Is CWD inside any of the worktrees being removed?
	cwd, err := os.Getwd()
	if err == nil {
		for _, wt := range worktrees {
			if strings.HasPrefix(cwd, wt.WorktreePath) {
				return errors.New(
					errors.ErrCodeInvalidOperation,
					"Cannot remove worktree you're currently in",
					"Change directory outside the worktree: cd "+ws.Path,
				)
			}
		}
	}

	// Check 2: Are we removing default branch? (unless --force)
	if !removeForce {
		for _, wt := range worktrees {
			if wt.Branch == wt.RepoConfig.DefaultBranch {
				validationErrors = append(validationErrors,
					fmt.Sprintf("%s: attempting to remove default branch '%s'", wt.RepoName, wt.Branch))
			}
		}
	}

	// Check 3: Do any worktrees have uncommitted changes? (unless --force)
	if !removeForce {
		var dirtyWorktrees []string
		for _, wt := range worktrees {
			hasChanges, _ := git.HasUncommittedChanges(wt.WorktreePath)
			hasUntracked, _ := git.HasUntrackedFiles(wt.WorktreePath)
			
			if hasChanges || hasUntracked {
				dirtyWorktrees = append(dirtyWorktrees, fmt.Sprintf("%s: %s", wt.RepoName, wt.WorktreePath))
			}
		}
		
		if len(dirtyWorktrees) > 0 {
			validationErrors = append(validationErrors, "Worktrees with uncommitted changes:")
			validationErrors = append(validationErrors, dirtyWorktrees...)
		}
	}

	if len(validationErrors) > 0 {
		errorMsg := "Validation failed:\n  " + strings.Join(validationErrors, "\n  ")
		remediation := "Use --force to remove anyway, or commit/stash changes first"
		if removeForce {
			remediation = "Cannot remove default branch or worktree you're inside"
		}
		
		return errors.New(
			errors.ErrCodeInvalidOperation,
			errorMsg,
			remediation,
		)
	}

	return nil
}

func removeWorktreesParallel(ws *workspace.Workspace, cfg *config.Config, worktrees []worktreeToRemove) []removeResult {
	var wg sync.WaitGroup
	results := make([]removeResult, len(worktrees))

	for i, wt := range worktrees {
		wg.Add(1)
		go func(idx int, w worktreeToRemove) {
			defer wg.Done()
			
			result := removeResult{
				RepoName:     w.RepoName,
				Branch:       w.Branch,
				WorktreePath: w.WorktreePath,
			}

			// Remove using git worktree remove
			err := git.WorktreeRemove(w.BareRepoPath, w.WorktreePath, removeForce)
			if err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				results[idx] = result
				return
			}

			// Verify directory is gone (git should have removed it)
			if _, err := os.Stat(w.WorktreePath); err == nil {
				// Directory still exists, try to remove it manually
				if err := os.RemoveAll(w.WorktreePath); err != nil {
					result.Status = "failed"
					result.Error = fmt.Sprintf("failed to remove directory: %v", err)
					results[idx] = result
					return
				}
			}

			result.Status = "removed"
			results[idx] = result
		}(i, wt)
	}

	wg.Wait()
	return results
}

func deleteBranchesForRepos(ws *workspace.Workspace, cfg *config.Config, branch string, worktrees []worktreeToRemove) error {
	// Check if branches are merged before deletion (unless --force)
	if !removeForce {
		var unmergedRepos []string
		for _, wt := range worktrees {
			isMerged, err := git.IsBranchMerged(wt.BareRepoPath, branch, wt.RepoConfig.DefaultBranch)
			if err != nil {
				continue
			}
			if !isMerged {
				unmergedRepos = append(unmergedRepos, wt.RepoName)
			}
		}

		if len(unmergedRepos) > 0 {
			return errors.New(
				errors.ErrCodeInvalidOperation,
				fmt.Sprintf("Branch '%s' is not fully merged in: %s", branch, strings.Join(unmergedRepos, ", ")),
				"Use --force to delete anyway, or merge the branch first",
			)
		}
	}

	// Delete branches
	for _, wt := range worktrees {
		if err := git.DeleteBranch(wt.BareRepoPath, branch, removeForce); err != nil {
			return err
		}
	}

	return nil
}

func removeWorktreeFoldersFromVSCode(ws *workspace.Workspace, worktrees []worktreeToRemove) error {
	vscodeWS, err := ws.LoadVSCodeWorkspace()
	if err != nil {
		return err
	}

	// Build set of paths to remove
	pathsToRemove := make(map[string]bool)
	for _, wt := range worktrees {
		// Make path relative to workspace root
		relPath, err := filepath.Rel(ws.Path, wt.WorktreePath)
		if err != nil {
			continue
		}
		pathsToRemove[relPath] = true
	}

	// Filter out removed paths
	var newFolders []workspace.VSCodeFolder
	for _, folder := range vscodeWS.Folders {
		if !pathsToRemove[folder.Path] {
			newFolders = append(newFolders, folder)
		}
	}

	vscodeWS.Folders = newFolders
	return ws.SaveVSCodeWorkspace(vscodeWS)
}

func printRemoveJSON(output removeOutput) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
