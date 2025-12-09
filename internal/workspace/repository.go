package workspace

import (
	"path/filepath"
	"time"
)

// Repository represents a repository in the workspace
type Repository struct {
	Name          string    `json:"name" yaml:"name"`
	URL           string    `json:"url" yaml:"url"`
	DefaultBranch string    `json:"default_branch,omitempty" yaml:"default_branch,omitempty"`
	BareRepoPath  string    `json:"bare_repo_path,omitempty" yaml:"-"`
	AddedAt       time.Time `json:"added_at,omitempty" yaml:"-"`
	Worktrees     []string  `json:"worktrees,omitempty" yaml:"-"`
}

// BareRepoPath returns the path to the bare repository
func (w *Workspace) BareRepoPath(repoName string) string {
	return filepath.Join(w.Path, ReposDir, BareDir, repoName+".git")
}

// WorktreeBasePath returns the base path for worktrees of a repository
func (w *Workspace) WorktreeBasePath(repoName string) string {
	return filepath.Join(w.Path, ReposDir, WorktreesDir, repoName)
}

// WorktreePath returns the path to a specific worktree
func (w *Workspace) WorktreePath(repoName, branch string) string {
	return filepath.Join(w.WorktreeBasePath(repoName), branch)
}

// AddRepository registers a repository in the workspace
func (w *Workspace) AddRepository(repo *Repository) error {
	// Load current state
	state, err := w.LoadState()
	if err != nil {
		return err
	}

	// Add repo to state
	if state.Repositories == nil {
		state.Repositories = make(map[string]*Repository)
	}
	state.Repositories[repo.Name] = repo

	return w.SaveState(state)
}

// GetRepository retrieves a repository from state
func (w *Workspace) GetRepository(name string) (*Repository, error) {
	state, err := w.LoadState()
	if err != nil {
		return nil, err
	}

	if repo, ok := state.Repositories[name]; ok {
		return repo, nil
	}

	return nil, nil
}

// HasRepository checks if a repository exists in the workspace
func (w *Workspace) HasRepository(name string) (bool, error) {
	repo, err := w.GetRepository(name)
	if err != nil {
		return false, err
	}
	return repo != nil, nil
}
