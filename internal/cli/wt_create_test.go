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
