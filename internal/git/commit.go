package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// CommitOptions represents options for git commit
type CommitOptions struct {
	Message string
	All     bool // Stage all tracked modifications (-a)
	Amend   bool // Amend previous commit
}

// CommitStats represents statistics about a commit
type CommitStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

// HasStagedChanges checks if a worktree has staged changes
func HasStagedChanges(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff", "--cached", "--quiet")
	err := cmd.Run()
	if err != nil {
		// Exit code 1 means there are staged changes
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return true, nil
		}
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check for staged changes",
			"Verify the worktree path is valid",
			err,
		)
	}
	// Exit code 0 means no staged changes
	return false, nil
}

// Commit creates a commit in the given worktree
func Commit(worktreePath string, opts CommitOptions) (string, error) {
	args := []string{"-C", worktreePath, "commit"}

	if opts.All {
		args = append(args, "-a")
	}

	if opts.Amend {
		args = append(args, "--amend")
		if opts.Message != "" {
			args = append(args, "-m", opts.Message)
		} else {
			args = append(args, "--no-edit")
		}
	} else {
		if opts.Message == "" {
			return "", errors.New(
				errors.ErrCodeEmptyCommitMessage,
				"Commit message cannot be empty",
				"Provide a message with -m or let editor open",
			)
		}
		args = append(args, "-m", opts.Message)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for nothing to commit
		if strings.Contains(outputStr, "nothing to commit") ||
			strings.Contains(outputStr, "no changes added") {
			return "", errors.New(
				errors.ErrCodeNothingToCommit,
				"Nothing to commit",
				"Stage changes first or use -a flag",
			)
		}

		// Check for pre-commit hook failure
		if strings.Contains(outputStr, "pre-commit hook") ||
			strings.Contains(outputStr, "hook failed") {
			return "", errors.Wrap(
				errors.ErrCodeCommitFailed,
				"Pre-commit hook failed",
				"Fix the issues reported by the pre-commit hook",
				err,
			)
		}

		return "", errors.Wrap(
			errors.ErrCodeCommitFailed,
			"Commit failed: "+strings.TrimSpace(outputStr),
			"Check git status for details",
			err,
		)
	}

	// Get the commit SHA
	sha, err := GetHeadSHA(worktreePath)
	if err != nil {
		return "", err
	}

	return sha, nil
}

// GetStagedFiles returns the list of staged files in a worktree
func GetStagedFiles(worktreePath string) ([]string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff", "--cached", "--name-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get staged files",
			"Verify the worktree path is valid",
			err,
		)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	return strings.Split(outputStr, "\n"), nil
}

// GetStagedFilesWithStatus returns staged files with their status (M, A, D, etc.)
func GetStagedFilesWithStatus(worktreePath string) ([]string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff", "--cached", "--name-status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get staged files with status",
			"Verify the worktree path is valid",
			err,
		)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	return strings.Split(outputStr, "\n"), nil
}

// GetCommitStats returns statistics for a specific commit
func GetCommitStats(worktreePath, sha string) (*CommitStats, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff-tree", "--no-commit-id", "--numstat", "-r", sha)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get commit stats",
			"Verify the commit SHA is valid",
			err,
		)
	}

	stats := &CommitStats{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			stats.FilesChanged++
			// Parse insertions
			if ins, err := strconv.Atoi(parts[0]); err == nil {
				stats.Insertions += ins
			}
			// Parse deletions
			if del, err := strconv.Atoi(parts[1]); err == nil {
				stats.Deletions += del
			}
		}
	}

	return stats, nil
}

// GetHeadSHA returns the short SHA of HEAD
func GetHeadSHA(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--short", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get HEAD SHA",
			"Verify the repository has commits",
			err,
		)
	}

	return strings.TrimSpace(string(output)), nil
}

// StageAllTracked stages all tracked file modifications (git add -u)
func StageAllTracked(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "add", "-u")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to stage tracked files: "+strings.TrimSpace(string(output)),
			"Check git status for details",
			err,
		)
	}
	return nil
}

// HasTrackedChanges checks if there are modifications to tracked files
func HasTrackedChanges(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff", "--quiet")
	err := cmd.Run()
	if err != nil {
		// Exit code 1 means there are changes
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return true, nil
		}
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check for tracked changes",
			"Verify the worktree path is valid",
			err,
		)
	}
	return false, nil
}

// GetCommitMessage returns the commit message for a given SHA
func GetCommitMessage(worktreePath, sha string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "log", "-1", "--format=%s", sha)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get commit message",
			"Verify the commit SHA is valid",
			err,
		)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get current branch",
			"Verify the worktree path is valid",
			err,
		)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		// Detached HEAD state
		return "", nil
	}

	return branch, nil
}

// FormatDiffStat formats a diff stat for display
func FormatDiffStat(stats *CommitStats) string {
	if stats.FilesChanged == 0 {
		return "no changes"
	}

	fileWord := "file"
	if stats.FilesChanged != 1 {
		fileWord = "files"
	}
	parts := []string{fmt.Sprintf("%d %s", stats.FilesChanged, fileWord)}

	if stats.Insertions > 0 {
		parts = append(parts, fmt.Sprintf("+%d", stats.Insertions))
	}
	if stats.Deletions > 0 {
		parts = append(parts, fmt.Sprintf("-%d", stats.Deletions))
	}

	return strings.Join(parts, ", ")
}
