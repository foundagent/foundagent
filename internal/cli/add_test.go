package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAddArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []repoToAdd
	}{
		{
			name: "single URL",
			args: []string{"https://github.com/org/repo.git"},
			expected: []repoToAdd{
				{URL: "https://github.com/org/repo.git", Name: ""},
			},
		},
		{
			name: "URL with custom name",
			args: []string{"https://github.com/org/repo.git", "my-name"},
			expected: []repoToAdd{
				{URL: "https://github.com/org/repo.git", Name: "my-name"},
			},
		},
		{
			name: "multiple URLs",
			args: []string{
				"https://github.com/org/repo1.git",
				"https://github.com/org/repo2.git",
			},
			expected: []repoToAdd{
				{URL: "https://github.com/org/repo1.git", Name: ""},
				{URL: "https://github.com/org/repo2.git", Name: ""},
			},
		},
		{
			name: "URL, name, URL pattern",
			args: []string{
				"https://github.com/org/repo1.git",
				"custom-name",
				"https://github.com/org/repo2.git",
			},
			expected: []repoToAdd{
				{URL: "https://github.com/org/repo1.git", Name: "custom-name"},
				{URL: "https://github.com/org/repo2.git", Name: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAddArgs(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigUpdateOnAdd(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Initially, config should have no repos
	cfg, err := config.Load(ws.Path)
	require.NoError(t, err)
	assert.Empty(t, cfg.Repos)

	// Note: We can't actually test the full add flow without real git repos
	// But we can test the config manipulation logic
	config.AddRepo(cfg, "https://github.com/org/repo1.git", "repo1", "main")
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Verify config was updated
	cfg, err = config.Load(ws.Path)
	require.NoError(t, err)
	assert.Len(t, cfg.Repos, 1)
	assert.Equal(t, "repo1", cfg.Repos[0].Name)
	assert.Equal(t, "https://github.com/org/repo1.git", cfg.Repos[0].URL)
	assert.Equal(t, "main", cfg.Repos[0].DefaultBranch)
}

func TestReconcileFlow(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Add repos to config
	cfg, err := config.Load(ws.Path)
	require.NoError(t, err)
	
	config.AddRepo(cfg, "https://github.com/org/repo1.git", "repo1", "main")
	config.AddRepo(cfg, "https://github.com/org/repo2.git", "repo2", "master")
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Reconcile should find 2 repos to clone
	result, err := workspace.Reconcile(ws)
	require.NoError(t, err)
	assert.Len(t, result.ReposToClone, 2)
	assert.Empty(t, result.ReposUpToDate)
	assert.Empty(t, result.ReposStale)

	// Simulate that repo1 was cloned
	repo1 := &workspace.Repository{
		Name:          "repo1",
		URL:           "https://github.com/org/repo1.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("repo1"),
	}
	err = ws.AddRepository(repo1)
	require.NoError(t, err)

	// Reconcile again
	result, err = workspace.Reconcile(ws)
	require.NoError(t, err)
	assert.Len(t, result.ReposToClone, 1)
	assert.Len(t, result.ReposUpToDate, 1)
	assert.Empty(t, result.ReposStale)
	assert.Equal(t, "repo2", result.ReposToClone[0].Name)
}

func TestConfigTemplateGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-template", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Verify config file exists and has template
	configPath := filepath.Join(ws.Path, ".foundagent.yaml")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# Foundagent Workspace Configuration")
	assert.Contains(t, contentStr, "test-template")
	assert.Contains(t, contentStr, "repos: []")
	assert.Contains(t, contentStr, "auto_create_worktree: true")
	assert.Contains(t, contentStr, "# Example repository entry:")
}
