package config

import "fmt"

// DefaultTemplate returns a default YAML config with comments
func DefaultTemplate(workspaceName string) string {
	return fmt.Sprintf(`# Foundagent Workspace Configuration
# This file defines your multi-repository workspace

workspace:
  # Workspace name
  name: %s

# List of repositories in this workspace
repos: []
  # Example repository entry:
  # - url: git@github.com:org/my-repo.git
  #   name: my-repo              # Optional: override inferred name
  #   default_branch: main       # Optional: override detected default branch

# Workspace settings
settings:
  # Automatically create a worktree for the default branch when adding a repo
  auto_create_worktree: true
`, workspaceName)
}

// AddRepo adds a repository to the configuration
func AddRepo(config *Config, url, name, defaultBranch string) {
	repo := RepoConfig{
		URL:           url,
		Name:          name,
		DefaultBranch: defaultBranch,
	}

	// Check if repo already exists
	for i, r := range config.Repos {
		if r.Name == name {
			// Update existing entry
			config.Repos[i] = repo
			return
		}
	}

	// Add new entry
	config.Repos = append(config.Repos, repo)
}

// RemoveRepo removes a repository from the configuration
func RemoveRepo(config *Config, name string) bool {
	for i, r := range config.Repos {
		if r.Name == name {
			config.Repos = append(config.Repos[:i], config.Repos[i+1:]...)
			return true
		}
	}
	return false
}

// HasRepo checks if a repository exists in the configuration
func HasRepo(config *Config, name string) bool {
	for _, r := range config.Repos {
		if r.Name == name {
			return true
		}
	}
	return false
}

// GetRepo retrieves a repository from the configuration
func GetRepo(config *Config, name string) *RepoConfig {
	for _, r := range config.Repos {
		if r.Name == name {
			return &r
		}
	}
	return nil
}
