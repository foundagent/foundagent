package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoriesCheck_AllValid(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create valid bare repo structure
	bareRepo := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(filepath.Join(bareRepo, "objects"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(bareRepo, "refs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(bareRepo, "HEAD"), []byte("ref: refs/heads/main"), 0644))

	check := RepositoriesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusPass, result.Status)
	assert.Contains(t, result.Message, "All 1 repositories valid")
}

func TestRepositoriesCheck_MissingBareClone(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state but don't create directory
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	check := RepositoriesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Message, "issue(s)")
	assert.False(t, result.Fixable)
}

func TestRepositoriesCheck_InvalidGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create directory but not a valid git repo (no objects folder)
	bareRepo := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(bareRepo, 0755))

	check := RepositoriesCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Message, "issue(s)")
}

func TestOrphanedReposCheck_NoOrphans(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create matching directory
	bareRepo := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(bareRepo, 0755))

	check := OrphanedReposCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusPass, result.Status)
}

func TestOrphanedReposCheck_WithOrphans(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create orphaned repo directory with .bare subdirectory (not in state)
	orphanedPath := filepath.Join(ws.Path, "repos", "orphaned-repo", ".bare")
	require.NoError(t, os.MkdirAll(orphanedPath, 0755))

	check := OrphanedReposCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusWarn, result.Status)
	assert.Contains(t, result.Message, "orphaned")
	assert.True(t, result.Fixable)
}

func TestOrphanedReposCheck_NoReposDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove repos directory
	reposDir := filepath.Join(ws.Path, "repos")
	require.NoError(t, os.RemoveAll(reposDir))

	check := OrphanedReposCheck{Workspace: ws}
	result := check.Run()

	assert.Equal(t, StatusPass, result.Status)
	assert.Contains(t, result.Message, "No repositories found")
}
