package cli

import (
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
