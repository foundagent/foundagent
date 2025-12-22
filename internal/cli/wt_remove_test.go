package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindWorktreesForBranch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-remove", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Create mock config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-remove"},
		Repos: []config.RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
			{URL: "https://github.com/test/repo2.git", Name: "repo2", DefaultBranch: "main"},
		},
	}

	// Create mock worktree directories
	wt1Path := ws.WorktreePath("repo1", "feature-test")
	wt2Path := ws.WorktreePath("repo2", "feature-test")
	require.NoError(t, os.MkdirAll(wt1Path, 0755))
	require.NoError(t, os.MkdirAll(wt2Path, 0755))

	// Find worktrees
	worktrees, err := findWorktreesForBranch(ws, cfg, "feature-test")
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)

	// Test with non-existent branch
	worktrees, err = findWorktreesForBranch(ws, cfg, "nonexistent")
	require.NoError(t, err)
	assert.Len(t, worktrees, 0)
}

func TestPreValidateRemoval_DirtyWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-dirty", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-dirty"},
		Repos: []config.RepoConfig{
			{URL: "https://github.com/test/repo1.git", Name: "repo1", DefaultBranch: "main"},
		},
	}

	// Create worktree
	wtPath := ws.WorktreePath("repo1", "feature-test")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	worktrees := []worktreeToRemove{
		{
			RepoName:     "repo1",
			RepoConfig:   cfg.Repos[0],
			Branch:       "feature-test",
			WorktreePath: wtPath,
			BareRepoPath: ws.BareRepoPath("repo1"),
		},
	}

	// Validation with non-existent worktree should pass (no git repo to check)
	err = preValidateRemoval(ws, cfg, worktrees, "feature-test")
	// Since there's no actual git repo, it won't fail on dirty check
	assert.NoError(t, err)
}

func TestRemoveOutput(t *testing.T) {
	output := removeOutput{
		Branch:          "feature-123",
		TotalRemoved:    2,
		TotalSkipped:    0,
		TotalFailed:     0,
		BranchesDeleted: true,
		Results: []removeResult{
			{RepoName: "repo1", Branch: "feature-123", Status: "removed"},
			{RepoName: "repo2", Branch: "feature-123", Status: "removed"},
		},
	}

	assert.Equal(t, 2, output.TotalRemoved)
	assert.Equal(t, 0, output.TotalFailed)
	assert.True(t, output.BranchesDeleted)
}

func TestWtRemoveCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to non-workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(tmpDir)

	// Reset flags
	removeJSON = false
	removeForce = false
	removeDeleteBranch = false

	// Run remove command
	err := runRemove(removeCmd, []string{"feature-test"})

	// Should fail with workspace not found error
	assert.Error(t, err)
}

func TestWtRemoveCommand_NoRepositories(t *testing.T) {
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
	removeJSON = false
	removeForce = false
	removeDeleteBranch = false

	// Run remove command (should fail - no repos)
	err = runRemove(removeCmd, []string{"feature-test"})

	// Should fail with "no repositories" error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repositories")
}

func TestPrintRemoveJSON(t *testing.T) {
	output := removeOutput{
		Branch:          "feature-123",
		TotalRemoved:    2,
		TotalSkipped:    1,
		TotalFailed:     0,
		BranchesDeleted: true,
		Results: []removeResult{
			{RepoName: "repo1", Branch: "feature-123", Status: "removed"},
			{RepoName: "repo2", Branch: "feature-123", Status: "removed"},
			{RepoName: "repo3", Branch: "feature-123", Status: "skipped", Reason: "not found"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printRemoveJSON(output)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Should succeed
	assert.NoError(t, err)

	// Parse JSON to verify it's valid
	var result map[string]interface{}
	jsonErr := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, jsonErr)
	assert.Equal(t, "feature-123", result["branch"])
	assert.Equal(t, float64(2), result["total_removed"])
}
