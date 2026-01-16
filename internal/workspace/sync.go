package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/foundagent/foundagent/internal/git"
)

// Sync status constants
const (
	SyncStatusSynced      = "synced"
	SyncStatusUpdated     = "updated"
	SyncStatusFailed      = "failed"
	SyncStatusSkipped     = "skipped"
	SyncStatusPushed      = "pushed"
	SyncStatusUpToDate    = "up-to-date"
	SyncStatusNothingPush = "nothing-to-push"
)

// Status symbol constants for output formatting
const (
	StatusSymbolSuccess = "✓"
	StatusSymbolFailed  = "✗"
	StatusSymbolSkipped = "⊘"
)

// SyncResult represents the result of syncing a single repo
type SyncResult struct {
	RepoName      string
	Status        string // "success", "failed", "skipped", "up-to-date", "updated"
	Error         error
	RefsUpdated   []string
	CommitsBehind int
	CommitsAhead  int
	Pushed        bool
}

// SyncSummary aggregates results across all repos
type SyncSummary struct {
	Total   int
	Synced  int
	Updated int
	Failed  int
	Skipped int
	Pushed  int
}

// SyncAllRepos fetches from all repos in parallel
func (w *Workspace) SyncAllRepos(verbose bool) ([]SyncResult, error) {
	// Load state to get repo list
	state, err := w.LoadState()
	if err != nil {
		return nil, err
	}

	if len(state.Repositories) == 0 {
		return []SyncResult{}, nil
	}

	// Collect repo names
	var repoNames []string
	for name := range state.Repositories {
		repoNames = append(repoNames, name)
	}

	// Execute fetch in parallel
	parallelResults := ExecuteParallel(repoNames, func(repoName string) error {
		bareRepoPath := filepath.Join(w.Path, ReposDir, repoName, BareDir)
		return git.Fetch(bareRepoPath)
	})

	// Convert to SyncResult
	results := make([]SyncResult, len(parallelResults))
	for i, pr := range parallelResults {
		if pr.Error != nil {
			results[i] = SyncResult{
				RepoName: pr.RepoName,
				Status:   SyncStatusFailed,
				Error:    pr.Error,
			}
		} else {
			// Check if there are updates available
			// TODO: Detect actual updates by comparing refs
			results[i] = SyncResult{
				RepoName: pr.RepoName,
				Status:   SyncStatusSynced,
			}
		}
	}

	return results, nil
}

// PullAllWorktrees pulls all worktrees for a branch
func (w *Workspace) PullAllWorktrees(branch string, stash bool, verbose bool) ([]SyncResult, error) {
	// First fetch
	_, err := w.SyncAllRepos(verbose)
	if err != nil {
		return nil, err
	}

	// Load state to get repo list
	state, err := w.LoadState()
	if err != nil {
		return nil, err
	}

	// Now pull each worktree for the branch
	results := make([]SyncResult, 0)

	for repoName := range state.Repositories {
		worktreePath := filepath.Join(w.Path, ReposDir, repoName, WorktreesDir, branch)

		// Check if worktree exists
		result := SyncResult{
			RepoName: repoName,
		}

		// Check for detached HEAD
		isDetached, err := git.IsDetachedHead(worktreePath)
		if err != nil {
			// Worktree might not exist
			result.Status = SyncStatusSkipped
			result.Error = fmt.Errorf("branch %s not found", branch)
			results = append(results, result)
			continue
		}

		if isDetached {
			result.Status = SyncStatusSkipped
			result.Error = fmt.Errorf("detached HEAD - cannot pull")
			results = append(results, result)
			continue
		}

		// Check if worktree has uncommitted changes
		hasChanges, err := git.HasUncommittedChanges(worktreePath)
		if err != nil {
			// Worktree might not exist
			result.Status = SyncStatusSkipped
			result.Error = fmt.Errorf("branch %s not found", branch)
			results = append(results, result)
			continue
		}

		if hasChanges {
			if stash {
				// Stash changes
				if err := git.Stash(worktreePath); err != nil {
					result.Status = SyncStatusFailed
					result.Error = fmt.Errorf("failed to stash: %w", err)
					results = append(results, result)
					continue
				}
			} else {
				// Skip dirty worktree
				result.Status = SyncStatusSkipped
				result.Error = fmt.Errorf("uncommitted changes")
				results = append(results, result)
				continue
			}
		}

		// Pull
		pullErr := git.Pull(worktreePath)

		// Pop stash if we stashed
		if stash && hasChanges {
			if popErr := git.StashPop(worktreePath); popErr != nil {
				result.Status = SyncStatusFailed
				result.Error = fmt.Errorf("pull succeeded but failed to pop stash: %w", popErr)
				results = append(results, result)
				continue
			}
		}

		if pullErr != nil {
			result.Status = SyncStatusFailed
			result.Error = pullErr
		} else {
			result.Status = SyncStatusUpdated
		}

		results = append(results, result)
	}

	// Merge fetch and pull results
	return results, nil
}

// PushAllRepos pushes all repos with unpushed commits
func (w *Workspace) PushAllRepos(verbose bool) ([]SyncResult, error) {
	// Load state
	state, err := w.LoadState()
	if err != nil {
		return nil, err
	}

	results := make([]SyncResult, 0)

	// For each repo, check all worktrees for unpushed commits
	for repoName := range state.Repositories {
		worktreesDir := filepath.Join(w.Path, ReposDir, WorktreesDir, repoName)

		// List all branches (subdirectories)
		entries, err := filepath.Glob(filepath.Join(worktreesDir, "*"))
		if err != nil {
			continue
		}

		pushed := false
		var pushErr error

		for _, entry := range entries {
			branch := filepath.Base(entry)
			worktreePath := entry

			// Check ahead/behind
			ahead, _, err := git.GetAheadBehindCount(worktreePath, branch)
			if err != nil || ahead == 0 {
				continue // No unpushed commits
			}

			// Push this worktree
			if err := git.Push(worktreePath); err != nil {
				pushErr = err
				break
			}
			pushed = true
		}

		result := SyncResult{
			RepoName: repoName,
			Pushed:   pushed,
		}

		if pushErr != nil {
			result.Status = SyncStatusFailed
			result.Error = pushErr
		} else if pushed {
			result.Status = SyncStatusPushed
		} else {
			result.Status = SyncStatusNothingPush
		}

		results = append(results, result)
	}

	return results, nil
}

// CalculateSummary aggregates sync results into a summary
func CalculateSummary(results []SyncResult) SyncSummary {
	summary := SyncSummary{
		Total: len(results),
	}

	for _, r := range results {
		switch r.Status {
		case SyncStatusSynced, SyncStatusUpToDate:
			summary.Synced++
		case SyncStatusUpdated:
			summary.Updated++
		case SyncStatusFailed:
			summary.Failed++
		case SyncStatusSkipped:
			summary.Skipped++
		case SyncStatusPushed:
			summary.Pushed++
		}
	}

	return summary
}

// FormatSyncResults formats sync results for human output
func FormatSyncResults(results []SyncResult, operation string) string {
	var output strings.Builder

	for _, r := range results {
		status := StatusSymbolSuccess
		if r.Status == SyncStatusFailed {
			status = StatusSymbolFailed
		} else if r.Status == SyncStatusSkipped {
			status = StatusSymbolSkipped
		}

		output.WriteString(fmt.Sprintf("%s %s: %s", status, r.RepoName, r.Status))

		if r.Error != nil {
			output.WriteString(fmt.Sprintf(" (%s)", r.Error.Error()))
		}

		output.WriteString("\n")
	}

	return output.String()
}
