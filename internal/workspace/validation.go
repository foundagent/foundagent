package workspace

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// Platform-specific path limits
const (
	maxPathLengthUnix    = 4096
	maxPathLengthWindows = 260
)

// ValidateName validates a workspace name for filesystem compatibility
func ValidateName(name string) error {
	// Trim spaces
	name = strings.TrimSpace(name)

	// Check if empty
	if name == "" {
		return errors.New(
			errors.ErrCodeInvalidName,
			"Workspace name cannot be empty",
			"Provide a valid workspace name, e.g., 'fa init my-workspace'",
		)
	}

	// Check for . and ..
	if name == "." || name == ".." {
		return errors.New(
			errors.ErrCodeInvalidName,
			"Workspace name cannot be '.' or '..'",
			"Choose a descriptive name for your workspace",
		)
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return errors.New(
				errors.ErrCodeInvalidName,
				fmt.Sprintf("Workspace name contains invalid character: %s", char),
				fmt.Sprintf("Invalid characters: %s", strings.Join(invalidChars, " ")),
			)
		}
	}

	// Check for leading/trailing spaces (after trim check)
	originalName := name
	trimmedName := strings.TrimSpace(name)
	if originalName != trimmedName {
		return errors.New(
			errors.ErrCodeInvalidName,
			"Workspace name cannot have leading or trailing spaces",
			"Remove spaces from the beginning and end of the name",
		)
	}

	return nil
}

// ValidatePathLength checks if the resulting path would be too long
func ValidatePathLength(path string) error {
	maxLength := maxPathLengthUnix
	if runtime.GOOS == "windows" {
		maxLength = maxPathLengthWindows
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to resolve absolute path",
			"Ensure the path is valid",
			err,
		)
	}

	if len(absPath) > maxLength {
		return errors.New(
			errors.ErrCodePathTooLong,
			fmt.Sprintf("Path exceeds maximum length of %d characters", maxLength),
			fmt.Sprintf("Current path length: %d. Use a shorter workspace name or create it in a directory with a shorter path", len(absPath)),
		)
	}

	return nil
}
