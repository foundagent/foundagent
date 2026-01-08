package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWarnUncommittedChanges_LoadStateError tests error handling when state load fails
func TestWarnUncommittedChanges_LoadStateError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt state file
	stateFile := filepath.Join(ws.Path, ".foundagent", "state.json")
	require.NoError(t, os.WriteFile(stateFile, []byte("invalid"), 0644))

	err = warnUncommittedChanges(ws, "main")

	assert.Error(t, err)
}

// TestWarnUncommittedChanges_NoRepos tests with no repositories
func TestWarnUncommittedChanges_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = warnUncommittedChanges(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	// Should have no output for no repos
	assert.Empty(t, buf.String())
}

// TestWarnUncommittedChanges_NoWorktree tests when worktree doesn't exist
func TestWarnUncommittedChanges_NoWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repo but don't create worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = warnUncommittedChanges(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	// Should have no output for no worktrees
	assert.Empty(t, buf.String())
}

// TestWarnUncommittedChanges_CleanWorktree tests with clean worktree
func TestWarnUncommittedChanges_CleanWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a real git repository
	repoPath := filepath.Join(tmpDir, "test-git-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Create and commit a file
	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Add repo and link worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           repoPath,
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(filepath.Dir(wtPath), 0755))
	require.NoError(t, os.Symlink(repoPath, wtPath))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = warnUncommittedChanges(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	// Should have no warning for clean worktree
	assert.Empty(t, buf.String())
}

// TestWarnUncommittedChanges_DirtyWorktree tests with uncommitted changes
func TestWarnUncommittedChanges_DirtyWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a real git repository with uncommitted changes
	repoPath := filepath.Join(tmpDir, "test-git-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Create and commit initial file
	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Make uncommitted changes
	require.NoError(t, os.WriteFile(testFile, []byte("modified"), 0644))

	// Add repo and link worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           repoPath,
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(filepath.Dir(wtPath), 0755))
	require.NoError(t, os.Symlink(repoPath, wtPath))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = warnUncommittedChanges(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	// Should have warning for dirty worktree
	assert.Contains(t, output, "Warning")
	assert.Contains(t, output, "Uncommitted changes")
	assert.Contains(t, output, "test-repo/main")
}

// TestWarnUncommittedChanges_InvalidGitDir tests handling of invalid git directory
func TestWarnUncommittedChanges_InvalidGitDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repo with invalid worktree (just a directory, not git)
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory but not a git repo
	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = warnUncommittedChanges(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	// Should continue despite git error
	_ = buf.String()
}
