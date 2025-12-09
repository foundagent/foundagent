package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/git"
	"github.com/foundagent/foundagent/internal/output"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	addForce bool
	addJSON  bool
)

var addCmd = &cobra.Command{
	Use:   "add [url] [name] [url2] [url3] ...",
	Short: "Add repositories to the workspace",
	Long: `Add repositories to the workspace by cloning them as bare clones
and creating worktrees for the default branch.

The repository is cloned to repos/.bare/<name>.git/ and a worktree
for the default branch is created at repos/worktrees/<name>/<branch>/.

Multiple repositories can be added in parallel by providing multiple URLs.
An optional custom name can be provided after the URL.

If no URLs are provided, repositories from .foundagent.yaml will be cloned
to match the configuration (reconciliation mode).`,
	Example: `  # Add a single repository
  fa add git@github.com:org/my-repo.git

  # Add with custom name
  fa add git@github.com:org/my-repo.git api-service

  # Add multiple repositories
  fa add git@github.com:org/repo1.git git@github.com:org/repo2.git

  # Sync workspace to match config
  fa add

  # Add with JSON output
  fa add git@github.com:org/my-repo.git --json

  # Force re-clone existing repository
  fa add git@github.com:org/my-repo.git --force`,
	Args: cobra.MinimumNArgs(0),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVar(&addForce, "force", false, "Force re-clone if repository already exists")
	addCmd.Flags().BoolVar(&addJSON, "json", false, "Output result as JSON")
	rootCmd.AddCommand(addCmd)
}

type addResult struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	BareRepoPath string `json:"bare_repo_path,omitempty"`
	WorktreePath string `json:"worktree_path,omitempty"`
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
	Skipped      bool   `json:"skipped,omitempty"`
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Discover workspace
	ws, err := workspace.Discover("")
	if err != nil {
		if addJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// If no URLs provided, sync from config (reconciliation mode)
	if len(args) == 0 {
		return runReconcile(ws)
	}

	// Parse arguments - support both "url name" and "url url url" patterns
	repos := parseAddArgs(args)

	// Add repositories
	results := addRepositories(ws, repos)

	// Output results
	if addJSON {
		if len(results) == 1 {
			return output.PrintJSON(results[0])
		}
		return output.PrintJSON(results)
	}

	// Human-readable output
	success := 0
	skipped := 0
	failed := 0

	for _, result := range results {
		if result.Status == "success" {
			success++
			if result.Skipped {
				skipped++
				output.PrintMessage("⊘ Repository '%s' already exists (skipped)", result.Name)
			} else {
				output.PrintMessage("✓ Added repository '%s'", result.Name)
				output.PrintMessage("  Bare clone: %s", result.BareRepoPath)
				output.PrintMessage("  Worktree:   %s", result.WorktreePath)
			}
		} else {
			failed++
			output.PrintErrorMessage("✗ Failed to add repository '%s': %s", result.Name, result.Error)
		}
	}

	if len(results) > 1 {
		output.PrintMessage("")
		output.PrintMessage("Summary: %d succeeded, %d skipped, %d failed", success-skipped, skipped, failed)
	}

	if failed > 0 {
		return fmt.Errorf("failed to add %d repository(ies)", failed)
	}

	return nil
}

