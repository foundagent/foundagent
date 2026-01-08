package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/git"
)

// RemovalResult represents the result of a repo removal operation
type RemovalResult struct {
	RepoName          string `json:"repo_name"`
	RemovedFromConfig bool   `json:"removed_from_config"`
	BareCloneDeleted  bool   `json:"bare_clone_deleted"`
	WorktreesDeleted  int    `json:"worktrees_deleted"`
	ConfigOnly        bool   `json:"config_only"`
	Error             string `json:"error,omitempty"`
}

// RemoveRepo removes a repository from the workspace
func (w *Workspace) RemoveRepo(repoName string, force bool, configOnly bool) RemovalResult {
	result := RemovalResult{
		RepoName:   repoName,
		ConfigOnly: configOnly,
	}

	// Check if repo exists in state
	state, err := w.LoadState()
	if err != nil {
		result.Error = fmt.Sprintf("failed to load state: %v", err)
		return result
	}

	_, repoExists := state.Repositories[repoName]
	if !repoExists {
		result.Error = fmt.Sprintf("repo '%s' not found in workspace", repoName)
		return result
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		result.Error = fmt.Sprintf("failed to get current directory: %v", err)
		return result
	}

	// Check if CWD is inside any worktree of this repo
	worktreeBase := w.WorktreeBasePath(repoName)
	if strings.HasPrefix(cwd, worktreeBase) {
		result.Error = "cannot remove repo while inside its worktree - change directory first"
		return result
	}

	// If not config-only, check for dirty worktrees
	if !configOnly && !force {
		dirtyWorktrees, err := w.findDirtyWorktrees(repoName)
		if err != nil {
			result.Error = fmt.Sprintf("failed to check worktrees: %v", err)
			return result
		}

		if len(dirtyWorktrees) > 0 {
			result.Error = fmt.Sprintf("uncommitted changes in worktrees: %s. Use --force to remove anyway", strings.Join(dirtyWorktrees, ", "))
			return result
		}
	}

	// Remove from config
	err = w.removeRepoFromConfig(repoName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to remove from config: %v", err)
		return result
	}
	result.RemovedFromConfig = true

	// If config-only, skip file deletion
	if configOnly {
		// Still update workspace file
		err = w.removeRepoFromWorkspaceFile(repoName)
		if err != nil {
			result.Error = fmt.Sprintf("failed to update workspace file: %v", err)
			return result
		}
		return result
	}

	// Remove worktrees
	worktreesDeleted, err := w.removeAllWorktrees(repoName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to remove worktrees: %v", err)
		return result
	}
	result.WorktreesDeleted = worktreesDeleted

	// Remove bare clone
	bareRepoPath := w.BareRepoPath(repoName)
	err = os.RemoveAll(bareRepoPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to delete bare clone: %v", err)
		return result
	}
	result.BareCloneDeleted = true

	// Update workspace file
	err = w.removeRepoFromWorkspaceFile(repoName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to update workspace file: %v", err)
		return result
	}

	// Remove from state
	err = w.removeRepoFromState(repoName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to update state: %v", err)
		return result
	}

	return result
}

// findDirtyWorktrees finds all worktrees with uncommitted changes
func (w *Workspace) findDirtyWorktrees(repoName string) ([]string, error) {
	worktreeBase := w.WorktreeBasePath(repoName)

	// Check if worktree base exists
	if _, err := os.Stat(worktreeBase); os.IsNotExist(err) {
		return nil, nil
	}

	// List all worktrees
	entries, err := os.ReadDir(worktreeBase)
	if err != nil {
		return nil, err
	}

	var dirtyWorktrees []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		worktreePath := filepath.Join(worktreeBase, entry.Name())

		// Check for uncommitted changes
		hasChanges, err := git.HasUncommittedChanges(worktreePath)
		if err != nil {
			// Skip if not a valid git worktree
			continue
		}

		if hasChanges {
			dirtyWorktrees = append(dirtyWorktrees, entry.Name())
		}
	}

	return dirtyWorktrees, nil
}

// removeAllWorktrees removes all worktrees for a repo
func (w *Workspace) removeAllWorktrees(repoName string) (int, error) {
	worktreeBase := w.WorktreeBasePath(repoName)

	// Check if worktree base exists
	if _, err := os.Stat(worktreeBase); os.IsNotExist(err) {
		return 0, nil
	}

	// List all worktrees
	entries, err := os.ReadDir(worktreeBase)
	if err != nil {
		return 0, err
	}

	bareRepoPath := w.BareRepoPath(repoName)
	count := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		worktreePath := filepath.Join(worktreeBase, entry.Name())

		// Use git worktree remove
		cmd := exec.Command("git", "--git-dir="+bareRepoPath, "worktree", "remove", "--force", worktreePath)
		err := cmd.Run()
		if err != nil {
			// If git worktree remove fails, try manual removal
			err = os.RemoveAll(worktreePath)
			if err != nil {
				return count, fmt.Errorf("failed to remove worktree %s: %v", entry.Name(), err)
			}
		}
		count++
	}

	// Remove the worktree base directory
	err = os.RemoveAll(worktreeBase)
	if err != nil {
		return count, err
	}

	return count, nil
}

// removeRepoFromConfig removes a repo from .foundagent.yaml
func (w *Workspace) removeRepoFromConfig(repoName string) error {
	// Load using the config package
	cfg, err := w.loadFoundagentConfig()
	if err != nil {
		return err
	}

	// Find and remove repo
	newRepos := make([]config.RepoConfig, 0)
	for _, repo := range cfg.Repos {
		if repo.Name != repoName {
			newRepos = append(newRepos, repo)
		}
	}

	cfg.Repos = newRepos
	return w.saveFoundagentConfig(cfg)
}

// removeRepoFromState removes a repo from state.json
func (w *Workspace) removeRepoFromState(repoName string) error {
	state, err := w.LoadState()
	if err != nil {
		return err
	}

	if state.Repositories != nil {
		delete(state.Repositories, repoName)
	}

	return w.SaveState(state)
}

// removeRepoFromWorkspaceFile removes all worktree folders for a repo from the VS Code workspace
func (w *Workspace) removeRepoFromWorkspaceFile(repoName string) error {
	// Get all worktree paths for this repo
	worktreeBase := w.WorktreeBasePath(repoName)

	// Read workspace file
	workspaceData, err := w.LoadVSCodeWorkspace()
	if err != nil {
		return err
	}

	// Filter out folders that start with this repo's worktree base
	newFolders := make([]VSCodeFolder, 0)
	for _, folder := range workspaceData.Folders {
		// Check if folder is under this repo's worktree base
		folderPath := folder.Path
		if !filepath.IsAbs(folderPath) {
			folderPath = filepath.Join(w.Path, folderPath)
		}

		if !strings.HasPrefix(folderPath, worktreeBase) {
			newFolders = append(newFolders, folder)
		}
	}

	workspaceData.Folders = newFolders
	return w.SaveVSCodeWorkspace(workspaceData)
}

// loadFoundagentConfig loads the config using the config package
func (w *Workspace) loadFoundagentConfig() (*config.Config, error) {
	return config.Load(w.Path)
}

// saveFoundagentConfig saves the config using the config package
func (w *Workspace) saveFoundagentConfig(cfg *config.Config) error {
	return config.Save(w.Path, cfg)
}
