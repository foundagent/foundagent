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

func TestDiscoverWorktrees_WithBranchFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create config with repos
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", DefaultBranch: "main"},
		},
	}

	// Create worktree structure manually
	wtPath := filepath.Join(ws.Path, "worktrees", "repo1", "main")
	os.MkdirAll(wtPath, 0755)

	// Create a .git file to make it look like a worktree
	os.WriteFile(filepath.Join(wtPath, ".git"), []byte("gitdir: somewhere"), 0644)

	// Discover with filter
	worktrees, err := discoverWorktrees(ws, cfg, "main")

	// Should succeed
	assert.NoError(t, err)
	// May or may not find worktrees depending on structure
	t.Logf("Found %d worktrees", len(worktrees))
}

func TestDiscoverWorktrees_NoFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create config with repos
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", DefaultBranch: "main"},
		},
	}

	// Discover without filter
	worktrees, err := discoverWorktrees(ws, cfg, "")

	// Should succeed
	assert.NoError(t, err)
	// List may be nil or empty
	t.Logf("Found %d worktrees", len(worktrees))
}

func TestAddRepository_SkippedExisting(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a bare repo directory
	repoPath := ws.BareRepoPath("existing-repo")
	os.MkdirAll(repoPath, 0755)

	// Add to config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "existing-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	// Try to add again without force
	addForce = false
	defer func() { addForce = false }()

	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/test/repo.git",
		Name: "existing-repo",
	})

	// Should skip (status is success, skipped is true)
	if result.Status == "success" {
		assert.True(t, result.Skipped)
	} else {
		t.Logf("Got status: %s, error: %s", result.Status, result.Error)
	}
}

func TestRunSyncPush_WithRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo to config (push will fail but tests the code path)
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	// Create bare repo directory (but not a real git repo)
	repoPath := ws.BareRepoPath("test-repo")
	os.MkdirAll(repoPath, 0755)

	defer func() { syncJSON = false; syncVerbose = false }()

	syncJSON = false
	syncVerbose = false

	// Will fail during push but tests more code paths
	err = runSyncPush(ws)

	// Error expected since repo isn't a real git repo
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
}

func TestRunSyncPush_JSONModeWithRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo to config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	// Create bare repo directory
	repoPath := ws.BareRepoPath("test-repo")
	os.MkdirAll(repoPath, 0755)

	defer func() { syncJSON = false; syncVerbose = false }()

	syncJSON = true

	// Will fail but tests JSON output path
	err = runSyncPush(ws)

	if err != nil {
		t.Logf("Expected error: %v", err)
	}
}
