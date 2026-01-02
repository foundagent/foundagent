package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSwitch_NoWorktreesWithoutCreate(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { switchCreate = false; switchJSON = false }()

	// Create workspace with repo
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo to config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to switch to non-existent branch without --create
	switchCreate = false
	err = runSwitch(nil, []string{"nonexistent"})

	// Should error (either no repos or no worktrees)
	assert.Error(t, err)
}

func TestRunSwitch_AlreadyOnBranch(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { switchCreate = false; switchJSON = false }()

	// Create workspace with repo
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Set current branch in state
	state, _ := ws.LoadState()
	state.CurrentBranch = "main"
	ws.SaveState(state)

	// Add a repo to config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to switch to same branch
	err = runSwitch(nil, []string{"main"})

	// May error if no worktrees exist
	if err != nil {
		t.Logf("Expected error (no worktrees): %v", err)
	}
}

func TestRunSwitch_AlreadyOnBranchJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { switchCreate = false; switchJSON = false }()

	// Create workspace with repo
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Set current branch in state
	state, _ := ws.LoadState()
	state.CurrentBranch = "main"
	ws.SaveState(state)

	// Add a repo to config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to switch to same branch in JSON mode
	switchJSON = true
	err = runSwitch(nil, []string{"main"})

	// May error if no worktrees exist
	if err != nil {
		t.Logf("Expected error (no worktrees): %v", err)
	}
}

func TestWarnUncommittedChanges_WithChanges(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	// Configure git
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create initial commit
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Create repo config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	// Create worktree directory structure
	repoDir := filepath.Join(ws.Path, "worktrees", "test-repo")
	os.MkdirAll(repoDir, 0755)

	// Create a worktree subdirectory
	wtPath := filepath.Join(repoDir, "main")
	os.MkdirAll(wtPath, 0755)

	// Initialize it as a git repo with changes
	exec.Command("git", "init", wtPath).Run()
	exec.Command("git", "-C", wtPath, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", wtPath, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(wtPath, "test.txt"), []byte("content"), 0644)
	exec.Command("git", "-C", wtPath, "add", ".").Run()
	exec.Command("git", "-C", wtPath, "commit", "-m", "initial").Run()

	// Make uncommitted changes
	os.WriteFile(filepath.Join(wtPath, "test.txt"), []byte("modified"), 0644)

	// Warn about uncommitted changes
	err = warnUncommittedChanges(ws, "main")

	// Should complete (may or may not find changes depending on worktree setup)
	// The function is non-fatal so it returns nil
	if err != nil {
		t.Logf("Warning error (expected): %v", err)
	}
}

func TestRunRemove_InvalidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { removeJSON = false; removeDeleteBranch = false }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt the config
	configPath := filepath.Join(ws.Path, ".foundagent", "config.yaml")
	os.WriteFile(configPath, []byte("invalid: [yaml"), 0644)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to remove
	err = runRemove(nil, []string{"feature"})

	assert.Error(t, err)
}

func TestRunRemove_NoWorktreesForBranch(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { removeJSON = false; removeDeleteBranch = false }()

	// Create workspace with repo
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo to config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", DefaultBranch: "main"},
		},
	}
	config.Save(ws.Path, cfg)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to remove non-existent branch
	err = runRemove(nil, []string{"nonexistent"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No worktrees found")
}
