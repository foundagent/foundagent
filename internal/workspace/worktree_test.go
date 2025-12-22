package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeExists(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Non-existent worktree
	exists, err := ws.WorktreeExists("test-repo", "feature-branch")
	require.NoError(t, err)
	assert.False(t, exists)

	// Create a worktree directory manually
	wtPath := ws.WorktreePath("test-repo", "feature-branch")
	err = os.MkdirAll(wtPath, 0755)
	require.NoError(t, err)

	// Now it should exist
	exists, err = ws.WorktreeExists("test-repo", "feature-branch")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestGetWorktreesForRepo_Method(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create worktree directories
	wtBase := ws.WorktreeBasePath("test-repo")
	err = os.MkdirAll(filepath.Join(wtBase, "main"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(wtBase, "feature"), 0755)
	require.NoError(t, err)

	// Get worktrees
	worktrees, err := ws.GetWorktreesForRepo("test-repo")
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)
	assert.Contains(t, worktrees, "main")
	assert.Contains(t, worktrees, "feature")

	// Get worktrees for non-existent repo
	empty, err := ws.GetWorktreesForRepo("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestGetAllWorktrees(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create worktrees for multiple repos
	err = os.MkdirAll(ws.WorktreePath("repo1", "main"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(ws.WorktreePath("repo2", "feature"), 0755)
	require.NoError(t, err)

	// Get all worktrees
	allWorktrees, err := ws.GetAllWorktrees()
	require.NoError(t, err)
	assert.Len(t, allWorktrees, 2)
	assert.Contains(t, allWorktrees, "repo1")
	assert.Contains(t, allWorktrees, "repo2")
	assert.Contains(t, allWorktrees["repo1"], "main")
	assert.Contains(t, allWorktrees["repo2"], "feature")
}

func TestFindWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Save current directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	t.Run("from outside repos directory", func(t *testing.T) {
		// When called from outside repos, should return empty string
		path, err := ws.FindWorktree("feature")
		require.NoError(t, err)
		assert.Empty(t, path)
	})

	t.Run("from inside repos directory", func(t *testing.T) {
		// Create a repo worktree structure
		repoWorktreeDir := filepath.Join(ws.Path, ReposDir, "myrepo", WorktreesDir, "main")
		err := os.MkdirAll(repoWorktreeDir, 0755)
		require.NoError(t, err)

		// Change to the worktree directory
		err = os.Chdir(repoWorktreeDir)
		require.NoError(t, err)

		// Get the absolute workspace path and resolve symlinks (macOS /var -> /private/var)
		absWsPath, err := filepath.Abs(ws.Path)
		require.NoError(t, err)
		absWsPath, err = filepath.EvalSymlinks(absWsPath)
		require.NoError(t, err)

		// Create a new workspace instance with absolute path
		wsAbs := &Workspace{
			Name: ws.Name,
			Path: absWsPath,
		}

		// Should return the path to the requested branch
		path, err := wsAbs.FindWorktree("feature-123")
		require.NoError(t, err)
		expected := wsAbs.WorktreePath("myrepo", "feature-123")
		assert.Equal(t, expected, path)
	})
}

func TestGetWorktreesForRepo_Function(t *testing.T) {
	tmpDir := t.TempDir()

	// Create worktree directories using the expected structure
	wtBase := filepath.Join(tmpDir, ReposDir, "test-repo", WorktreesDir)
	err := os.MkdirAll(filepath.Join(wtBase, "main"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(wtBase, "feature"), 0755)
	require.NoError(t, err)

	// Use package-level function
	details, err := GetWorktreesForRepo(tmpDir, "test-repo")
	require.NoError(t, err)
	require.Len(t, details, 2)

	// Check details
	assert.Equal(t, "test-repo", details[0].Repo)
	assert.NotEmpty(t, details[0].Branch)
	assert.NotEmpty(t, details[0].Path)
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "testfile")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Test existing path
	exists := PathExists(testFile)
	assert.True(t, exists)

	// Test non-existent path
	exists = PathExists(filepath.Join(tmpDir, "nonexistent"))
	assert.False(t, exists)
}
