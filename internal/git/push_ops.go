package git

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// HasUnpushedCommits checks if a worktree has commits ahead of upstream
func HasUnpushedCommits(worktreePath string) (bool, error) {
	count, err := GetUnpushedCount(worktreePath)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUnpushedCount returns the number of commits ahead of upstream
func GetUnpushedCount(worktreePath string) (int, error) {
	// First check if there's an upstream
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "@{upstream}")
	_, err := cmd.Output()
	if err != nil {
		// No upstream configured - check if remote tracking branch exists
		branch, branchErr := GetCurrentBranch(worktreePath)
		if branchErr != nil || branch == "" {
			return 0, nil // Detached HEAD or error
		}

		// Check if origin/<branch> exists
		cmd = exec.Command("git", "-C", worktreePath, "rev-parse", "--verify", "origin/"+branch)
		if cmd.Run() != nil {
			return 0, nil // No remote tracking branch
		}

		// Count commits ahead of origin/<branch>
		cmd = exec.Command("git", "-C", worktreePath, "rev-list", "--count", "origin/"+branch+"..HEAD")
		output, err := cmd.Output()
		if err != nil {
			return 0, nil
		}

		count, _ := strconv.Atoi(strings.TrimSpace(string(output)))
		return count, nil
	}

	// Has upstream - count commits ahead
	cmd = exec.Command("git", "-C", worktreePath, "rev-list", "--count", "@{upstream}..HEAD")
	output, err := cmd.Output()
	if err != nil {
		return 0, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to count unpushed commits",
			"Verify branch has upstream tracking",
			err,
		)
	}

	count, _ := strconv.Atoi(strings.TrimSpace(string(output)))
	return count, nil
}

// GetUnpushedCommits returns the list of unpushed commits (SHA + message)
func GetUnpushedCommits(worktreePath string) ([]string, error) {
	// First check if there's an upstream
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "@{upstream}")
	_, err := cmd.Output()

	var refRange string
	if err != nil {
		// No upstream - try origin/<branch>
		branch, branchErr := GetCurrentBranch(worktreePath)
		if branchErr != nil || branch == "" {
			return []string{}, nil
		}

		// Check if origin/<branch> exists
		cmd = exec.Command("git", "-C", worktreePath, "rev-parse", "--verify", "origin/"+branch)
		if cmd.Run() != nil {
			return []string{}, nil
		}

		refRange = "origin/" + branch + "..HEAD"
	} else {
		refRange = "@{upstream}..HEAD"
	}

	cmd = exec.Command("git", "-C", worktreePath, "log", "--oneline", refRange)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get unpushed commits",
			"Verify branch has upstream tracking",
			err,
		)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	return strings.Split(outputStr, "\n"), nil
}

// PushWithOptions pushes with additional options
func PushWithOptions(worktreePath string, force bool) error {
	args := []string{"-C", worktreePath, "push"}

	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for remote has new commits scenario
		if strings.Contains(outputStr, "rejected") ||
			strings.Contains(outputStr, "non-fast-forward") {
			return errors.New(
				errors.ErrCodePushFailed,
				"Push rejected - remote has new commits",
				"Run 'fa sync --pull' first to update your branch",
			)
		}

		// Check for auth errors
		if strings.Contains(outputStr, "Authentication failed") ||
			strings.Contains(outputStr, "Permission denied") {
			return errors.New(
				errors.ErrCodeAuthenticationFailed,
				"Git authentication failed",
				"Check SSH keys or Git credentials",
			)
		}

		// Check for no upstream
		if strings.Contains(outputStr, "no upstream branch") ||
			strings.Contains(outputStr, "has no upstream") {
			return errors.New(
				errors.ErrCodeNoUpstream,
				"No upstream branch configured",
				"Set upstream with: git push -u origin <branch>",
			)
		}

		return errors.Wrap(
			errors.ErrCodePushFailed,
			"Failed to push: "+strings.TrimSpace(outputStr),
			"Check network connection and remote URL",
			err,
		)
	}

	return nil
}

// GetPushRefspec returns the refspec that would be pushed (e.g., "main -> origin/main")
func GetPushRefspec(worktreePath string) (string, error) {
	branch, err := GetCurrentBranch(worktreePath)
	if err != nil || branch == "" {
		return "", err
	}

	return branch + " -> origin/" + branch, nil
}

// HasUpstreamConfigured checks if the current branch has an upstream configured
func HasUpstreamConfigured(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "@{upstream}")
	err := cmd.Run()
	if err != nil {
		// Check if it's just "no upstream" vs actual error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			return false, nil
		}
		return false, nil // Treat errors as "no upstream"
	}
	return true, nil
}
