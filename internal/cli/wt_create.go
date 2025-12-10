package cli

import (
	"fmt"
	"sync"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/errors"
	"github.com/foundagent/foundagent/internal/git"
	"github.com/foundagent/foundagent/internal/output"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	createFrom  string
	createForce bool
	createJSON  bool
)

var createCmd = &cobra.Command{
	Use:   "create <branch>",
	Short: "Create a new worktree across all repositories",
	Long: `Create a new worktree with the specified branch name across ALL repositories
in the workspace.

This command creates a new branch and worktree in every repository, based on
each repo's default branch (or a branch specified with --from). The worktrees
are created atomically - if validation fails for any repo, no worktrees are created.

Worktrees are created at: repos/worktrees/<repo>/<branch>/

The VS Code workspace file is automatically updated to include the new worktree
directories.`,
	Example: `  # Create worktree from default branch in all repos
  fa wt create feature-123

  # Create worktree from specific branch
  fa wt create hotfix-1 --from release-2.0

  # Force recreate existing worktree
  fa wt create feature-123 --force

  # JSON output for automation
  fa wt create feature-123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createFrom, "from", "", "Source branch to create from (defaults to each repo's default branch)")
	createCmd.Flags().BoolVar(&createForce, "force", false, "Force recreate if worktree already exists")
	createCmd.Flags().BoolVar(&createJSON, "json", false, "Output result as JSON")
	worktreeCmd.AddCommand(createCmd)
}

type createResult struct {
	RepoName     string `json:"repo_name"`
	Branch       string `json:"branch"`
	SourceBranch string `json:"source_branch"`
	WorktreePath string `json:"worktree_path"`
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
}

func runCreate(cmd *cobra.Command, args []string) error {
	targetBranch := args[0]

	// Validate branch name
	if err := git.ValidateBranchName(targetBranch); err != nil {
		if createJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		if createJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Load config to get repos
	cfg, err := config.Load(ws.Path)
	if err != nil {
		if createJSON {
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
		if createJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Phase 1: Pre-validation (atomic all-or-nothing)
	if err := preValidateWorktreeCreate(ws, cfg, targetBranch, createFrom, createForce); err != nil {
		if createJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Phase 2: Create worktrees in parallel
	results := createWorktreesParallel(ws, cfg.Repos, targetBranch, createFrom, createForce)

	// Check for failures
	failed := 0
	for _, r := range results {
		if r.Status == "error" {
			failed++
		}
	}

	// Phase 3: Update VS Code workspace
	if failed == 0 {
		worktreePaths := make([]string, len(results))
		for i, r := range results {
			worktreePaths[i] = r.WorktreePath
		}
		if err := ws.AddWorktreeFolders(worktreePaths); err != nil {
			if !createJSON {
				output.PrintErrorMessage("Warning: Failed to update VS Code workspace: %v", err)
			}
		}
	}

	// Output results
	if createJSON {
		if len(results) == 1 {
			return output.PrintJSON(results[0])
		}
		return output.PrintJSON(results)
	}

	// Human-readable output
	for _, r := range results {
		if r.Status == "success" {
			output.PrintMessage("✓ Created worktree for %s: %s", r.RepoName, r.WorktreePath)
		} else {
			output.PrintErrorMessage("✗ Failed to create worktree for %s: %s", r.RepoName, r.Error)
		}
	}

	if failed > 0 {
		return fmt.Errorf("failed to create worktrees in %d repository(ies)", failed)
	}

	output.PrintMessage("")
	output.PrintMessage("✓ Created worktrees for branch '%s' in %d repository(ies)", targetBranch, len(results))
	return nil
}

func preValidateWorktreeCreate(ws *workspace.Workspace, cfg *config.Config, targetBranch, sourceBranch string, force bool) error {
	var validationErrors []string

	for _, repo := range cfg.Repos {
		bareRepoPath := ws.BareRepoPath(repo.Name)

		// Check if worktree already exists
		exists, err := ws.WorktreeExists(repo.Name, targetBranch)
		if err != nil {
			return err
		}

		if exists && !force {
			validationErrors = append(validationErrors, 
				fmt.Sprintf("%s: worktree already exists", repo.Name))
			continue
		}

		// Check if target branch already exists (without worktree)
		branchExists, err := git.BranchExists(bareRepoPath, targetBranch)
		if err != nil {
			return err
		}

		if branchExists && !force {
			validationErrors = append(validationErrors,
				fmt.Sprintf("%s: branch '%s' already exists", repo.Name, targetBranch))
			continue
		}

		// If --from specified, validate it exists in all repos
		if sourceBranch != "" {
			exists, err := git.BranchExists(bareRepoPath, sourceBranch)
			if err != nil {
				return err
			}
			if !exists {
				validationErrors = append(validationErrors,
					fmt.Sprintf("%s: source branch '%s' not found", repo.Name, sourceBranch))
			}
		}
	}

	if len(validationErrors) > 0 {
		msg := fmt.Sprintf("Validation failed:\n")
		for _, e := range validationErrors {
			msg += fmt.Sprintf("  - %s\n", e)
		}
		
		if sourceBranch != "" {
			msg += "\nUse --force to recreate existing worktrees"
		} else {
			msg += "\nUse 'fa wt switch <branch>' to switch to existing branch, or --force to recreate"
		}

		return errors.New(
			errors.ErrCodeWorktreeExists,
			msg,
			"",
		)
	}

	return nil
}

func createWorktreesParallel(ws *workspace.Workspace, repos []config.RepoConfig, targetBranch, sourceBranch string, force bool) []createResult {
	results := make([]createResult, len(repos))
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(index int, r config.RepoConfig) {
			defer wg.Done()
			results[index] = createWorktreeForRepo(ws, r, targetBranch, sourceBranch, force)
		}(i, repo)
	}

	wg.Wait()
	return results
}

func createWorktreeForRepo(ws *workspace.Workspace, repo config.RepoConfig, targetBranch, sourceBranch string, force bool) createResult {
	bareRepoPath := ws.BareRepoPath(repo.Name)
	worktreePath := ws.WorktreePath(repo.Name, targetBranch)

	// Determine source branch
	source := sourceBranch
	if source == "" {
		// Use repo's default branch
		source = repo.DefaultBranch
		if source == "" {
			// Fallback to detecting default branch
			detected, err := git.GetDefaultBranch(bareRepoPath)
			if err != nil {
				return createResult{
					RepoName: repo.Name,
					Branch:   targetBranch,
					Status:   "error",
					Error:    fmt.Sprintf("Failed to detect default branch: %v", err),
				}
			}
			source = detected
		}
	}

	// If force and worktree exists, remove it
	if force {
		exists, _ := ws.WorktreeExists(repo.Name, targetBranch)
		if exists {
			// Remove worktree
			if err := git.WorktreeRemove(bareRepoPath, worktreePath, true); err != nil {
				return createResult{
					RepoName: repo.Name,
					Branch:   targetBranch,
					Status:   "error",
					Error:    fmt.Sprintf("Failed to remove existing worktree: %v", err),
				}
			}

			// Delete branch if it exists
			branchExists, _ := git.BranchExists(bareRepoPath, targetBranch)
			if branchExists {
				if err := git.DeleteBranch(bareRepoPath, targetBranch, true); err != nil {
					return createResult{
						RepoName: repo.Name,
						Branch:   targetBranch,
						Status:   "error",
						Error:    fmt.Sprintf("Failed to delete existing branch: %v", err),
					}
				}
			}
		}
	}

	// Create worktree with new branch
	if err := git.WorktreeAddNew(bareRepoPath, worktreePath, targetBranch, source); err != nil {
		return createResult{
			RepoName:     repo.Name,
			Branch:       targetBranch,
			SourceBranch: source,
			Status:       "error",
			Error:        err.Error(),
		}
	}

	return createResult{
		RepoName:     repo.Name,
		Branch:       targetBranch,
		SourceBranch: source,
		WorktreePath: worktreePath,
		Status:       "success",
	}
}
