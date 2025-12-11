package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// VSCodeWorkspace represents a VS Code workspace file structure
type VSCodeWorkspace struct {
	Folders []VSCodeFolder `json:"folders"`
}

// VSCodeFolder represents a folder in a VS Code workspace
type VSCodeFolder struct {
	Path string `json:"path"`
}

// createVSCodeWorkspace creates the .code-workspace file
func (w *Workspace) createVSCodeWorkspace() error {
	workspace := VSCodeWorkspace{
		Folders: []VSCodeFolder{
			{Path: "."},
		},
	}

	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal VS Code workspace",
			"This is an internal error, please report it",
			err,
		)
	}

	workspacePath := w.VSCodeWorkspacePath()
	if err := os.WriteFile(workspacePath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write VS Code workspace file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}

// LoadVSCodeWorkspace loads the VS Code workspace file
func (w *Workspace) LoadVSCodeWorkspace() (*VSCodeWorkspace, error) {
	workspacePath := w.VSCodeWorkspacePath()

	data, err := os.ReadFile(workspacePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(
				errors.ErrCodeFileNotFound,
				"VS Code workspace file not found",
				"Initialize the workspace with 'fa init'",
				err,
			)
		}
		return nil, errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to read VS Code workspace file",
			"Check file permissions",
			err,
		)
	}

	var workspace VSCodeWorkspace
	if err := json.Unmarshal(data, &workspace); err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to parse VS Code workspace file",
			"The workspace file may be corrupted. Try 'fa init --force' to reinitialize",
			err,
		)
	}

	return &workspace, nil
}

// SaveVSCodeWorkspace saves the VS Code workspace file
func (w *Workspace) SaveVSCodeWorkspace(workspace *VSCodeWorkspace) error {
	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal VS Code workspace",
			"This is an internal error, please report it",
			err,
		)
	}

	workspacePath := w.VSCodeWorkspacePath()
	if err := os.WriteFile(workspacePath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write VS Code workspace file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}

// AddWorktreeFolder adds a worktree folder to the VS Code workspace
func (w *Workspace) AddWorktreeFolder(worktreePath string) error {
	workspace, err := w.LoadVSCodeWorkspace()
	if err != nil {
		return err
	}

	// Make path relative to workspace root
	relPath, err := filepath.Rel(w.Path, worktreePath)
	if err != nil {
		// If relative path fails, use absolute path
		relPath = worktreePath
	}

	// Check if folder already exists
	for _, folder := range workspace.Folders {
		if folder.Path == relPath {
			return nil // Already exists
		}
	}

	// Add new folder
	workspace.Folders = append(workspace.Folders, VSCodeFolder{Path: relPath})

	return w.SaveVSCodeWorkspace(workspace)
}

// AddWorktreeFolders adds multiple worktree folders to the VS Code workspace
func (w *Workspace) AddWorktreeFolders(worktreePaths []string) error {
	workspace, err := w.LoadVSCodeWorkspace()
	if err != nil {
		return err
	}

	// Convert to relative paths and check for duplicates
	existingPaths := make(map[string]bool)
	for _, folder := range workspace.Folders {
		existingPaths[folder.Path] = true
	}

	for _, worktreePath := range worktreePaths {
		// Make path relative to workspace root
		relPath, err := filepath.Rel(w.Path, worktreePath)
		if err != nil {
			// If relative path fails, use absolute path
			relPath = worktreePath
		}

		// Skip if already exists
		if existingPaths[relPath] {
			continue
		}

		// Add new folder
		workspace.Folders = append(workspace.Folders, VSCodeFolder{Path: relPath})
		existingPaths[relPath] = true
	}

	return w.SaveVSCodeWorkspace(workspace)
}

// RemoveWorktreeFoldersFromVSCode removes worktree folders from the VS Code workspace
func (w *Workspace) RemoveWorktreeFoldersFromVSCode(worktreePaths []string) error {
	workspace, err := w.LoadVSCodeWorkspace()
	if err != nil {
		return err
	}

	// Convert to relative paths for comparison
	relPaths := make(map[string]bool)
	for _, worktreePath := range worktreePaths {
		relPath, err := filepath.Rel(w.Path, worktreePath)
		if err != nil {
			// If relative path fails, use absolute path
			relPath = worktreePath
		}
		relPaths[relPath] = true
	}

	// Filter out folders to remove
	newFolders := make([]VSCodeFolder, 0)
	for _, folder := range workspace.Folders {
		if !relPaths[folder.Path] {
			newFolders = append(newFolders, folder)
		}
	}

	workspace.Folders = newFolders
	return w.SaveVSCodeWorkspace(workspace)
}

