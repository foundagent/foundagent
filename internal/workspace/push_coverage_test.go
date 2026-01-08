package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushAllRepos_LoadStateError(t *testing.T) {
	// Create workspace with invalid state path
	ws := &Workspace{
		Name: "test",
		Path: "/nonexistent/path",
	}

	results, err := ws.PushAllRepos(false)

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestPushAllRepos_WithWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{"main"},
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory structure
	worktreesDir := filepath.Join(tmpDir, "test-ws", ReposDir, WorktreesDir, "test-repo")
	mainPath := filepath.Join(worktreesDir, "main")
	require.NoError(t, os.MkdirAll(mainPath, 0755))

	// Initialize git repo in worktree
	cmd := exec.Command("git", "init")
	cmd.Dir = mainPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = mainPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = mainPath
	require.NoError(t, cmd.Run())

	// Create and commit a file
	testFile := filepath.Join(mainPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = mainPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = mainPath
	require.NoError(t, cmd.Run())

	results, err := ws.PushAllRepos(false)

	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "test-repo", results[0].RepoName)
	// Status should be "nothing-to-push" since there's no remote
	assert.Contains(t, []string{"nothing-to-push", "failed"}, results[0].Status)
}

func TestPushAllRepos_NoWorktreesDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository but don't create worktrees directory
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{},
	}
	require.NoError(t, ws.AddRepository(repo))

	results, err := ws.PushAllRepos(false)

	require.NoError(t, err)
	// Should skip repos with no worktrees
	assert.NotEmpty(t, results)
}

func TestPushAllRepos_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{},
	}
	require.NoError(t, ws.AddRepository(repo))

	results, err := ws.PushAllRepos(true)

	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestPushAllRepos_GlobError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{},
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktrees directory but with invalid glob pattern scenario
	worktreesDir := filepath.Join(tmpDir, "test-ws", ReposDir, WorktreesDir, "test-repo")
	require.NoError(t, os.MkdirAll(worktreesDir, 0755))

	results, err := ws.PushAllRepos(false)

	require.NoError(t, err)
	// Should handle glob errors gracefully
	assert.NotNil(t, results)
}