func runReconcile(ws *workspace.Workspace) error {
	// Reconcile config with state
	result, err := workspace.Reconcile(ws)
	if err != nil {
		if addJSON {
			output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// If nothing to clone, report and exit
	if len(result.ReposToClone) == 0 {
		if addJSON {
			return output.PrintSuccess(map[string]interface{}{
				"message":        "All repositories are up-to-date",
				"repos_to_clone": 0,
				"repos_up_to_date": len(result.ReposUpToDate),
				"repos_stale":    len(result.ReposStale),
			})
		}
		output.PrintMessage("✓ All repositories are up-to-date")
		if len(result.ReposStale) > 0 {
			workspace.PrintReconcileResult(result)
		}
		return nil
	}

	// Clone missing repositories
	if !addJSON {
		output.PrintMessage("Cloning %d repository(ies) from config...", len(result.ReposToClone))
	}

	repos := make([]repoToAdd, len(result.ReposToClone))
	for i, r := range result.ReposToClone {
		repos[i] = repoToAdd{URL: r.URL, Name: r.Name}
	}

	results := addRepositories(ws, repos)

	// Output results
	if addJSON {
		return output.PrintJSON(map[string]interface{}{
			"repos_cloned": results,
			"repos_stale":  result.ReposStale,
		})
	}

	// Human-readable output
	success := 0
	failed := 0
	for _, r := range results {
		if r.Status == "success" && !r.Skipped {
			success++
		} else if r.Status == "error" {
			failed++
		}
	}

	output.PrintMessage("")
	output.PrintMessage("✓ Cloned %d repository(ies)", success)
	if failed > 0 {
		output.PrintMessage("✗ Failed to clone %d repository(ies)", failed)
	}

	if len(result.ReposStale) > 0 {
		workspace.PrintReconcileResult(result)
	}

	return nil
}

type repoToAdd struct {
	URL  string
	Name string
}

func parseAddArgs(args []string) []repoToAdd {
	var repos []repoToAdd

	i := 0
	for i < len(args) {
		url := args[i]

		// Check if next arg is a custom name (not a URL)
		var name string
		if i+1 < len(args) {
			nextArg := args[i+1]
			if err := git.ValidateURL(nextArg); err != nil {
				// Next arg is not a URL, treat it as a custom name
				name = nextArg
				i += 2
			} else {
				// Next arg is a URL
				i++
			}
		} else {
			i++
		}

		repos = append(repos, repoToAdd{URL: url, Name: name})
	}

	return repos
}

func addRepositories(ws *workspace.Workspace, repos []repoToAdd) []addResult {
	if len(repos) == 1 {
		// Single repository - no parallelization
		return []addResult{addRepository(ws, repos[0])}
	}

	// Multiple repositories - add in parallel
	results := make([]addResult, len(repos))
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(index int, r repoToAdd) {
			defer wg.Done()
			results[index] = addRepository(ws, r)
		}(i, repo)
	}

	wg.Wait()
	return results
}

func addRepository(ws *workspace.Workspace, repo repoToAdd) addResult {
	// Validate URL
	if err := git.ValidateURL(repo.URL); err != nil {
		return addResult{
			URL:    repo.URL,
			Status: "error",
			Error:  err.Error(),
		}
	}

	// Infer name if not provided
	name := repo.Name
	if name == "" {
		var err error
		name, err = git.InferName(repo.URL)
		if err != nil {
			return addResult{
				URL:    repo.URL,
				Status: "error",
				Error:  err.Error(),
			}
		}
	}

	// Check if repository already exists
	hasRepo, err := ws.HasRepository(name)
	if err != nil {
		return addResult{
			Name:   name,
			URL:    repo.URL,
			Status: "error",
			Error:  err.Error(),
		}
	}

	if hasRepo && !addForce {
		bareRepoPath := ws.BareRepoPath(name)
		return addResult{
			Name:         name,
			URL:          repo.URL,
			BareRepoPath: bareRepoPath,
			Status:       "success",
			Skipped:      true,
		}
	}

	// Create bare repository path
	bareRepoPath := ws.BareRepoPath(name)

	// Remove existing if force
	if hasRepo && addForce {
		if err := os.RemoveAll(bareRepoPath); err != nil {
			return addResult{
				Name:   name,
				URL:    repo.URL,
				Status: "error",
				Error:  fmt.Sprintf("Failed to remove existing repository: %v", err),
			}
		}
	}

	// Clone bare repository
	if !addJSON {
		output.PrintMessage("Cloning %s...", name)
	}

	if err := git.CloneBare(repo.URL, bareRepoPath, !addJSON); err != nil {
		return addResult{
			Name:   name,
			URL:    repo.URL,
			Status: "error",
			Error:  err.Error(),
		}
	}

	// Get default branch
	defaultBranch, err := git.GetDefaultBranch(bareRepoPath)
	if err != nil {
		// Clean up on failure
		os.RemoveAll(bareRepoPath)
		return addResult{
			Name:   name,
			URL:    repo.URL,
			Status: "error",
			Error:  fmt.Sprintf("Failed to determine default branch: %v", err),
		}
	}

	// Create worktree for default branch
	worktreePath := ws.WorktreePath(name, defaultBranch)
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		// Clean up on failure
		os.RemoveAll(bareRepoPath)
		return addResult{
			Name:   name,
			URL:    repo.URL,
			Status: "error",
			Error:  fmt.Sprintf("Failed to create worktree directory: %v", err),
		}
	}

	if err := git.WorktreeAdd(git.WorktreeAddOptions{
		BareRepoPath: bareRepoPath,
		WorktreePath: worktreePath,
		Branch:       defaultBranch,
	}); err != nil {
		// Clean up on failure
		os.RemoveAll(bareRepoPath)
		return addResult{
			Name:   name,
			URL:    repo.URL,
			Status: "error",
			Error:  err.Error(),
		}
	}

	// Register repository in workspace
	repository := &workspace.Repository{
		Name:          name,
		URL:           repo.URL,
		DefaultBranch: defaultBranch,
		BareRepoPath:  bareRepoPath,
		AddedAt:       time.Now(),
		Worktrees:     []string{defaultBranch},
	}

	if err := ws.AddRepository(repository); err != nil {
		// Clean up on failure
		os.RemoveAll(bareRepoPath)
		os.RemoveAll(worktreePath)
		return addResult{
			Name:   name,
			URL:    repo.URL,
			Status: "error",
			Error:  fmt.Sprintf("Failed to register repository: %v", err),
		}
	}

	// Update config file with new repository
	cfg, err := config.Load(ws.Path)
	if err != nil {
		// Config load failed, but repo is already added - just warn
		if !addJSON {
			output.PrintErrorMessage("Warning: Failed to update config: %v", err)
		}
	} else {
		config.AddRepo(cfg, repo.URL, name, defaultBranch)
		if err := config.Save(ws.Path, cfg); err != nil {
			// Config save failed, but repo is already added - just warn
			if !addJSON {
				output.PrintErrorMessage("Warning: Failed to save config: %v", err)
			}
		}
	}

	// Update VS Code workspace
	if err := ws.AddWorktreeFolder(worktreePath); err != nil {
		// Don't fail the whole operation if VS Code update fails
		if !addJSON {
			output.PrintErrorMessage("Warning: Failed to update VS Code workspace: %v", err)
		}
	}

	return addResult{
		Name:         name,
		URL:          repo.URL,
		BareRepoPath: bareRepoPath,
		WorktreePath: worktreePath,
		Status:       "success",
	}
}
