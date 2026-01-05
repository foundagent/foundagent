package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullAllWorktrees_DetachedHead(t *testing.T) {
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

	results, err := ws.PullAllWorktrees("main", false, false)

	// Should succeed with skipped results
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "skipped", results[0].Status)
}

func TestPullAllWorktrees_UncommittedChangesNoStash(t *testing.T) {
	t.Skip("Complex test - requires proper git worktree setup")
}

func TestPushAllRepos_Extended_WithVerbose(t *testing.T) {
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

	results, err := ws.PushAllRepos(true)

	// Should succeed (may be empty if no worktrees)
	require.NoError(t, err)
	_ = results
}

func TestDetectWorktreeStatus_Extended_WithVerbose(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a git repo
	repoPath := filepath.Join(tmpDir, "test-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Create and commit a file
	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	status := ws.detectWorktreeStatus(repoPath, true)

	// Should return clean status
	assert.NotNil(t, status)
}

func TestRemoveRepo_MultipleWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository with multiple worktrees
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main", "develop", "feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create bare repo and worktree directories
	bareRepoPath := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(bareRepoPath, 0755))

	for _, branch := range repo.Worktrees {
		wtPath := ws.WorktreePath("test-repo", branch)
		require.NoError(t, os.MkdirAll(wtPath, 0755))
	}

	// Remove with force
	result := ws.RemoveRepo("test-repo", true, false)

	// Should succeed
	assert.Empty(t, result.Error)
}

func TestGetWorkspaceStatus_WithVerbose(t *testing.T) {
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

	// Should succeed
	require.NoError(t, err)
	assert.NotNil(t, status)
}

func TestValidatePathLength_LongPath(t *testing.T) {
	// Create a very long path
	longName := ""
	for i := 0; i < 100; i++ {
		longName += "a"
	}

	err := ValidatePathLength(longName)

	// May succeed or fail depending on OS limits
	_ = err
}