// GetCurrentBranchFromWorkspace detects the current branch from workspace file
// by looking at the worktree folders present
func (w *Workspace) GetCurrentBranchFromWorkspace() (string, error) {
	workspace, err := w.LoadVSCodeWorkspace()
	if err != nil {
		return "", err
	}

	// Look for worktree folders pattern: repos/worktrees/<repo>/<branch>
	worktreePrefix := filepath.Join(ReposDir, WorktreesDir) + string(filepath.Separator)

	for _, folder := range workspace.Folders {
		// Clean the path
		cleanPath := filepath.Clean(folder.Path)

		// Check if it's a worktree folder
		if !filepath.IsAbs(cleanPath) {
			// Convert to absolute for checking
			cleanPath = filepath.Join(w.Path, cleanPath)
		}

		// Get relative path from workspace root
		relPath, err := filepath.Rel(w.Path, cleanPath)
		if err != nil {
			continue
		}

		// Check if it starts with repos/worktrees/
		if !strings.HasPrefix(relPath, worktreePrefix) {
			continue
		}

		// Extract branch from path: repos/worktrees/<repo>/<branch>
		rel, err := filepath.Rel(worktreePrefix, relPath)
		if err != nil {
			continue
		}

		// Split path and get branch name
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) >= 2 {
			// Return the branch name (last component)
			return parts[1], nil
		}
	}

	return "", errors.New(
		errors.ErrCodeWorktreeNotFound,
		"No worktrees found in workspace file",
		"Initialize worktrees with 'fa wt create'",
	)
}

// GetAvailableBranches returns all branches that have worktrees
func (w *Workspace) GetAvailableBranches() ([]string, error) {
	allWorktrees, err := w.GetAllWorktrees()
	if err != nil {
		return nil, err
	}

	// Collect unique branches
	branchSet := make(map[string]bool)
	for _, branches := range allWorktrees {
		for _, branch := range branches {
			branchSet[branch] = true
		}
	}

	// Convert to slice
	branches := make([]string, 0, len(branchSet))
	for branch := range branchSet {
		branches = append(branches, branch)
	}

	return branches, nil
}

// ReplaceWorktreeFolders replaces all worktree folders in workspace with new branch
func (w *Workspace) ReplaceWorktreeFolders(targetBranch string) error {
	workspace, err := w.LoadVSCodeWorkspace()
	if err != nil {
		return err
	}

	// Get all repositories
	state, err := w.LoadState()
	if err != nil {
		return err
	}

	if len(state.Repositories) == 0 {
		return errors.New(
			errors.ErrCodeInvalidConfig,
			"No repositories configured in workspace",
			"Add repositories with 'fa add'",
		)
	}

	// Build new folder list
	newFolders := make([]VSCodeFolder, 0)

	// Keep non-worktree folders (like workspace root ".")
	worktreePrefix := filepath.Join(ReposDir, WorktreesDir) + string(filepath.Separator)
	for _, folder := range workspace.Folders {
		relPath, err := filepath.Rel(w.Path, folder.Path)
		if err != nil {
			// If relative path fails, check the path directly
			relPath = folder.Path
		}

		// Keep folders that aren't worktrees
		if !strings.HasPrefix(relPath, worktreePrefix) && relPath != filepath.Join(ReposDir, WorktreesDir) {
			newFolders = append(newFolders, folder)
		}
	}

	// Add new worktree folders for target branch
	for repoName := range state.Repositories {
		worktreePath := w.WorktreePath(repoName, targetBranch)

		// Check if worktree exists
		if _, err := os.Stat(worktreePath); err != nil {
			continue // Skip if worktree doesn't exist
		}

		// Make path relative
		relPath, err := filepath.Rel(w.Path, worktreePath)
		if err != nil {
			relPath = worktreePath
		}

		newFolders = append(newFolders, VSCodeFolder{Path: relPath})
	}

	workspace.Folders = newFolders
	return w.SaveVSCodeWorkspace(workspace)
}
