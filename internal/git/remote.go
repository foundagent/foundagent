package git

import (
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
