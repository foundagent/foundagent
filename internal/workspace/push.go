package workspace

import (
	"fmt"
	"strings"

	"github.com/foundagent/foundagent/internal/git"
)

// Push status constants
const (
	PushStatusPushed    = "pushed"
	PushStatusWouldPush = "would-push"
	PushStatusSkipped   = "skipped"
	PushStatusFailed    = "failed"
)

// PushResult represents the result of pushing a single repo
type PushResult struct {
	RepoName      string   `json:"name"`
	Status        string   `json:"status"` // "pushed", "skipped", "failed"
	RefsPushed    []string `json:"refs_pushed,omitempty"`
	CommitsPushed int      `json:"commits_pushed"`
	Error         error    `json:"-"`
	ErrorMessage  string   `json:"error,omitempty"`
}

// PushSummary aggregates push results across all repos
type PushSummary struct {
	Total   int `json:"total"`
	Pushed  int `json:"pushed"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

// PushOptions represents options for cross-repo push
type PushOptions struct {
	DryRun  bool     // Preview without executing
	Repos   []string // Limit to specific repos (nil = all)
	Verbose bool     // Show detailed output
	Force   bool     // Force push (dangerous)
}

// PushAllReposNew pushes all repos with unpushed commits
func (w *Workspace) PushAllReposNew(opts PushOptions) ([]PushResult, error) {
	state, err := w.LoadState()
	if err != nil {
		return nil, err
	}

	if len(state.Repositories) == 0 {
		return []PushResult{}, nil
	}

	currentBranch := getPushCurrentBranch(state)
	repoNames, err := w.filterPushRepoNames(state, opts.Repos)
	if err != nil {
		return nil, err
	}

	// First pass: check what will be pushed
	repoStates := w.preparePushStates(repoNames, currentBranch)

	// Execute pushes in parallel
	parallelResults := ExecuteParallel(repoNames, func(repoName string) error {
		return w.executePush(repoStates[repoName], opts)
	})

	// Convert parallel results to push results
	return w.buildPushResults(parallelResults, repoStates, opts), nil
}

func getPushCurrentBranch(state *State) string {
	if state.CurrentBranch != "" {
		return state.CurrentBranch
	}
	return "main"
}

func (w *Workspace) filterPushRepoNames(state *State, requestedRepos []string) ([]string, error) {
	var repoNames []string
	for name := range state.Repositories {
		if len(requestedRepos) > 0 && !containsPushString(requestedRepos, name) {
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

func containsPushString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

type pushRepoState struct {
	worktreePath  string
	hasUnpushed   bool
	unpushedCount int
	refspec       string
}

func (w *Workspace) preparePushStates(repoNames []string, currentBranch string) map[string]*pushRepoState {
	repoStates := make(map[string]*pushRepoState)

	for _, repoName := range repoNames {
		worktreePath := w.WorktreePath(repoName, currentBranch)
		rs := &pushRepoState{worktreePath: worktreePath}

		hasUnpushed, _ := git.HasUnpushedCommits(worktreePath)
		rs.hasUnpushed = hasUnpushed

		if hasUnpushed {
			rs.unpushedCount, _ = git.GetUnpushedCount(worktreePath)
			rs.refspec, _ = git.GetPushRefspec(worktreePath)
		}

		repoStates[repoName] = rs
	}
	return repoStates
}

func (w *Workspace) executePush(rs *pushRepoState, opts PushOptions) error {
	if !rs.hasUnpushed {
		return nil // Will be marked as skipped
	}
	if opts.DryRun {
		return nil
	}
	return git.PushWithOptions(rs.worktreePath, opts.Force)
}

func (w *Workspace) buildPushResults(parallelResults []ParallelResult, repoStates map[string]*pushRepoState, opts PushOptions) []PushResult {
	results := make([]PushResult, len(parallelResults))
	for i, pr := range parallelResults {
		rs := repoStates[pr.RepoName]
		results[i] = buildSinglePushResult(pr, rs, opts)
	}
	return results
}

func buildSinglePushResult(pr ParallelResult, rs *pushRepoState, opts PushOptions) PushResult {
	result := PushResult{RepoName: pr.RepoName}

	switch {
	case pr.Error != nil:
		result.Status = PushStatusFailed
		result.Error = pr.Error
		result.ErrorMessage = pr.Error.Error()
	case !rs.hasUnpushed:
		result.Status = PushStatusSkipped
		result.ErrorMessage = "nothing to push"
	case opts.DryRun:
		result.Status = PushStatusWouldPush
		result.CommitsPushed = rs.unpushedCount
		if rs.refspec != "" {
			result.RefsPushed = []string{rs.refspec}
		}
	default:
		result.Status = PushStatusPushed
		result.CommitsPushed = rs.unpushedCount
		if rs.refspec != "" {
			result.RefsPushed = []string{rs.refspec}
		}
	}
	return result
}

// CalculatePushSummary aggregates push results into a summary
func CalculatePushSummary(results []PushResult) PushSummary {
	summary := PushSummary{
		Total: len(results),
	}

	for _, r := range results {
		switch r.Status {
		case PushStatusPushed, PushStatusWouldPush:
			summary.Pushed++
		case PushStatusSkipped:
			summary.Skipped++
		case PushStatusFailed:
			summary.Failed++
		}
	}

	return summary
}

// FormatPushResults formats push results for human output
func FormatPushResults(results []PushResult) string {
	var output strings.Builder

	for _, r := range results {
		var status string
		switch r.Status {
		case PushStatusPushed:
			status = StatusSymbolSuccess
		case PushStatusFailed:
			status = StatusSymbolFailed
		case PushStatusSkipped, PushStatusWouldPush:
			status = StatusSymbolSkipped
		default:
			status = "?"
		}

		output.WriteString(fmt.Sprintf("%s %s: %s", status, r.RepoName, r.Status))

		if (r.Status == PushStatusPushed || r.Status == PushStatusWouldPush) && r.CommitsPushed > 0 {
			output.WriteString(fmt.Sprintf(" %d commit", r.CommitsPushed))
			if r.CommitsPushed != 1 {
				output.WriteString("s")
			}
			if len(r.RefsPushed) > 0 {
				output.WriteString(fmt.Sprintf(" (%s)", r.RefsPushed[0]))
			}
		}

		if r.ErrorMessage != "" && r.Status != PushStatusPushed {
			output.WriteString(fmt.Sprintf(" (%s)", r.ErrorMessage))
		}

		output.WriteString("\n")
	}

	return output.String()
}

// GetPushDryRunPreview returns preview info for dry-run mode
func (w *Workspace) GetPushDryRunPreview(opts PushOptions) ([]PushResult, error) {
	opts.DryRun = true
	return w.PushAllReposNew(opts)
}
