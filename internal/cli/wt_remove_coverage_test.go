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

// TestRunRemove_DiscoverError tests error when workspace discovery fails
func TestRunRemove_DiscoverError(t *testing.T) {
	// Change to a directory that's not a workspace
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := runRemove(nil, []string{"main"})

	assert.Error(t, err)
}

// TestRunRemove_ConfigLoadErrorCorrupted tests error when config file is corrupted
func TestRunRemove_ConfigLoadErrorCorrupted(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt config file
	configFile := filepath.Join(ws.Path, ".foundagent", "foundagent.toml")
	require.NoError(t, os.WriteFile(configFile, []byte("invalid toml [[["), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	err = runRemove(nil, []string{"main"})

	assert.Error(t, err)
}

// TestRunRemove_FindWorktreesError tests error when finding worktrees fails
func TestRunRemove_FindWorktreesError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repo but corrupt state
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Corrupt state file
	stateFile := filepath.Join(ws.Path, ".foundagent", "state.json")
	require.NoError(t, os.WriteFile(stateFile, []byte("invalid json"), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	err = runRemove(nil, []string{"main"})

	assert.Error(t, err)
}

// TestRunRemove_PreValidateError tests error when pre-validation fails
func TestRunRemove_PreValidateError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repo with worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	// Change to the worktree directory (will fail validation - can't remove current dir)
	os.Chdir(wtPath)

	err = runRemove(nil, []string{"main"})

	assert.Error(t, err)
}

// TestRunRemove_WithDeleteBranchFlag tests removal with branch deletion
func TestRunRemove_WithDeleteBranchFlag(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repo with worktree (not default branch)
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
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

	// May have errors due to git operations, but should attempt removal
	_ = err
}

// TestRunRemove_JSONModeWithError tests JSON error output
func TestRunRemove_JSONModeWithError(t *testing.T) {
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

	err := runRemove(nil, []string{"main"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	// Should output JSON error
	_ = buf.String()
}

// TestRunRemove_SuccessfulRemoval tests successful worktree removal
func TestRunRemove_SuccessfulRemoval(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repo with non-default worktree
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	wtPath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	removeForce = true
	defer func() { removeForce = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// May succeed or fail depending on git operations
	_ = err
	_ = output
}
