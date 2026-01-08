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

func TestRunReconcile_UpToDate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Reconcile with empty config (no repos to clone)
	addJSON = false
	err = runReconcile(ws)
	assert.NoError(t, err)

	// Reset flag
	addJSON = false
}

func TestRunReconcile_WithReposToClone(t *testing.T) {
	t.Skip("Skipping test that would attempt network git clone operations")
}

func TestRunReconcile_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Reconcile in JSON mode
	addJSON = true
	err = runReconcile(ws)
	assert.NoError(t, err)

	// Reset flag
	addJSON = false
}

func TestAddRepositories_Single(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Test adding single invalid repo
	repos := []repoToAdd{
		{URL: "not-a-url", Name: ""},
	}

	results := addRepositories(ws, repos)
	assert.Len(t, results, 1)
	assert.Equal(t, "error", results[0].Status)
	assert.NotEmpty(t, results[0].Error)
}

func TestAddRepositories_Multiple(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Test adding multiple repos with invalid URLs (fail validation, not clone)
	repos := []repoToAdd{
		{URL: "invalid-url-1", Name: "repo1"},
		{URL: "invalid-url-2", Name: "repo2"},
	}

	results := addRepositories(ws, repos)
	assert.Len(t, results, 2)
	// Both should fail validation
	for _, r := range results {
		assert.Equal(t, "error", r.Status)
		assert.NotEmpty(t, r.Error)
	}
}

func TestAddRepository_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Test invalid URL
	repo := repoToAdd{URL: "not-a-url", Name: ""}
	result := addRepository(ws, repo)

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "Invalid")
}

func TestAddRepository_WithCustomName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Test with custom name using invalid URL (fails validation, not clone)
	repo := repoToAdd{URL: "not-a-valid-url", Name: "custom-name"}
	result := addRepository(ws, repo)

	assert.Equal(t, "error", result.Status) // URL validation fails
	assert.NotEmpty(t, result.Error)
}

func TestAddRepository_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Add a repository to state
	repo := &workspace.Repository{
		Name:          "existing-repo",
		URL:           "https://github.com/org/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("existing-repo"),
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Try to add again without force
	addForce = false
	repoToAddAgain := repoToAdd{URL: "https://github.com/org/repo.git", Name: "existing-repo"}
	result := addRepository(ws, repoToAddAgain)

	assert.Equal(t, "success", result.Status)
	assert.True(t, result.Skipped)
	assert.Equal(t, "existing-repo", result.Name)

	// Reset flag
	addForce = false
}

func TestAddRepository_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Add a repository and create its bare repo directory
	repo := &workspace.Repository{
		Name:          "existing-repo",
		URL:           "https://github.com/org/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("existing-repo"),
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Create the bare repo directory
	err = os.MkdirAll(repo.BareRepoPath, 0755)
	require.NoError(t, err)

	// Try to add again with force using invalid URL
	addForce = true
	repoToAddAgain := repoToAdd{URL: "invalid-url", Name: "existing-repo"}
	result := addRepository(ws, repoToAddAgain)

	// Should fail on URL validation
	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)

	// Reset flag
	addForce = false
}

func TestAddRepository_InferName(t *testing.T) {
	t.Skip("Skipping test that would attempt network git clone operations")
}

func TestRunAdd_NoWorkspace(t *testing.T) {
	// Save original working directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Change to temp dir with no workspace
	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Try to run add - should fail with no workspace
	addJSON = false
	err = runAdd(addCmd, []string{"https://github.com/org/repo.git"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")

	// Reset flag
	addJSON = false
}

func TestRunAdd_ReconcileMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original working directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Run add with no args (reconcile mode)
	addJSON = false
	err = runAdd(addCmd, []string{})
	assert.NoError(t, err)

	// Reset flag
	addJSON = false
}

func TestRunAdd_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original working directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to add invalid URL
	addJSON = false
	err = runAdd(addCmd, []string{"not-a-valid-url"})
	assert.Error(t, err)

	// Reset flag
	addJSON = false
}

func TestRunAdd_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original working directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Add invalid URL in JSON mode - single result doesn't return error
	addJSON = true
	err = runAdd(addCmd, []string{"invalid-url"})
	// JSON mode with single result prints JSON but doesn't error
	assert.NoError(t, err)

	// Reset flag
	addJSON = false
}

func TestRunAdd_MultipleInvalidURLs(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original working directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to add multiple invalid URLs
	addJSON = false
	err = runAdd(addCmd, []string{"invalid-1", "invalid-2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add")

	// Reset flag
	addJSON = false
}
