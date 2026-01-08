package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreesCheck_AllValid(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktrees
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create valid worktree with .git file
	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))
	gitFile := filepath.Join(wtPath, ".git")
	require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: ../../../.bare/worktrees/main"), 0644))

	check := WorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusPass, result.Status)
	assert.Contains(t, result.Message, "All 1 worktrees valid")
}

func TestWorktreesCheck_MissingWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktrees but don't create directories
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main", "develop"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	check := WorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Message, "issue(s)")
	assert.False(t, result.Fixable)
}

func TestWorktreesCheck_InvalidGitWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktrees
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory but no .git file
	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	check := WorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Message, "issue(s)")
}

func TestWorktreesCheck_NoWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with no worktrees
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	check := WorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusPass, result.Status)
	assert.Contains(t, result.Message, "All 0 worktrees valid")
}

func TestOrphanedWorktreesCheck_NoOrphans(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktrees
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create matching worktree directory
	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	check := OrphanedWorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusPass, result.Status)
	assert.Contains(t, result.Message, "No orphaned worktrees")
}

func TestOrphanedWorktreesCheck_WithOrphanedWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with one worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create matching worktree
	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	// Create orphaned worktree
	orphanedPath := ws.WorktreePath("test-repo", "orphaned")
	require.NoError(t, os.MkdirAll(orphanedPath, 0755))

	check := OrphanedWorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusWarn, result.Status)
	assert.Contains(t, result.Message, "orphaned")
	assert.True(t, result.Fixable)
}

func TestOrphanedWorktreesCheck_WithOrphanedRepo(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create orphaned repo directory (not in state)
	orphanedRepoPath := filepath.Join(ws.Path, "repos", "orphaned-repo")
	require.NoError(t, os.MkdirAll(orphanedRepoPath, 0755))

	check := OrphanedWorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusWarn, result.Status)
	assert.Contains(t, result.Message, "orphaned")
	assert.True(t, result.Fixable)
}

func TestOrphanedWorktreesCheck_NoReposDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove repos directory
	reposDir := filepath.Join(ws.Path, "repos")
	require.NoError(t, os.RemoveAll(reposDir))

	check := OrphanedWorktreesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusPass, result.Status)
	assert.Contains(t, result.Message, "No repos found")
}
