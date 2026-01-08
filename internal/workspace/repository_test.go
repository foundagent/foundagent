package workspace

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeBasePath(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)

	path := ws.WorktreeBasePath("myrepo")
	expectedPath := filepath.Join(tmpDir, "test-workspace", "repos", "myrepo", "worktrees")
	assert.Equal(t, expectedPath, path)
}

func TestWorktreePath(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)

	path := ws.WorktreePath("myrepo", "feature-branch")
	expectedPath := filepath.Join(tmpDir, "test-workspace", "repos", "myrepo", "worktrees", "feature-branch")
	assert.Equal(t, expectedPath, path)
}

func TestGetRepository(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository to state
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/org/test-repo.git",
		DefaultBranch: "main",
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Get repository
	retrieved, err := ws.GetRepository("test-repo")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "test-repo", retrieved.Name)
	assert.Equal(t, "https://github.com/org/test-repo.git", retrieved.URL)
	assert.Equal(t, "main", retrieved.DefaultBranch)

	// Get non-existent repository
	nonExistent, err := ws.GetRepository("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, nonExistent)
}

func TestHasRepository(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Should not have repo initially
	has, err := ws.HasRepository("test-repo")
	require.NoError(t, err)
	assert.False(t, has)

	// Add repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/org/test-repo.git",
		DefaultBranch: "main",
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Should have repo now
	has, err = ws.HasRepository("test-repo")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestAddRepository_NilState(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository when state.Repositories is nil (normal case)
	repo := &Repository{
		Name:          "first-repo",
		URL:           "https://github.com/org/first-repo.git",
		DefaultBranch: "main",
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Verify it was added
	retrieved, err := ws.GetRepository("first-repo")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "first-repo", retrieved.Name)
}

func TestAddRepository_Multiple(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add multiple repositories
	repo1 := &Repository{
		Name:          "repo1",
		URL:           "https://github.com/org/repo1.git",
		DefaultBranch: "main",
	}
	repo2 := &Repository{
		Name:          "repo2",
		URL:           "https://github.com/org/repo2.git",
		DefaultBranch: "master",
	}

	err = ws.AddRepository(repo1)
	require.NoError(t, err)
	err = ws.AddRepository(repo2)
	require.NoError(t, err)

	// Verify both were added
	has1, err := ws.HasRepository("repo1")
	require.NoError(t, err)
	assert.True(t, has1)

	has2, err := ws.HasRepository("repo2")
	require.NoError(t, err)
	assert.True(t, has2)
}

func TestBareRepoPath(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)

	path := ws.BareRepoPath("myrepo")
	expectedPath := filepath.Join(tmpDir, "test-workspace", "repos", "myrepo", ".bare")
	assert.Equal(t, expectedPath, path)
}
