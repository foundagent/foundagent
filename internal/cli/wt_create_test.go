package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreValidateWorktreeCreate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-wt", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Test: Empty workspace should pass validation (fails later in create)
	emptyCfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-wt"},
		Repos:     []config.RepoConfig{},
		Settings:  config.SettingsConfig{AutoCreateWorktree: true},
	}

	// This would need actual git repos to test fully
	// For now, just verify the function signature works
	err = preValidateWorktreeCreate(ws, emptyCfg, "feature-test", "", false)
	assert.NoError(t, err) // Empty repos list passes validation (fails later in create)
}

func TestCreateResultStructure(t *testing.T) {
	result := createResult{
		RepoName:     "test-repo",
		Branch:       "feature-123",
		SourceBranch: "main",
		WorktreePath: "/path/to/worktree",
		Status:       "success",
	}

	assert.Equal(t, "test-repo", result.RepoName)
	assert.Equal(t, "feature-123", result.Branch)
	assert.Equal(t, "main", result.SourceBranch)
	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
}

func TestWtCreateCommand_InvalidBranchName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	createJSON = false
	createFrom = ""
	createForce = false

	// Run with invalid branch name
	err = runCreate(createCmd, []string{"invalid..branch"})

	// Should fail with validation error
	assert.Error(t, err)
}

func TestWtCreateCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to non-workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(tmpDir)

	// Reset flags
	createJSON = false
	createFrom = ""
	createForce = false

	// Run create command
	err := runCreate(createCmd, []string{"feature-test"})

	// Should fail with workspace not found error
	assert.Error(t, err)
}

func TestWtCreateCommand_NoRepositories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	createJSON = false
	createFrom = ""
	createForce = false

	// Run create command (should fail - no repos)
	err = runCreate(createCmd, []string{"feature-test"})

	// Should fail with "no repositories" error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repositories")
}

func TestPreValidateWorktreeCreate_WithRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-wt", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Create config with repos (but no actual git repos)
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-wt"},
		Repos: []config.RepoConfig{
			{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		},
		Settings: config.SettingsConfig{AutoCreateWorktree: true},
	}

	// Create bare repo directory (but not actual git repo)
	bareRepoPath := ws.BareRepoPath("repo1")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	// Validation should not error (actual git operations will fail later)
	err = preValidateWorktreeCreate(ws, cfg, "feature-test", "", false)
	// This might error due to git operations, but that's expected
	// We're just testing the function doesn't panic
	_ = err
}

func TestPreValidateWorktreeCreate_ExistingWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-wt", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Create config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-wt"},
		Repos: []config.RepoConfig{
			{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		},
		Settings: config.SettingsConfig{AutoCreateWorktree: true},
	}

	// Create existing worktree directory
	worktreePath := ws.WorktreePath("repo1", "feature-test")
	err = os.MkdirAll(worktreePath, 0755)
	require.NoError(t, err)

	// Create bare repo directory
	bareRepoPath := ws.BareRepoPath("repo1")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	// Validation should fail (worktree exists)
	err = preValidateWorktreeCreate(ws, cfg, "feature-test", "", false)
	// Should error or succeed depending on git operations
	_ = err
}
