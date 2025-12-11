package git

import (
	"os/exec"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// Stash saves uncommitted changes to the stash
func Stash(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "stash", "push", "-m", "Foundagent auto-stash before sync")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to stash changes",
			"Check that the worktree has valid uncommitted changes",
			err,
		)
	}

	// Check if stash actually created something
	outputStr := string(output)
	if strings.Contains(outputStr, "No local changes to save") {
		return nil // Nothing to stash, not an error
	}

	return nil
}

// StashPop restores the most recent stashed changes
func StashPop(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "stash", "pop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for conflicts during pop
		if strings.Contains(outputStr, "CONFLICT") {
			return errors.New(
				errors.ErrCodeGitOperationFailed,
				"Stash pop created conflicts",
				"Resolve conflicts manually with 'git status' and 'git mergetool'",
			)
		}

		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to pop stash",
			"Check stash list with 'git stash list'",
			err,
		)
	}

	return nil
}

// HasStash checks if there are any stashed changes
func HasStash(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "stash", "list")
	output, err := cmd.Output()
	if err != nil {
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check stash",
			"Verify the worktree path is valid",
			err,
		)
	}

	return strings.TrimSpace(string(output)) != "", nil
}
