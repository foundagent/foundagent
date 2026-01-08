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

// TestDiscoverWorktrees_EmptyConfig tests discovery with no repos
func TestDiscoverWorktrees_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := &config.Config{
		Repos: []config.RepoConfig{},
	}

	worktrees, err := discoverWorktrees(ws, cfg, "")

	assert.NoError(t, err)
	assert.Empty(t, worktrees)
}

// TestDiscoverWorktrees_GetWorktreesError tests handling errors from GetWorktreesForRepo
func TestDiscoverWorktrees_GetWorktreesError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{
				Name: "nonexistent-repo",
				URL:  "https://github.com/test/repo.git",
			},
		},
	}

	// Don't create the worktrees directory - this will cause an error
	worktrees, err := discoverWorktrees(ws, cfg, "")

	// Error should be non-fatal, continue with empty list
	assert.NoError(t, err)
	assert.Empty(t, worktrees)
}

// TestDiscoverWorktrees_BranchFiltering tests filtering by branch
func TestDiscoverWorktrees_BranchFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main", "develop"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directories
	for _, branch := range []string{"main", "develop"} {
		wtPath := ws.WorktreePath("test-repo", branch)
		require.NoError(t, os.MkdirAll(wtPath, 0755))
	}

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{
				Name: "test-repo",
				URL:  "https://github.com/test/repo.git",
			},
		},
	}

	worktrees, err := discoverWorktrees(ws, cfg, "main")

	assert.NoError(t, err)
	assert.Len(t, worktrees, 1)
	if len(worktrees) > 0 {
		assert.Equal(t, "main", worktrees[0].Branch)
	}
}

// TestDiscoverWorktrees_MultipleRepos tests discovery across multiple repos
func TestDiscoverWorktrees_MultipleRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add multiple repositories
	for i := 1; i <= 2; i++ {
		repoName := "test-repo-" + string(rune('0'+i))
		repo := &workspace.Repository{
			Name:          repoName,
			URL:           "https://github.com/test/repo.git",
			DefaultBranch: "main",
			Worktrees:     []string{"main"},
			BareRepoPath:  ws.BareRepoPath(repoName),
		}
		require.NoError(t, ws.AddRepository(repo))

		// Create worktree directory
		wtPath := ws.WorktreePath(repoName, "main")
		require.NoError(t, os.MkdirAll(wtPath, 0755))
	}

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{Name: "test-repo-1", URL: "https://github.com/test/repo1.git"},
			{Name: "test-repo-2", URL: "https://github.com/test/repo2.git"},
		},
	}

	worktrees, err := discoverWorktrees(ws, cfg, "")

	assert.NoError(t, err)
	assert.Len(t, worktrees, 2)
}

// TestDiscoverWorktrees_Sorting tests that results are sorted by branch then repo
func TestDiscoverWorktrees_Sorting(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repositories with different branches
	for i := 1; i <= 2; i++ {
		repoName := "repo-" + string(rune('0'+i))
		repo := &workspace.Repository{
			Name:          repoName,
			URL:           "https://github.com/test/repo.git",
			DefaultBranch: "main",
			Worktrees:     []string{"develop", "main"},
			BareRepoPath:  ws.BareRepoPath(repoName),
		}
		require.NoError(t, ws.AddRepository(repo))

		// Create worktree directories
		for _, branch := range []string{"main", "develop"} {
			wtPath := ws.WorktreePath(repoName, branch)
			require.NoError(t, os.MkdirAll(wtPath, 0755))
		}
	}

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{Name: "repo-2", URL: "https://github.com/test/repo2.git"},
			{Name: "repo-1", URL: "https://github.com/test/repo1.git"},
		},
	}

	worktrees, err := discoverWorktrees(ws, cfg, "")

	assert.NoError(t, err)
	assert.Len(t, worktrees, 4)

	// Should be sorted by branch first (develop, main), then by repo (repo-1, repo-2)
	if len(worktrees) == 4 {
		assert.Equal(t, "develop", worktrees[0].Branch)
		assert.Equal(t, "repo-1", worktrees[0].Repo)
		assert.Equal(t, "develop", worktrees[1].Branch)
		assert.Equal(t, "repo-2", worktrees[1].Repo)
		assert.Equal(t, "main", worktrees[2].Branch)
		assert.Equal(t, "repo-1", worktrees[2].Repo)
		assert.Equal(t, "main", worktrees[3].Branch)
		assert.Equal(t, "repo-2", worktrees[3].Repo)
	}
}

// TestDiscoverWorktrees_FilterExcludesNonMatching tests that filter excludes non-matching branches
func TestDiscoverWorktrees_FilterExcludesNonMatching(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main", "feature", "develop"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directories
	for _, branch := range []string{"main", "feature", "develop"} {
		wtPath := ws.WorktreePath("test-repo", branch)
		require.NoError(t, os.MkdirAll(wtPath, 0755))
	}

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git"},
		},
	}

	worktrees, err := discoverWorktrees(ws, cfg, "feature")

	assert.NoError(t, err)
	assert.Len(t, worktrees, 1)
	if len(worktrees) > 0 {
		assert.Equal(t, "feature", worktrees[0].Branch)
	}
}

// TestDiscoverWorktrees_WithSymlinks tests discovery with symlinked worktree paths
func TestDiscoverWorktrees_WithSymlinks(t *testing.T) {
	t.Skip("Complex test - symlink handling varies")
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create actual worktree directory elsewhere
	actualPath := filepath.Join(tmpDir, "actual-worktree")
	require.NoError(t, os.MkdirAll(actualPath, 0755))

	// Create symlink in expected location
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(filepath.Dir(wtPath), 0755))
	require.NoError(t, os.Symlink(actualPath, wtPath))

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git"},
		},
	}

	worktrees, err := discoverWorktrees(ws, cfg, "")

	assert.NoError(t, err)
	assert.Len(t, worktrees, 1)
}
