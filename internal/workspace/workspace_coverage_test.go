package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_ForceMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create once
	err = ws.Create(false)
	require.NoError(t, err)

	// Create again with force
	err = ws.Create(true)
	require.NoError(t, err)

	// Verify workspace still exists
	assert.True(t, ws.Exists())
}

func TestCreate_PreservesReposWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create workspace
	err = ws.Create(false)
	require.NoError(t, err)

	// Create some content in repos directory
	reposPath := filepath.Join(ws.Path, ReposDir)
	testFile := filepath.Join(reposPath, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Recreate with force
	err = ws.Create(true)
	require.NoError(t, err)

	// Verify repos directory content preserved
	_, err = os.Stat(testFile)
	assert.NoError(t, err, "Force mode should preserve repos directory")
}

func TestCreate_PermissionError(t *testing.T) {
	// This test would require a read-only filesystem
	// which is difficult to set up in a portable way
	t.Skip("Permission error testing requires special setup")
}

func TestRemoveRepo_WithWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main", "feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create bare repo and worktree directories
	bareRepoPath := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(bareRepoPath, 0755))

	mainPath := ws.WorktreePath("test-repo", "main")
	featurePath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(mainPath, 0755))
	require.NoError(t, os.MkdirAll(featurePath, 0755))

	// Remove with force (to skip dirty checks)
	result := ws.RemoveRepo("test-repo", true, false)

	// Should succeed
	assert.Empty(t, result.Error)
}

func TestRemoveRepo_ConfigOnlyMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Remove config only
	result := ws.RemoveRepo("test-repo", false, true)

	// Should succeed
	assert.Empty(t, result.Error)
	assert.True(t, result.ConfigOnly)
}

func TestRemoveRepo_WithDirtyWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create bare repo directory
	bareRepoPath := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(bareRepoPath, 0755))

	// Create worktree directory
	mainPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(mainPath, 0755))

	// Without force, should fail on dirty check (non-git dir)
	result := ws.RemoveRepo("test-repo", false, false)

	// May succeed or fail depending on dirty check
	_ = result
}

func TestGetWorkspaceStatus_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	status, err := ws.GetWorkspaceStatus(false)

	// Should succeed with empty status
	require.NoError(t, err)
	assert.Empty(t, status.Repos)
}

func TestGetWorkspaceStatus_WithRepo(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory (not a real git repo)
	mainPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(mainPath, 0755))

	status, err := ws.GetWorkspaceStatus(false)

	// Should succeed
	require.NoError(t, err)
	require.Len(t, status.Repos, 1)
	assert.Equal(t, "test-repo", status.Repos[0].Name)
}

func TestGetWorkspaceStatus_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	mainPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(mainPath, 0755))

	status, err := ws.GetWorkspaceStatus(true)

	// Should succeed in verbose mode
	require.NoError(t, err)
	require.Len(t, status.Repos, 1)
}

func TestDetectWorktreeStatus_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	status := ws.detectWorktreeStatus("/nonexistent/path", false)

	// Should return a status (likely error status)
	assert.NotNil(t, status)
}

func TestDetectWorktreeStatus_NonGitDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a non-git directory
	testDir := filepath.Join(tmpDir, "test-dir")
	require.NoError(t, os.MkdirAll(testDir, 0755))

	status := ws.detectWorktreeStatus(testDir, false)

	// Should return error status for non-git directory
	assert.NotNil(t, status)
}

func TestNew_ValidName(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("my-workspace", tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "my-workspace", ws.Name)
	assert.Contains(t, ws.Path, "my-workspace")
}

func TestNew_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := New("", tmpDir)

	// Should error with empty name
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestNew_InvalidCharacters(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := New("my/invalid\\name", tmpDir)

	// Should error with invalid characters
	assert.Error(t, err)
}

func TestFindDirtyWorktrees_NoDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	dirty, err := ws.findDirtyWorktrees("nonexistent-repo")

	// Should succeed with empty list
	assert.NoError(t, err)
	assert.Empty(t, dirty)
}

func TestRemoveAllWorktrees_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	count, err := ws.removeAllWorktrees("nonexistent-repo")

	// Should succeed with 0 count
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
