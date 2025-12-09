package config

// Config represents the workspace configuration
type Config struct {
	Workspace WorkspaceConfig `yaml:"workspace" toml:"workspace" json:"workspace"`
	Repos     []RepoConfig    `yaml:"repos" toml:"repos" json:"repos"`
	Settings  SettingsConfig  `yaml:"settings" toml:"settings" json:"settings"`
}

// WorkspaceConfig represents workspace-level configuration
type WorkspaceConfig struct {
	Name string `yaml:"name" toml:"name" json:"name"`
}

// RepoConfig represents a repository configuration entry
type RepoConfig struct {
	URL           string `yaml:"url" toml:"url" json:"url"`
	Name          string `yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty"`
	DefaultBranch string `yaml:"default_branch,omitempty" toml:"default_branch,omitempty" json:"default_branch,omitempty"`
}

// SettingsConfig represents workspace settings
type SettingsConfig struct {
	AutoCreateWorktree bool `yaml:"auto_create_worktree" toml:"auto_create_worktree" json:"auto_create_worktree"`
}

// DefaultConfig returns a default configuration
func DefaultConfig(workspaceName string) *Config {
	return &Config{
		Workspace: WorkspaceConfig{
			Name: workspaceName,
		},
		Repos: []RepoConfig{},
		Settings: SettingsConfig{
			AutoCreateWorktree: true,
		},
	}
}
