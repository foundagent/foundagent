package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// GetDefaultBranch retrieves the default branch from a bare repository
func GetDefaultBranch(bareRepoPath string) (string, error) {
	// Try to get the symbolic ref for HEAD
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = bareRepoPath

	output, err := cmd.Output()
	if err != nil {
		// If symbolic-ref fails, try to get the default branch from remote
		cmd = exec.Command("git", "remote", "show", "origin")
		cmd.Dir = bareRepoPath

		output, err = cmd.Output()
		if err != nil {
			// Fallback to common default branches
			return "main", nil
		}

		// Parse "HEAD branch: main" from output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "HEAD branch:") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					branch := strings.TrimSpace(parts[1])
					if branch != "" {
						return branch, nil
					}
				}
			}
		}

		return "main", nil
	}

	// Parse refs/remotes/origin/HEAD -> refs/remotes/origin/main
	ref := strings.TrimSpace(string(output))
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		branch := parts[len(parts)-1]
		if branch != "" {
			return branch, nil
		}
	}

	return "main", nil
}

// ListRemoteBranches lists all remote branches
func ListRemoteBranches(bareRepoPath string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	cmd.Dir = bareRepoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to list remote branches",
			"Ensure the repository is valid",
			err,
		)
	}

	var branches []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "->") {
			continue
		}
		// Remove "origin/" prefix
		if strings.HasPrefix(line, "origin/") {
			branches = append(branches, line[7:])
		}
	}

	return branches, nil
}

// Fetch fetches from origin remote
func Fetch(repoPath string) error {
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeNetworkError,
			"Failed to fetch from remote",
			"Check network connection and remote URL",
			err,
		)
	}

	// Check for auth errors in output
	outputStr := string(output)
	if strings.Contains(outputStr, "Authentication failed") ||
		strings.Contains(outputStr, "Permission denied") {
		return errors.New(
			errors.ErrCodeAuthenticationFailed,
			"Git authentication failed",
			"Check SSH keys or Git credentials",
		)
	}

	return nil
}

// Pull performs a fast-forward pull on a worktree
func Pull(worktreePath string) error {
	// Use --ff-only to ensure fast-forward only
	cmd := exec.Command("git", "-C", worktreePath, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for non-fast-forward scenario
		if strings.Contains(outputStr, "Not possible to fast-forward") ||
			strings.Contains(outputStr, "divergent branches") {
			return errors.New(
				errors.ErrCodeGitOperationFailed,
				"Cannot fast-forward - branches have diverged",
				"Run 'git merge' or 'git rebase' manually to resolve",
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

		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to pull",
			"Ensure worktree is clean and branch tracking is set up",
			err,
		)
	}

	return nil
}

// Push pushes local commits to remote
func Push(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "push")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for remote has new commits scenario
		if strings.Contains(outputStr, "rejected") ||
			strings.Contains(outputStr, "non-fast-forward") {
			return errors.New(
				errors.ErrCodeGitOperationFailed,
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

		return errors.Wrap(
			errors.ErrCodeNetworkError,
			"Failed to push to remote",
			"Check network connection and remote URL",
			err,
		)
	}

	return nil
}

// GetAheadBehindCount returns the number of commits ahead and behind remote
func GetAheadBehindCount(worktreePath, branch string) (ahead int, behind int, err error) {
	// Get tracking branch
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		// No tracking branch set up
		return 0, 0, nil
	}

	upstream := strings.TrimSpace(string(output))

	// Get ahead/behind counts
	cmd = exec.Command("git", "-C", worktreePath, "rev-list", "--left-right", "--count", branch+"..."+upstream)
	output, cmdErr = cmd.Output()
	if cmdErr != nil {
		return 0, 0, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to get ahead/behind count",
			"Ensure branch has upstream tracking",
			cmdErr,
		)
	}

	// Parse output: "ahead\tbehind"
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) == 2 {
		_, scanErr := fmt.Sscanf(parts[0], "%d", &ahead)
		if scanErr != nil {
			return 0, 0, scanErr
		}
		_, scanErr = fmt.Sscanf(parts[1], "%d", &behind)
		if scanErr != nil {
			return 0, 0, scanErr
		}
	}

	return ahead, behind, nil
}
