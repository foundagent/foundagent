package workspace

import (
	"fmt"
	"strings"

	"github.com/foundagent/foundagent/internal/git"
)

// Commit status constants
const (
	CommitStatusCommitted   = "committed"
	CommitStatusWouldCommit = "would-commit"
	CommitStatusSkipped     = "skipped"
	CommitStatusFailed      = "failed"
)

// CommitResult represents the result of committing a single repo
type CommitResult struct {
	RepoName     string `json:"name"`
	Status       string `json:"status"` // "committed", "skipped", "failed"
	CommitSHA    string `json:"commit_sha,omitempty"`
	FilesChanged int    `json:"files_changed"`
	Insertions   int    `json:"insertions"`
	Deletions    int    `json:"deletions"`
	Error        error  `json:"-"`
	ErrorMessage string `json:"error,omitempty"`
}

// CommitSummary aggregates commit results across all repos
type CommitSummary struct {
	Total     int `json:"total"`
	Committed int `json:"committed"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// CommitOptions represents options for cross-repo commit
type CommitOptions struct {
	Message       string
	All           bool     // Stage all tracked modifications (-a)
	Amend         bool     // Amend previous commit
	DryRun        bool     // Preview without executing
	Repos         []string // Limit to specific repos (nil = all)
	Verbose       bool     // Show detailed output
	AllowDetached bool     // Allow commits in detached HEAD
}

// CommitAllRepos commits across all repos with staged changes
func (w *Workspace) CommitAllRepos(opts CommitOptions) ([]CommitResult, error) {
	state, err := w.LoadState()
	if err != nil {
		return nil, err
	}

	if len(state.Repositories) == 0 {
		return []CommitResult{}, nil
	}

	currentBranch := getCurrentBranch(state)
	repoNames, err := w.filterRepoNames(state, opts.Repos)
	if err != nil {
		return nil, err
	}

	// First pass: check what will be committed and get pre-commit state
	repoStates := w.prepareCommitStates(repoNames, currentBranch, opts)

	// Execute commits in parallel
	parallelResults := ExecuteParallel(repoNames, func(repoName string) error {
		return w.executeCommit(repoStates[repoName], opts)
	})

	// Convert parallel results to commit results
	return w.buildCommitResults(parallelResults, repoStates, opts), nil
}

func getCurrentBranch(state *State) string {
	if state.CurrentBranch != "" {
		return state.CurrentBranch
	}
	return "main"
}

func (w *Workspace) filterRepoNames(state *State, requestedRepos []string) ([]string, error) {
	var repoNames []string
	for name := range state.Repositories {
		if len(requestedRepos) > 0 && !containsString(requestedRepos, name) {
			continue
		}
		repoNames = append(repoNames, name)
	}

	// Validate requested repos exist
	for _, r := range requestedRepos {
		if _, exists := state.Repositories[r]; !exists {
			return nil, fmt.Errorf("repository '%s' not found in workspace", r)
		}
	}
	return repoNames, nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

type commitRepoState struct {
	worktreePath string
	hasStaged    bool
	isDetached   bool
	preCommitSHA string
}

func (w *Workspace) prepareCommitStates(repoNames []string, currentBranch string, opts CommitOptions) map[string]*commitRepoState {
	repoStates := make(map[string]*commitRepoState)

	for _, repoName := range repoNames {
		worktreePath := w.WorktreePath(repoName, currentBranch)
		rs := &commitRepoState{worktreePath: worktreePath}

		rs.isDetached, _ = git.IsDetachedHead(worktreePath)
		rs.preCommitSHA, _ = git.GetHeadSHA(worktreePath)

		// If -a flag, stage all tracked changes first
		if opts.All {
			if hasTracked, err := git.HasTrackedChanges(worktreePath); err == nil && hasTracked {
				_ = git.StageAllTracked(worktreePath)
			}
		}

		rs.hasStaged, _ = git.HasStagedChanges(worktreePath)
		repoStates[repoName] = rs
	}
	return repoStates
}

func (w *Workspace) executeCommit(rs *commitRepoState, opts CommitOptions) error {
	if rs.isDetached && !opts.AllowDetached {
		return fmt.Errorf("detached HEAD - use --allow-detached to commit anyway")
	}
	if !rs.hasStaged && !opts.Amend {
		return nil // Will be marked as skipped
	}
	if opts.DryRun {
		return nil
	}
	_, err := git.Commit(rs.worktreePath, git.CommitOptions{
		Message: opts.Message,
		All:     false,
		Amend:   opts.Amend,
	})
	return err
}

func (w *Workspace) buildCommitResults(parallelResults []ParallelResult, repoStates map[string]*commitRepoState, opts CommitOptions) []CommitResult {
	results := make([]CommitResult, len(parallelResults))
	for i, pr := range parallelResults {
		rs := repoStates[pr.RepoName]
		results[i] = w.buildSingleCommitResult(pr, rs, opts)
	}
	return results
}

func (w *Workspace) buildSingleCommitResult(pr ParallelResult, rs *commitRepoState, opts CommitOptions) CommitResult {
	result := CommitResult{RepoName: pr.RepoName}

	switch {
	case pr.Error != nil:
		result.Status = CommitStatusFailed
		result.Error = pr.Error
		result.ErrorMessage = pr.Error.Error()
	case !rs.hasStaged && !opts.Amend:
		result.Status = CommitStatusSkipped
		result.ErrorMessage = "nothing to commit"
	case opts.DryRun:
		result.Status = CommitStatusWouldCommit
		files, _ := git.GetStagedFiles(rs.worktreePath)
		result.FilesChanged = len(files)
	default:
		w.populateCommitSuccess(&result, rs)
	}
	return result
}

func (w *Workspace) populateCommitSuccess(result *CommitResult, rs *commitRepoState) {
	newSHA, _ := git.GetHeadSHA(rs.worktreePath)
	if newSHA != rs.preCommitSHA {
		result.Status = CommitStatusCommitted
		result.CommitSHA = newSHA
		if stats, _ := git.GetCommitStats(rs.worktreePath, newSHA); stats != nil {
			result.FilesChanged = stats.FilesChanged
			result.Insertions = stats.Insertions
			result.Deletions = stats.Deletions
		}
	} else {
		result.Status = CommitStatusSkipped
		result.ErrorMessage = "nothing to commit"
	}
}

// CalculateCommitSummary aggregates commit results into a summary
func CalculateCommitSummary(results []CommitResult) CommitSummary {
	summary := CommitSummary{
		Total: len(results),
	}

	for _, r := range results {
		switch r.Status {
		case CommitStatusCommitted, CommitStatusWouldCommit:
			summary.Committed++
		case CommitStatusSkipped:
			summary.Skipped++
		case CommitStatusFailed:
			summary.Failed++
		}
	}

	return summary
}

// FormatCommitResults formats commit results for human output
func FormatCommitResults(results []CommitResult) string {
	var output strings.Builder

	for _, r := range results {
		var status string
		switch r.Status {
		case CommitStatusCommitted:
			status = StatusSymbolSuccess
		case CommitStatusFailed:
			status = StatusSymbolFailed
		case CommitStatusSkipped, CommitStatusWouldCommit:
			status = StatusSymbolSkipped
		default:
			status = "?"
		}

		output.WriteString(fmt.Sprintf("%s %s: %s", status, r.RepoName, r.Status))

		if r.Status == CommitStatusCommitted && r.CommitSHA != "" {
			output.WriteString(fmt.Sprintf(" %s", r.CommitSHA))
			if r.FilesChanged > 0 {
				output.WriteString(fmt.Sprintf(" (%d file", r.FilesChanged))
				if r.FilesChanged != 1 {
					output.WriteString("s")
				}
				if r.Insertions > 0 || r.Deletions > 0 {
					output.WriteString(fmt.Sprintf(", +%d -%d", r.Insertions, r.Deletions))
				}
				output.WriteString(")")
			}
		}

		if r.ErrorMessage != "" && r.Status != CommitStatusCommitted {
			output.WriteString(fmt.Sprintf(" (%s)", r.ErrorMessage))
		}

		output.WriteString("\n")
	}

	return output.String()
}

// GetDryRunPreview returns preview info for dry-run mode
func (w *Workspace) GetDryRunPreview(opts CommitOptions) ([]CommitResult, error) {
	opts.DryRun = true
	return w.CommitAllRepos(opts)
}
