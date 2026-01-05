package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullAllWorktrees_NoReposEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	results, err := ws.PullAllWorktrees("main", false, false)
	
	// Should succeed with empty results
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestPullAllWorktrees_NonexistentWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository but don't create worktree
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	results, err := ws.PullAllWorktrees("feature", false, false)
	
	// Should succeed with skipped result
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "skipped", results[0].Status)
}

func TestPullAllWorktrees_WithStash(t *testing.T) {
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

	results, err := ws.PullAllWorktrees("main", true, false)
	
	// Should succeed but skip non-git worktree
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "skipped", results[0].Status)
}

func TestPullAllWorktrees_VerboseMode(t *testing.T) {
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

	results, err := ws.PullAllWorktrees("main", false, true)
	
	// Should succeed (verbose just affects output)
	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestPushAllRepos_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	results, err := ws.PushAllRepos(false)
	
	// Should succeed with empty results
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestPushAllRepos_NonexistentWT(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository but don't create worktree
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	results, err := ws.PushAllRepos(false)
	
	// Should succeed with empty or skipped results
	require.NoError(t, err)
	_ = results
}

func TestPushAllRepos_Verbose(t *testing.T) {
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

	results, err := ws.PushAllRepos(true)
	
	// Should succeed (verbose just affects output)
	require.NoError(t, err)
	_ = results
}

func TestRemoveRepo_SuccessfulRemoval(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create bare repo directory
	bareRepoPath := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(bareRepoPath, 0755))

	// Remove the repo
	result := ws.RemoveRepo("test-repo", true, false)
	
	// Should succeed
	assert.Empty(t, result.Error)
}

func TestDetectWorktreeStatus_ValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a directory (not a real git repo)
	repoPath := filepath.Join(tmpDir, "test-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	status := ws.detectWorktreeStatus(repoPath, false)
	
	// Status should be returned (may be error status for non-git dir)
	assert.NotNil(t, status)
}

func TestCreate_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	
	// Create once
	err = ws.Create(false)
	require.NoError(t, err)

	// Try to create again
	err = ws.Create(false)
	
	// Should error - already exists
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCreate_WithVSCode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	
	// Create with VS Code workspace file
	err = ws.Create(true)
	require.NoError(t, err)

	// Verify workspace file exists
	wsPath := ws.VSCodeWorkspacePath()
	_, err = os.Stat(wsPath)
	assert.NoError(t, err)
}

