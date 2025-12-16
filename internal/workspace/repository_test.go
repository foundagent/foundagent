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
