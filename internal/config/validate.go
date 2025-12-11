package config

import (
	"fmt"
	"strings"

	"github.com/foundagent/foundagent/internal/errors"
	"github.com/foundagent/foundagent/internal/git"
)

// Validate validates the configuration
func Validate(config *Config) error {
	if config == nil {
		return errors.New(
			errors.ErrCodeInvalidConfig,
			"Config is nil",
			"Ensure config file is properly formatted",
		)
	}

	// Validate workspace name
	if strings.TrimSpace(config.Workspace.Name) == "" {
		return errors.New(
			errors.ErrCodeInvalidConfig,
			"Workspace name cannot be empty",
			"Set workspace.name in your config file",
		)
	}

	// Validate repos
	repoNames := make(map[string]bool)
	repoURLs := make(map[string]string) // url -> first name that used it

	for i, repo := range config.Repos {
		// Validate URL format
		if err := git.ValidateURL(repo.URL); err != nil {
			return errors.New(
				errors.ErrCodeInvalidConfig,
				fmt.Sprintf("Invalid URL in repos[%d]: %s", i, repo.URL),
				"Ensure URL is in format git@host:owner/repo.git or https://host/owner/repo.git",
			)
		}

		// Infer name if not provided
		name := repo.Name
		if name == "" {
			inferredName, err := git.InferName(repo.URL)
			if err != nil {
				return errors.New(
					errors.ErrCodeInvalidConfig,
					fmt.Sprintf("Could not infer name for repos[%d]: %s", i, repo.URL),
					"Provide an explicit 'name' field for this repo",
				)
			}
			name = inferredName
			// Update the config with inferred name for consistency
			config.Repos[i].Name = name
		}

		// Check for duplicate names
		if repoNames[name] {
			return errors.New(
				errors.ErrCodeInvalidConfig,
				fmt.Sprintf("Duplicate repository name: %s", name),
				"Each repository must have a unique name",
			)
		}
		repoNames[name] = true

		// Check for duplicate URLs (warn, not error)
		if firstUser, exists := repoURLs[repo.URL]; exists {
			fmt.Printf("Warning: Repository URL %s is used by both '%s' and '%s'. This may cause confusion.\n",
				repo.URL, firstUser, name)
		} else {
			repoURLs[repo.URL] = name
		}
	}

	return nil
}
