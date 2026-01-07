package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRealGitRepoForPull creates a real git repository for testing pull operations
func setupRealGitRepoForPull(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755))

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test User"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = path
		require.NoError(t, cmd.Run())
	}

	// Create initial commit
	testFile := filepath.Join(path, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = path
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = path
	require.NoError(t, cmd.Run())
}

// setupBareRepoForPull creates a bare git repository
func setupBareRepoForPull(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755))

	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = path
	require.NoError(t, cmd.Run())
}

// TestPullAllWorktrees_RealDetachedHead tests detached HEAD with real git
func TestPullAllWorktrees_RealDetachedHead(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repoName := "detach-repo"
	bareRepoPath := ws.BareRepoPath(repoName)
	setupBareRepoForPull(t, bareRepoPath)

	// Create worktree with real git and detach HEAD
	wtPath := ws.WorktreePath(repoName, "main")
	setupRealGitRepoForPull(t, wtPath)

	// Detach HEAD
	cmd := exec.Command("git", "checkout", "--detach")
	cmd.Dir = wtPath
	require.NoError(t, cmd.Run())

	// Create state with repo
	state, _ := ws.LoadState()
	state.Repositories = map[string]*Repository{
		repoName: {
			Name:          repoName,
			DefaultBranch: "main",
			BareRepoPath:  bareRepoPath,
		},
	}
	ws.SaveState(state)

	// Pull should skip detached HEAD worktree
	results, err := ws.PullAllWorktrees("main", false, false)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "skipped", results[0].Status)
	assert.Contains(t, results[0].Error.Error(), "detached HEAD")
}

// TestPullAllWorktrees_RealUncommittedChanges tests dirty worktree without stash
func TestPullAllWorktrees_RealUncommittedChanges(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repoName := "dirty-repo"
	bareRepoPath := ws.BareRepoPath(repoName)
	setupBareRepoForPull(t, bareRepoPath)

	wtPath := ws.WorktreePath(repoName, "main")
	setupRealGitRepoForPull(t, wtPath)

	// Add uncommitted changes (modify tracked file)
	testFile := filepath.Join(wtPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("modified"), 0644))

	// Create state with repo
	state, _ := ws.LoadState()
	state.Repositories = map[string]*Repository{
		repoName: {
			Name:          repoName,
			DefaultBranch: "main",
			BareRepoPath:  bareRepoPath,
		},
	}
	ws.SaveState(state)

	// Pull without stash should skip dirty worktree
	results, err := ws.PullAllWorktrees("main", false, false)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "skipped", results[0].Status)
	assert.Contains(t, results[0].Error.Error(), "uncommitted changes")
}

// TestPullAllWorktrees_RealWithStash tests dirty worktree with stash
func TestPullAllWorktrees_RealWithStash(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repoName := "stash-repo"
	bareRepoPath := ws.BareRepoPath(repoName)
	setupBareRepoForPull(t, bareRepoPath)

	wtPath := ws.WorktreePath(repoName, "main")
	setupRealGitRepoForPull(t, wtPath)

	// Add uncommitted changes
	testFile := filepath.Join(wtPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("modified"), 0644))

	// Create state with repo
	state, _ := ws.LoadState()
	state.Repositories = map[string]*Repository{
		repoName: {
			Name:          repoName,
			DefaultBranch: "main",
			BareRepoPath:  bareRepoPath,
		},
	}
	ws.SaveState(state)

	// Pull with stash should attempt to stash and pull
	results, err := ws.PullAllWorktrees("main", true, false)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// May fail due to stash issues or no remote, but shouldn't be "skipped"
	assert.NotEqual(t, "skipped", results[0].Status)
}

// TestPullAllWorktrees_CleanWorktreePull tests clean worktree pull
func TestPullAllWorktrees_CleanWorktreePull(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repoName := "clean-repo"
	bareRepoPath := ws.BareRepoPath(repoName)
	setupBareRepoForPull(t, bareRepoPath)

	wtPath := ws.WorktreePath(repoName, "main")
	setupRealGitRepoForPull(t, wtPath)

	// Create state with repo
	state, _ := ws.LoadState()
	state.Repositories = map[string]*Repository{
		repoName: {
			Name:          repoName,
			DefaultBranch: "main",
			BareRepoPath:  bareRepoPath,
		},
	}
	ws.SaveState(state)

	// Pull clean worktree - will fail without remote but exercises the code path
	results, err := ws.PullAllWorktrees("main", false, false)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// Will fail due to no remote, but path is exercised
	assert.Contains(t, []string{"updated", "failed"}, results[0].Status)
}

// TestDetectWorktreeStatus_RealConflict tests conflict detection
func TestDetectWorktreeStatus_RealConflict(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create git repo
	repoPath := filepath.Join(tmpDir, "conflict-repo")
	setupRealGitRepoForPull(t, repoPath)

	// Create a branch for merging
	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("feature change"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "feature")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Go back to master/main
	cmd = exec.Command("git", "checkout", "-")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Make conflicting change
	require.NoError(t, os.WriteFile(testFile, []byte("main change"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "main")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Try to merge - this will create a conflict
	cmd = exec.Command("git", "merge", "feature")
	cmd.Dir = repoPath
	cmd.Run() // Ignore error - expected to conflict

	status := ws.detectWorktreeStatus(repoPath, false)
	// Should detect as conflict or modified
	assert.Contains(t, []string{"conflict", "modified"}, status.Status)
}

// TestDetectWorktreeStatus_RealModified tests modified detection with verbose
func TestDetectWorktreeStatus_RealModified(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	repoPath := filepath.Join(tmpDir, "modified-repo")
	setupRealGitRepoForPull(t, repoPath)

	// Modify tracked file
	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("modified"), 0644))

	status := ws.detectWorktreeStatus(repoPath, true)
	assert.Equal(t, "modified", status.Status)
	assert.NotEmpty(t, status.ModifiedFiles)
}

// TestDetectWorktreeStatus_RealUntracked tests untracked detection with verbose
func TestDetectWorktreeStatus_RealUntracked(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	repoPath := filepath.Join(tmpDir, "untracked-repo")
	setupRealGitRepoForPull(t, repoPath)

	// Add untracked file (only untracked, no modified files)
	untrackedFile := filepath.Join(repoPath, "untracked.txt")
	require.NoError(t, os.WriteFile(untrackedFile, []byte("new file"), 0644))

	status := ws.detectWorktreeStatus(repoPath, true)
	// Should not be clean when there are untracked files
	assert.NotEqual(t, "clean", status.Status)
}

// TestDetectWorktreeStatus_RealClean tests clean detection
func TestDetectWorktreeStatus_RealClean(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	repoPath := filepath.Join(tmpDir, "clean-repo")
	setupRealGitRepoForPull(t, repoPath)

	status := ws.detectWorktreeStatus(repoPath, false)
	assert.Equal(t, "clean", status.Status)
	assert.Empty(t, status.ModifiedFiles)
	assert.Empty(t, status.UntrackedFiles)
}
