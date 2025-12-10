package git

import (
	"os/exec"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// BranchExists checks if a branch exists in a repository
func BranchExists(bareRepoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "rev-parse", "--verify", "refs/heads/"+branchName)
	err := cmd.Run()
	if err != nil {
		// Exit code 128 means branch doesn't exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			return false, nil
		}
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check if branch exists",
			"Check repository state with 'git status'",
			err,
		)
	}
	return true, nil
}

// CreateBranch creates a new branch from a source branch in a bare repository
func CreateBranch(bareRepoPath, newBranch, sourceBranch string) error {
	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "branch", newBranch, sourceBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to create branch: "+strings.TrimSpace(string(output)),
			"Check that source branch exists",
			err,
		)
	}
	return nil
}

// DeleteBranch deletes a branch from a bare repository
func DeleteBranch(bareRepoPath, branchName string, force bool) error {
	deleteFlag := "-d"
	if force {
		deleteFlag = "-D"
	}
	
	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "branch", deleteFlag, branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to delete branch: "+strings.TrimSpace(string(output)),
			"Check that branch exists and is fully merged (or use --force)",
			err,
		)
	}
	return nil
}

// IsBranchMerged checks if a branch is fully merged into another branch
func IsBranchMerged(bareRepoPath, branch, baseBranch string) (bool, error) {
	// Use git branch --merged to check if branch is in the merged list
	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "branch", "--merged", baseBranch, "--format=%(refname:short)")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to check if branch is merged",
			"Check repository state",
			err,
		)
	}

	mergedBranches := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, b := range mergedBranches {
		if strings.TrimSpace(b) == branch {
			return true, nil
		}
	}
	return false, nil
}

// GetBranches lists all branches in a repository
func GetBranches(bareRepoPath string) ([]string, error) {
	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "branch", "--list", "--format=%(refname:short)")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to list branches",
			"Check repository state",
			err,
		)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			branches = append(branches, strings.TrimSpace(line))
		}
	}
	return branches, nil
}
