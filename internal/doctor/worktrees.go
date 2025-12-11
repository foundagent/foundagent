package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/workspace"
)

// WorktreesCheck checks worktree integrity
type WorktreesCheck struct {
	Workspace *workspace.Workspace
}

func (c WorktreesCheck) Name() string {
	return "Worktree integrity"
}

func (c WorktreesCheck) Run() CheckResult {
	state, err := c.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Could not load state file",
			Remediation: "Run 'fa doctor --fix' to regenerate state file",
			Fixable:     true,
		}
	}

	issues := make([]string, 0)

	// Check each worktree
	for _, repo := range state.Repositories {
		for _, wt := range repo.Worktrees {
			wtPath := filepath.Join(c.Workspace.Path, workspace.ReposDir, workspace.WorktreesDir, repo.Name, wt)

			// Check if directory exists
			if _, err := os.Stat(wtPath); os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("Missing worktree: %s/%s", repo.Name, wt))
				continue
			}

			// Check if it's a valid git worktree (has .git file or directory)
			gitPath := filepath.Join(wtPath, ".git")
			if _, err := os.Stat(gitPath); os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("Invalid git worktree: %s/%s", repo.Name, wt))
			}
		}
	}

	if len(issues) > 0 {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     fmt.Sprintf("Found %d worktree issue(s)", len(issues)),
			Remediation: "Run 'fa remove' to clean up or 'fa worktree create' to recreate",
			Fixable:     false,
		}
	}

	// Count total worktrees
	totalWorktrees := 0
	for _, repo := range state.Repositories {
		totalWorktrees += len(repo.Worktrees)
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("All %d worktrees valid", totalWorktrees),
		Fixable: false,
	}
}

// OrphanedWorktreesCheck checks for orphaned worktree directories
type OrphanedWorktreesCheck struct {
	Workspace *workspace.Workspace
}

func (c OrphanedWorktreesCheck) Name() string {
	return "Orphaned worktrees"
}

func (c OrphanedWorktreesCheck) Run() CheckResult {
	state, err := c.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Could not load state file",
			Remediation: "Run 'fa doctor --fix' to regenerate state file",
			Fixable:     true,
		}
	}

	worktreesDir := filepath.Join(c.Workspace.Path, workspace.ReposDir, workspace.WorktreesDir)

	// Get all directories in worktrees/
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		// Directory doesn't exist or can't be read
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "No worktrees found",
			Fixable: false,
		}
	}

	// Build map of known worktrees by repo name
	knownRepoWorktrees := make(map[string]map[string]bool)
	for _, repo := range state.Repositories {
		if _, exists := knownRepoWorktrees[repo.Name]; !exists {
			knownRepoWorktrees[repo.Name] = make(map[string]bool)
		}
		for _, wt := range repo.Worktrees {
			knownRepoWorktrees[repo.Name][wt] = true
		}
	}

	// Check for orphaned directories (repos or worktrees)
	orphaned := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoName := entry.Name()
		
		// Check if this repo exists in state
		worktreesForRepo, repoExists := knownRepoWorktrees[repoName]
		if !repoExists {
			// Entire repo directory is orphaned
			orphaned = append(orphaned, repoName)
			continue
		}
		
		// Check worktrees within this repo
		repoPath := filepath.Join(worktreesDir, repoName)
		wtEntries, err := os.ReadDir(repoPath)
		if err != nil {
			continue
		}
		
		for _, wtEntry := range wtEntries {
			if !wtEntry.IsDir() {
				continue
			}
			
			wtName := wtEntry.Name()
			if !worktreesForRepo[wtName] {
				orphaned = append(orphaned, repoName+"/"+wtName)
			}
		}
	}

	if len(orphaned) > 0 {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusWarn,
			Message:     fmt.Sprintf("Found %d orphaned worktree directories", len(orphaned)),
			Remediation: "Run 'fa doctor --fix' to remove orphaned entries",
			Fixable:     true,
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "No orphaned worktrees",
		Fixable: false,
	}
}
