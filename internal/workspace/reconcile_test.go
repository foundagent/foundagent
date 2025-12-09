package workspace

import (
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Load and modify config
	cfg, err := config.Load(ws.Path)
	require.NoError(t, err)
	
	cfg.Repos = []config.RepoConfig{
		{URL: "https://github.com/org/repo1.git", Name: "repo1", DefaultBranch: "main"},
		{URL: "https://github.com/org/repo2.git", Name: "repo2", DefaultBranch: "master"},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Reconcile - should find 2 repos to clone
	result, err := Reconcile(ws)
	require.NoError(t, err)
	assert.Len(t, result.ReposToClone, 2)
	assert.Empty(t, result.ReposUpToDate)
	assert.Empty(t, result.ReposStale)

	// Simulate adding repo1 to state
	repo1 := &Repository{
		Name:          "repo1",
		URL:           "https://github.com/org/repo1.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("repo1"),
	}
	err = ws.AddRepository(repo1)
	require.NoError(t, err)

	// Reconcile again - should find 1 repo to clone, 1 up-to-date
	result, err = Reconcile(ws)
	require.NoError(t, err)
	assert.Len(t, result.ReposToClone, 1)
	assert.Len(t, result.ReposUpToDate, 1)
	assert.Empty(t, result.ReposStale)
	assert.Equal(t, "repo2", result.ReposToClone[0].Name)
	assert.Contains(t, result.ReposUpToDate, "repo1")

	// Simulate adding repo2 to state
	repo2 := &Repository{
		Name:          "repo2",
		URL:           "https://github.com/org/repo2.git",
		DefaultBranch: "master",
		BareRepoPath:  ws.BareRepoPath("repo2"),
	}
	err = ws.AddRepository(repo2)
	require.NoError(t, err)

	// Reconcile again - all repos up-to-date
	result, err = Reconcile(ws)
	require.NoError(t, err)
	assert.Empty(t, result.ReposToClone)
	assert.Len(t, result.ReposUpToDate, 2)
	assert.Empty(t, result.ReposStale)
}

func TestReconcileStaleRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Add repo to state but not config
	repo := &Repository{
		Name:          "orphan-repo",
		URL:           "https://github.com/org/orphan.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("orphan-repo"),
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Reconcile - should find stale repo
	result, err := Reconcile(ws)
	require.NoError(t, err)
	assert.Empty(t, result.ReposToClone)
	assert.Empty(t, result.ReposUpToDate)
	assert.Len(t, result.ReposStale, 1)
	assert.Contains(t, result.ReposStale, "orphan-repo")
}

func TestReconcileInferNames(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Load and modify config with repos without explicit names
	cfg, err := config.Load(ws.Path)
	require.NoError(t, err)
	
	cfg.Repos = []config.RepoConfig{
		{URL: "https://github.com/org/my-repo.git"}, // Name will be inferred
	}
	
	// Validate should infer the name
	err = config.Validate(cfg)
	require.NoError(t, err)
	assert.Equal(t, "my-repo", cfg.Repos[0].Name)

	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Reconcile - should work with inferred name
	result, err := Reconcile(ws)
	require.NoError(t, err)
	assert.Len(t, result.ReposToClone, 1)
	assert.Equal(t, "my-repo", result.ReposToClone[0].Name)
}

func TestPrintReconcileResult(t *testing.T) {
	// This is more of a smoke test - just make sure it doesn't panic
	result := &ReconcileResult{
		ReposToClone: []config.RepoConfig{
			{URL: "https://github.com/org/repo.git", Name: "repo"},
		},
		ReposUpToDate: []string{"existing-repo"},
		ReposStale:    []string{"orphan-repo"},
	}

	// Should not panic
	PrintReconcileResult(result)
}
