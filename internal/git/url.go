package git

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
)

// URL patterns
var (
	// SSH: git@github.com:owner/repo.git
	sshPattern = regexp.MustCompile(`^git@[^:]+:(.+)$`)
	// HTTPS: https://github.com/owner/repo.git
	httpsPattern = regexp.MustCompile(`^https?://[^/]+/(.+)$`)
	// File: file:///path/to/repo (for local repositories)
	filePattern = regexp.MustCompile(`^file://(.+)$`)
)

// ParseURL parses a Git URL and extracts the repository path
func ParseURL(url string) (string, error) {
	url = strings.TrimSpace(url)

	if url == "" {
		return "", errors.New(
			errors.ErrCodeInvalidRepository,
			"Repository URL cannot be empty",
			"Provide a valid Git repository URL",
		)
	}

	// Try SSH pattern
	if matches := sshPattern.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1], nil
	}

	// Try HTTPS pattern
	if matches := httpsPattern.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1], nil
	}

	// Try file:// pattern (for local repositories)
	if matches := filePattern.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1], nil
	}

	return "", errors.New(
		errors.ErrCodeInvalidRepository,
		fmt.Sprintf("Invalid Git URL format: %s", url),
		"URL must be in SSH format (git@host:owner/repo.git) or HTTPS format (https://host/owner/repo.git)",
	)
}

// InferName extracts the repository name from a URL
func InferName(url string) (string, error) {
	repoPath, err := ParseURL(url)
	if err != nil {
		return "", err
	}

	// Remove .git suffix if present
	repoPath = strings.TrimSuffix(repoPath, ".git")

	// Get the base name (last component of path)
	name := path.Base(repoPath)

	if name == "" || name == "." || name == "/" {
		return "", errors.New(
			errors.ErrCodeInvalidRepository,
			fmt.Sprintf("Could not infer repository name from URL: %s", url),
			"Provide a custom name: fa add <url> <name>",
		)
	}

	return name, nil
}

// ValidateURL validates that a URL is a valid Git URL format
func ValidateURL(url string) error {
	_, err := ParseURL(url)
	return err
}
