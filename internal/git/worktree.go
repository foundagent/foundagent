package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/errors"
)

// WorktreeAddOptions represents options for adding a worktree
type WorktreeAddOptions struct {
	BareRepoPath string
	WorktreePath string
	Branch       string
	Track        bool
}

// WorktreeAdd creates a new worktree from a bare repository
func WorktreeAdd(opts WorktreeAddOptions) error {
	args := []string{"--git-dir=" + opts.BareRepoPath, "worktree", "add"}

	args = append(args, opts.WorktreePath, opts.Branch)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			fmt.Sprintf("Failed to create worktree for branch %s", opts.Branch),
			"Ensure the branch exists in the remote repository",
			err,
		)
	}

	return nil
}

// WorktreeAddNew creates a new worktree with a new branch from a source branch
func WorktreeAddNew(bareRepoPath, worktreePath, newBranch, sourceBranch string) error {
	// Create worktree with new branch checked out from source branch
	args := []string{"--git-dir=" + bareRepoPath, "worktree", "add", "-b", newBranch, worktreePath, sourceBranch}

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			fmt.Sprintf("Failed to create worktree with new branch %s from %s", newBranch, sourceBranch),
			"Ensure the source branch exists",
			err,
		)
	}

	return nil
}

// WorktreeRemove removes a worktree
func WorktreeRemove(bareRepoPath, worktreePath string, force bool) error {
	args := []string{"--git-dir=" + bareRepoPath, "worktree", "remove"}

	if force {
		args = append(args, "--force")
	}

	args = append(args, worktreePath)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			fmt.Sprintf("Failed to remove worktree at %s", worktreePath),
			"Check that worktree exists and has no uncommitted changes (use --force to override)",
			err,
		)
	}

	return nil
}

// WorktreeList lists all worktrees for a repository
func WorktreeList(bareRepoPath string) ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = bareRepoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			"Failed to list worktrees",
			"Ensure the repository is valid",
			err,
		)
	}

	// Parse worktree list output
	// Format: "worktree /path/to/worktree"
	var worktrees []string
	lines := string(output)
	for _, line := range filepath.SplitList(lines) {
		if len(line) > 9 && line[:9] == "worktree " {
			worktrees = append(worktrees, line[9:])
		}
	}

	return worktrees, nil
}
