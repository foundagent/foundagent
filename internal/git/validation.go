package git

import (
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// ValidateBranchName validates that a branch name is valid for git
func ValidateBranchName(name string) error {
	if name == "" {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name cannot be empty",
			"Provide a valid branch name",
		)
	}

	// Check for disallowed characters
	if strings.ContainsAny(name, " ~^:?*[\\") {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name contains invalid characters",
			"Branch names cannot contain: space ~ ^ : ? * [ \\",
		)
	}

	// Cannot start with -
	if strings.HasPrefix(name, "-") {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name cannot start with '-'",
			"Choose a different branch name",
		)
	}

	// Cannot end with /
	if strings.HasSuffix(name, "/") {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name cannot end with '/'",
			"Remove trailing slash from branch name",
		)
	}

	// Cannot start with /
	if strings.HasPrefix(name, "/") {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name cannot start with '/'",
			"Remove leading slash from branch name",
		)
	}

	// Cannot end with .lock
	if strings.HasSuffix(name, ".lock") {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name cannot end with '.lock'",
			"Choose a different branch name",
		)
	}

	// Cannot have consecutive slashes
	if strings.Contains(name, "//") {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name cannot contain consecutive slashes",
			"Remove extra slashes from branch name",
		)
	}

	// Cannot be just "." or ".."
	if name == "." || name == ".." {
		return errors.New(
			errors.ErrCodeInvalidInput,
			"Branch name cannot be '.' or '..'",
			"Choose a different branch name",
		)
	}

	return nil
}
