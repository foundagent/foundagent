package git

import (
	"os/exec"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// HasUncommittedChanges checks if a worktree has uncommitted changes
func HasUncommittedChanges(worktreePath string) (bool, error) {
	// Check for staged and unstaged changes
	cmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check git status",
			"Verify the worktree path is valid",
			err,
		)
	}

	// If output is non-empty, there are changes
	return strings.TrimSpace(string(output)) != "", nil
}

// IsClean checks if a worktree is clean (no uncommitted changes)
func IsClean(worktreePath string) (bool, error) {
	hasChanges, err := HasUncommittedChanges(worktreePath)
	if err != nil {
		return false, err
	}
	return !hasChanges, nil
}

// HasUntrackedFiles checks if a worktree has untracked files
func HasUntrackedFiles(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "ls-files", "--others", "--exclude-standard")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check for untracked files",
			"Verify the worktree path is valid",
			err,
		)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// HasConflicts checks if a worktree has merge conflicts
func HasConflicts(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check for merge conflicts",
			"Verify the worktree path is valid",
			err,
		)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// GetModifiedFiles returns a list of modified files
func GetModifiedFiles(worktreePath string) ([]string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff", "--name-only", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get modified files",
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

// GetUntrackedFiles returns a list of untracked files
func GetUntrackedFiles(worktreePath string) ([]string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "ls-files", "--others", "--exclude-standard")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get untracked files",
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
