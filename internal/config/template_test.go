package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRepo_NewRepo(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos:     []RepoConfig{},
	}

	AddRepo(cfg, "https://github.com/test/repo1.git", "repo1", "main")

	assert.Len(t, cfg.Repos, 1)
	assert.Equal(t, "repo1", cfg.Repos[0].Name)
	assert.Equal(t, "https://github.com/test/repo1.git", cfg.Repos[0].URL)
	assert.Equal(t, "main", cfg.Repos[0].DefaultBranch)
}

func TestAddRepo_UpdateExisting(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
		},
	}

	AddRepo(cfg, "https://github.com/test/repo1-new.git", "repo1", "develop")

	assert.Len(t, cfg.Repos, 1)
	assert.Equal(t, "https://github.com/test/repo1-new.git", cfg.Repos[0].URL)
	assert.Equal(t, "develop", cfg.Repos[0].DefaultBranch)
}

func TestAddRepo_MultipleRepos(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos:     []RepoConfig{},
	}

	AddRepo(cfg, "https://github.com/test/repo1.git", "repo1", "main")
	AddRepo(cfg, "https://github.com/test/repo2.git", "repo2", "master")

	assert.Len(t, cfg.Repos, 2)
	assert.Equal(t, "repo1", cfg.Repos[0].Name)
	assert.Equal(t, "repo2", cfg.Repos[1].Name)
}

func TestRemoveRepo_Exists(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
			{URL: "https://github.com/test/repo2.git", Name: "repo2", DefaultBranch: "main"},
		},
	}

	result := RemoveRepo(cfg, "repo1")

	assert.True(t, result)
	assert.Len(t, cfg.Repos, 1)
	assert.Equal(t, "repo2", cfg.Repos[0].Name)
}

func TestRemoveRepo_NotExists(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
		},
	}

	result := RemoveRepo(cfg, "repo-nonexistent")

	assert.False(t, result)
	assert.Len(t, cfg.Repos, 1)
}

func TestRemoveRepo_LastRepo(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
		},
	}

	result := RemoveRepo(cfg, "repo1")

	assert.True(t, result)
	assert.Empty(t, cfg.Repos)
}

func TestHasRepo_Exists(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
		},
	}

	result := HasRepo(cfg, "repo1")

	assert.True(t, result)
}

func TestHasRepo_NotExists(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
		},
	}

	result := HasRepo(cfg, "repo-nonexistent")

	assert.False(t, result)
}

func TestHasRepo_EmptyConfig(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos:     []RepoConfig{},
	}

	result := HasRepo(cfg, "repo1")

	assert.False(t, result)
}

func TestGetRepo_Exists(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
			{URL: "https://github.com/test/repo2.git", Name: "repo2", DefaultBranch: "develop"},
		},
	}

	repo := GetRepo(cfg, "repo2")

	assert.NotNil(t, repo)
	assert.Equal(t, "repo2", repo.Name)
	assert.Equal(t, "https://github.com/test/repo2.git", repo.URL)
	assert.Equal(t, "develop", repo.DefaultBranch)
}

func TestGetRepo_NotExists(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos: []RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
		},
	}

	repo := GetRepo(cfg, "repo-nonexistent")

	assert.Nil(t, repo)
}

func TestGetRepo_EmptyConfig(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: "test"},
		Repos:     []RepoConfig{},
	}

	repo := GetRepo(cfg, "repo1")

	assert.Nil(t, repo)
}

func TestDefaultTemplate_ContainsWorkspaceName(t *testing.T) {
	template := DefaultTemplate("my-workspace")

	assert.Contains(t, template, "my-workspace")
	assert.Contains(t, template, "workspace:")
	assert.Contains(t, template, "repos:")
	assert.Contains(t, template, "settings:")
}

func TestDefaultTemplate_ValidYAML(t *testing.T) {
	template := DefaultTemplate("test-ws")

	// Should be valid YAML structure
	assert.Contains(t, template, "name: test-ws")
	assert.Contains(t, template, "auto_create_worktree: true")
}
