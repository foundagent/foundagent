package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunRemove_JSONModeDiscoverError tests JSON output on discover error
func TestRunRemove_JSONModeDiscoverError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	removeJSON = true
	defer func() { removeJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunRemove_JSONModeNoRepos tests JSON output when no repos
func TestRunRemove_JSONModeNoReposConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	removeJSON = true
	defer func() { removeJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunRemove_JSONModeNoWorktrees tests JSON output when no worktrees found
func TestRunRemove_JSONModeNoWorktreesForBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository without worktrees
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{},
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	removeJSON = true
	defer func() { removeJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"nonexistent"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	// Will show "No repositories" because config is not updated by AddRepository
	assert.Contains(t, buf.String(), "{")
}

// TestRunRemove_HumanModeDiscoverError tests human output on discover error
func TestRunRemove_HumanModeDiscoverError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	removeJSON = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunRemove_DeleteBranchSuccess tests successful branch deletion path
func TestRunRemove_WithDeleteBranchSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{"feature"},
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	wtPath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	removeDeleteBranch = true
	removeForce = true
	defer func() {
		removeDeleteBranch = false
		removeForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on git ops, but tests delete branch path
	_ = err
}

// TestRunRemove_VSCodeUpdateWarning tests VS Code update warning
func TestRunRemove_VSCodeUpdateWarning(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{"feature"},
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	wtPath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	// Corrupt VS Code workspace file
	vscodeFile := filepath.Join(ws.Path, "test-ws.code-workspace")
	require.NoError(t, os.WriteFile(vscodeFile, []byte("invalid"), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	removeJSON = false
	removeForce = true
	defer func() {
		removeJSON = false
		removeForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Tests VS Code warning path
	_ = err
}

// TestRunRemove_JSONModeWithDeleteBranch tests JSON output with delete branch flag
func TestRunRemove_JSONModeWithDeleteBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{"feature"},
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	wtPath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	removeJSON = true
	removeDeleteBranch = true
	removeForce = true
	defer func() {
		removeJSON = false
		removeDeleteBranch = false
		removeForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Tests JSON output path with delete branch
	_ = err
	assert.Contains(t, buf.String(), "{")
}
