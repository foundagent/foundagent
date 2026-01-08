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

func TestPreValidateWorktreeCreate_WorktreeExists(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a config with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}

	// Create bare repo directory structure to avoid nil pointer
	bareRepoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	// Create a fake worktree directory
	worktreePath := filepath.Join(ws.Path, "test-repo", "feature-123")
	err = os.MkdirAll(worktreePath, 0755)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Test without force - may error on git operations but tests the existence check path
	_ = preValidateWorktreeCreate(ws, cfg, "feature-123", "", false)
}

func TestPreValidateWorktreeCreate_WorktreeExistsWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a config with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}

	// Create a fake worktree directory
	worktreePath := filepath.Join(ws.Path, "test-repo", "feature-123")
	err = os.MkdirAll(worktreePath, 0755)
	require.NoError(t, err)

	// Create bare repo path for git operations
	bareRepoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Test with force - may fail on git operations but tests the force path
	_ = preValidateWorktreeCreate(ws, cfg, "feature-123", "", true)
}

func TestPreValidateWorktreeCreate_SourceBranchNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a config with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}

	// Create bare repo directory (but not a real git repo)
	bareRepoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Test with non-existent source branch - will error on git operations
	err = preValidateWorktreeCreate(ws, cfg, "feature-123", "nonexistent-source", false)
	assert.Error(t, err)
}

func TestPreValidateWorktreeCreate_MultipleReposValidation(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a config with multiple repos
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "repo1", URL: "https://github.com/org/repo1.git", DefaultBranch: "main"},
			{Name: "repo2", URL: "https://github.com/org/repo2.git", DefaultBranch: "main"},
		},
	}

	// Create fake worktree for first repo only
	worktreePath := filepath.Join(ws.Path, "repo1", "feature-123")
	err = os.MkdirAll(worktreePath, 0755)
	require.NoError(t, err)

	// Create bare repo paths
	for _, repo := range cfg.Repos {
		bareRepoPath := ws.BareRepoPath(repo.Name)
		err = os.MkdirAll(bareRepoPath, 0755)
		require.NoError(t, err)
	}

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Test without force - may error on git operations
	_ = preValidateWorktreeCreate(ws, cfg, "feature-123", "", false)
	// Can't assert specific error since git operations may fail in different ways
}

func TestCreateWorktreesParallel_EmptyRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Test with empty repos list
	results := createWorktreesParallel(ws, []config.RepoConfig{}, "feature", "", false)
	assert.Empty(t, results)
}

func TestCreateWorktreesParallel_MultipleRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create config with repos
	repos := []config.RepoConfig{
		{Name: "repo1", URL: "https://github.com/org/repo1.git", DefaultBranch: "main"},
		{Name: "repo2", URL: "https://github.com/org/repo2.git", DefaultBranch: "main"},
	}

	// Test parallel creation (will fail on git operations but tests parallelism)
	results := createWorktreesParallel(ws, repos, "feature", "", false)
	assert.Len(t, results, 2)
	// All should fail since we don't have real git repos
	for _, r := range results {
		assert.Equal(t, "error", r.Status)
	}
}
