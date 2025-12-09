package git

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/foundagent/foundagent/internal/errors"
)

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	URL        string
	TargetPath string
	Bare       bool
	Progress   bool
}

// Clone performs a git clone operation
func Clone(opts CloneOptions) error {
	args := []string{"clone"}

	if opts.Bare {
		args = append(args, "--bare")
	}

	if !opts.Progress {
		args = append(args, "--quiet")
	}

	args = append(args, opts.URL, opts.TargetPath)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Check if it's an auth error
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if exitCode == 128 {
				return errors.Wrap(
					errors.ErrCodeGitOperationFailed,
					fmt.Sprintf("Failed to clone repository: %s", opts.URL),
					"Check that the repository exists and you have access. For private repos, ensure your SSH key is configured or use HTTPS with credentials",
					err,
				)
			}
		}

		return errors.Wrap(
			errors.ErrCodeGitOperationFailed,
			fmt.Sprintf("Git clone failed: %s", opts.URL),
			"Verify the repository URL and your network connection",
			err,
		)
	}

	return nil
}

// CloneBare clones a repository as a bare clone
func CloneBare(url, targetPath string, progress bool) error {
	return Clone(CloneOptions{
		URL:        url,
		TargetPath: targetPath,
		Bare:       true,
		Progress:   progress,
	})
}
