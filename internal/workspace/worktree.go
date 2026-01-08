package workspace

import (
	"os"
	"path/filepath"
	"strings"
)

// WorktreeDetail contains information about a single worktree
type WorktreeDetail struct {
	Branch string
	Repo   string
	Path   string
}

// WorktreeExists checks if a worktree exists for a given repo and branch
func (w *Workspace) WorktreeExists(repoName, branch string) (bool, error) {
	worktreePath := w.WorktreePath(repoName, branch)
	info, err := os.Stat(worktreePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// GetWorktreesForRepo returns all worktree branches for a repository
func (w *Workspace) GetWorktreesForRepo(repoName string) ([]string, error) {
	worktreeBase := w.WorktreeBasePath(repoName)

	// Check if directory exists
	if _, err := os.Stat(worktreeBase); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(worktreeBase)
	if err != nil {
		return nil, err
	}

	worktrees := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			worktrees = append(worktrees, entry.Name())
		}
	}
	return worktrees, nil
}

// GetAllWorktrees returns a map of repo -> []branch for all worktrees in workspace
func (w *Workspace) GetAllWorktrees() (map[string][]string, error) {
	reposDir := filepath.Join(w.Path, ReposDir)

	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return make(map[string][]string), nil
	}

	entries, err := os.ReadDir(reposDir)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() {
			repoName := entry.Name()
			worktrees, err := w.GetWorktreesForRepo(repoName)
			if err != nil {
				continue
			}
			if len(worktrees) > 0 {
				result[repoName] = worktrees
			}
		}
	}
	return result, nil
}

// FindWorktree finds the worktree directory for a branch, checking current directory
func (w *Workspace) FindWorktree(branch string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Resolve symlinks and get absolute paths for reliable comparison
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}
	absCwd, err = filepath.EvalSymlinks(absCwd)
	if err != nil {
		// If EvalSymlinks fails, use the absolute path as-is
		absCwd, _ = filepath.Abs(cwd)
	}

	absWsPath, err := filepath.Abs(w.Path)
	if err != nil {
		return "", err
	}
	absWsPath, err = filepath.EvalSymlinks(absWsPath)
	if err != nil {
		// If EvalSymlinks fails, use the absolute path as-is
		absWsPath, _ = filepath.Abs(w.Path)
	}

	// Check if we're in a repo directory
	reposBase := filepath.Join(absWsPath, ReposDir)
	if !strings.HasPrefix(absCwd, reposBase+string(filepath.Separator)) && absCwd != reposBase {
		return "", nil
	}

	// Extract repo name from path (format: repos/<repo-name>/worktrees/<branch>/...)
	rel, err := filepath.Rel(reposBase, absCwd)
	if err != nil {
		return "", err
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) < 1 {
		return "", nil
	}

	repoName := parts[0]
	return filepath.Join(absWsPath, ReposDir, repoName, WorktreesDir, branch), nil
}

// GetWorktreesForRepo returns all worktrees for a repository
func GetWorktreesForRepo(workspaceRoot, repoName string) ([]WorktreeDetail, error) {
	worktreeBase := filepath.Join(workspaceRoot, ReposDir, repoName, WorktreesDir)

	// Check if directory exists
	if _, err := os.Stat(worktreeBase); os.IsNotExist(err) {
		return []WorktreeDetail{}, nil
	}

	entries, err := os.ReadDir(worktreeBase)
	if err != nil {
		return nil, err
	}

	worktrees := make([]WorktreeDetail, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			worktrees = append(worktrees, WorktreeDetail{
				Branch: entry.Name(),
				Repo:   repoName,
				Path:   filepath.Join(worktreeBase, entry.Name()),
			})
		}
	}
	return worktrees, nil
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
