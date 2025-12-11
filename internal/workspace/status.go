package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/foundagent/foundagent/internal/git"
)

// WorkspaceStatus represents the complete status of a workspace
type WorkspaceStatus struct {
	WorkspaceName string
	WorkspacePath string
	Repos         []RepoStatus
	Worktrees     []WorktreeStatus
	Summary       StatusSummary
}

// RepoStatus represents the status of a single repository
type RepoStatus struct {
	Name      string
	URL       string
	IsCloned  bool
	InConfig  bool
	ClonePath string
}

// WorktreeStatus represents the status of a single worktree
type WorktreeStatus struct {
	Branch         string
	Repo           string
	Path           string
	Status         string // clean, modified, untracked, conflict
	IsCurrent      bool
	ModifiedFiles  []string
	UntrackedFiles []string
}

// StatusSummary provides aggregate counts and flags
type StatusSummary struct {
	TotalRepos            int
	TotalWorktrees        int
	TotalBranches         int
	DirtyWorktrees        int
	HasUncommittedChanges bool
	ConfigInSync          bool
	ReposNotCloned        int
	ReposNotInConfig      int
}

// GetWorkspaceStatus collects complete workspace status
func (w *Workspace) GetWorkspaceStatus(verbose bool) (*WorkspaceStatus, error) {
	// Load config and state
	state, err := w.LoadState()
	if err != nil {
		return nil, err
	}

	// Get repo statuses
	repoStatuses := w.getRepoStatuses(state)

	// Get worktree statuses
	worktreeStatuses, err := w.getWorktreeStatuses(state, verbose)
	if err != nil {
		return nil, err
	}

	// Calculate summary
	summary := w.calculateSummary(repoStatuses, worktreeStatuses)

	return &WorkspaceStatus{
		WorkspaceName: w.Name,
		WorkspacePath: w.Path,
		Repos:         repoStatuses,
		Worktrees:     worktreeStatuses,
		Summary:       summary,
	}, nil
}

// getRepoStatuses checks clone status for all repos
func (w *Workspace) getRepoStatuses(state *State) []RepoStatus {
	statuses := make([]RepoStatus, 0)

	if state.Repositories == nil {
		return statuses
	}

	for repoName, repo := range state.Repositories {
		// Check if bare clone exists
		bareRepoPath := w.BareRepoPath(repoName)
		isCloned := w.isBareCloneExists(bareRepoPath)

		statuses = append(statuses, RepoStatus{
			Name:      repoName,
			URL:       repo.URL,
			IsCloned:  isCloned,
			InConfig:  true,
			ClonePath: bareRepoPath,
		})
	}

	return statuses
}

// isBareCloneExists checks if a bare repository exists
func (w *Workspace) isBareCloneExists(bareRepoPath string) bool {
	// Check for bare repo markers
	headPath := filepath.Join(bareRepoPath, "HEAD")
	configPath := filepath.Join(bareRepoPath, "config")

	_, headErr := os.Stat(headPath)
	_, configErr := os.Stat(configPath)

	return headErr == nil && configErr == nil
}

// getWorktreeStatuses collects status for all worktrees
func (w *Workspace) getWorktreeStatuses(state *State, verbose bool) ([]WorktreeStatus, error) {
	allWorktrees, err := w.GetAllWorktrees()
	if err != nil {
		return nil, err
	}

	// Get current working directory to mark current worktree
	cwd, _ := os.Getwd()

	statuses := make([]WorktreeStatus, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Collect all worktrees to process
	type worktreeInfo struct {
		repo   string
		branch string
		path   string
	}
	worktreesToProcess := make([]worktreeInfo, 0)

	for repoName, branches := range allWorktrees {
		for _, branch := range branches {
			worktreePath := w.WorktreePath(repoName, branch)
			worktreesToProcess = append(worktreesToProcess, worktreeInfo{
				repo:   repoName,
				branch: branch,
				path:   worktreePath,
			})
		}
	}

	// Process worktrees in parallel
	for _, wt := range worktreesToProcess {
		wg.Add(1)
		go func(repo, branch, path string) {
			defer wg.Done()

			status := w.detectWorktreeStatus(path, verbose)
			isCurrent := cwd != "" && strings.HasPrefix(cwd, path)

			mu.Lock()
			statuses = append(statuses, WorktreeStatus{
				Branch:         branch,
				Repo:           repo,
				Path:           path,
				Status:         status.Status,
				IsCurrent:      isCurrent,
				ModifiedFiles:  status.ModifiedFiles,
				UntrackedFiles: status.UntrackedFiles,
			})
			mu.Unlock()
		}(wt.repo, wt.branch, wt.path)
	}

	wg.Wait()

	return statuses, nil
}

// worktreeStatusDetail holds detailed status information
type worktreeStatusDetail struct {
	Status         string
	ModifiedFiles  []string
	UntrackedFiles []string
}

// detectWorktreeStatus checks git status for a single worktree
func (w *Workspace) detectWorktreeStatus(worktreePath string, verbose bool) worktreeStatusDetail {
	// Check if path exists
	if _, err := os.Stat(worktreePath); err != nil {
		return worktreeStatusDetail{Status: "missing"}
	}

	// Detect different types of changes
	hasModified, _ := git.HasUncommittedChanges(worktreePath)
	hasUntracked, _ := git.HasUntrackedFiles(worktreePath)
	hasConflicts, _ := git.HasConflicts(worktreePath)

	status := "clean"
	var modifiedFiles, untrackedFiles []string

	if hasConflicts {
		status = "conflict"
	} else if hasModified {
		status = "modified"
		if verbose {
			modifiedFiles = w.getModifiedFiles(worktreePath)
		}
	} else if hasUntracked {
		status = "untracked"
		if verbose {
			untrackedFiles = w.getUntrackedFiles(worktreePath)
		}
	}

	return worktreeStatusDetail{
		Status:         status,
		ModifiedFiles:  modifiedFiles,
		UntrackedFiles: untrackedFiles,
	}
}

// getModifiedFiles returns list of modified files (for verbose mode)
func (w *Workspace) getModifiedFiles(worktreePath string) []string {
	// Use git status --porcelain to get modified files
	files, _ := git.GetModifiedFiles(worktreePath)
	return files
}

// getUntrackedFiles returns list of untracked files (for verbose mode)
func (w *Workspace) getUntrackedFiles(worktreePath string) []string {
	// Use git status --porcelain to get untracked files
	files, _ := git.GetUntrackedFiles(worktreePath)
	return files
}

// calculateSummary generates aggregate statistics
func (w *Workspace) calculateSummary(repos []RepoStatus, worktrees []WorktreeStatus) StatusSummary {
	summary := StatusSummary{
		TotalRepos:     len(repos),
		TotalWorktrees: len(worktrees),
		ConfigInSync:   true,
	}

	// Count repos not cloned
	for _, repo := range repos {
		if !repo.IsCloned {
			summary.ReposNotCloned++
			summary.ConfigInSync = false
		}
	}

	// Count unique branches and dirty worktrees
	branchSet := make(map[string]bool)
	for _, wt := range worktrees {
		branchSet[wt.Branch] = true
		if wt.Status != "clean" {
			summary.DirtyWorktrees++
			summary.HasUncommittedChanges = true
		}
	}
	summary.TotalBranches = len(branchSet)

	return summary
}
