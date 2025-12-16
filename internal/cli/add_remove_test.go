package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCommand_NoArgs_Reconcile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory (leave config empty for clean test)
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	addJSON = false
	addForce = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run add without args on empty workspace (should report up-to-date)
	err = runAdd(addCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed with no repos
	assert.NoError(t, err)
	assert.Contains(t, output, "up-to-date")
}

func TestAddCommand_WithArgs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	addJSON = false
	addForce = false
	// addBranch removed

	// Run add with URL (will fail without real git repo, but tests the flow)
	err = runAdd(addCmd, []string{"https://github.com/org/test-repo.git"})

	// Expected to fail because we can't clone a real repo, but the function executed
	assert.Error(t, err)
}

func TestAddCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	addJSON = true
	addForce = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run add with JSON output
	_ = runAdd(addCmd, []string{"https://github.com/org/test-repo.git"})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should produce JSON output
	assert.Contains(t, output, "{")
}

func TestAddCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (not a workspace)
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	addJSON = false
	addForce = false
	// addBranch removed

	// Run add outside workspace
	err = runAdd(addCmd, []string{"https://github.com/org/test-repo.git"})

	// Should error because not in a workspace
	assert.Error(t, err)
}

func TestRemoveCommand_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &workspace.Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{},
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Also add to config
	cfg := config.DefaultConfig(ws.Name)
	config.AddRepo(cfg, "https://github.com/org/test-repo.git", "test-repo", "main")
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Verify config file exists
	configPath := filepath.Join(ws.Path, ".foundagent.yaml")
	_, err = os.Stat(configPath)
	require.NoError(t, err, "Config file should exist at %s", configPath)

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	repoRemoveJSON = false
	repoRemoveForce = false
	repoRemoveConfigOnly = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run remove
	err = runRepoRemove(removeCmd, []string{"test-repo"})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed
	assert.NoError(t, err)
	assert.Contains(t, output, "Removed")
}

func TestRemoveCommand_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	repoRemoveJSON = false
	repoRemoveForce = false
	repoRemoveConfigOnly = false

	// Capture output
	var buf bytes.Buffer
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run remove on non-existent repo
	err = runRepoRemove(removeCmd, []string{"nonexistent"})

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr
	_, _ = buf.ReadFrom(r)

	// Should error
	assert.Error(t, err)
}

func TestRemoveCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{},
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	repoRemoveJSON = true
	repoRemoveForce = false
	repoRemoveConfigOnly = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run remove with JSON output
	err = runRepoRemove(removeCmd, []string{"test-repo"})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should produce JSON output
	assert.NoError(t, err)
	assert.Contains(t, output, "{")
	assert.Contains(t, output, "repos")
}

func TestRemoveCommand_ConfigOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to both config and state
	repo := &workspace.Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{},
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	cfg := config.DefaultConfig(ws.Name)
	config.AddRepo(cfg, "https://github.com/org/test-repo.git", "test-repo", "main")
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	repoRemoveJSON = false
	repoRemoveForce = false
	repoRemoveConfigOnly = true

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run remove with config-only flag
	err = runRepoRemove(removeCmd, []string{"test-repo"})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed and mention config-only
	assert.NoError(t, err)
	assert.Contains(t, output, "config-only")
}

func TestRemoveCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (not a workspace)
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	repoRemoveJSON = false
	repoRemoveForce = false
	repoRemoveConfigOnly = false

	// Run remove outside workspace
	err = runRepoRemove(removeCmd, []string{"test-repo"})

	// Should error because not in a workspace
	assert.Error(t, err)
}
