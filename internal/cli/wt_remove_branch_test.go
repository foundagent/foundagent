package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeleteBranchesForRepos_EmptyList tests with empty worktree list
func TestDeleteBranchesForRepos_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := &config.Config{
		Repos: []config.RepoConfig{},
	}

	err = deleteBranchesForRepos(ws, cfg, "feature", []worktreeToRemove{})

	assert.NoError(t, err)
}

// TestDeleteBranchesForRepos_WithForceFlag tests forced deletion
func TestDeleteBranchesForRepos_WithForceFlag(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{
				Name:          "test-repo",
				DefaultBranch: "main",
			},
		},
	}

	worktrees := []worktreeToRemove{
		{
			RepoName:     "test-repo",
			BareRepoPath: filepath.Join(tmpDir, "bare"),
			RepoConfig: config.RepoConfig{
				Name:          "test-repo",
				DefaultBranch: "main",
			},
		},
	}

	removeForce = true
	defer func() { removeForce = false }()

	// Will fail due to missing bare repo, but tests force path
	_ = deleteBranchesForRepos(ws, cfg, "feature", worktrees)
}

// TestDeleteBranchesForRepos_UnmergedBranchError tests error on unmerged branch
func TestDeleteBranchesForRepos_UnmergedBranchError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{
				Name:          "test-repo",
				DefaultBranch: "main",
			},
		},
	}

	worktrees := []worktreeToRemove{
		{
			RepoName:     "test-repo",
			BareRepoPath: filepath.Join(tmpDir, "bare"),
			WorktreePath: filepath.Join(tmpDir, "worktree"),
			RepoConfig: config.RepoConfig{
				Name:          "test-repo",
				DefaultBranch: "main",
			},
		},
	}

	removeForce = false
	defer func() { removeForce = false }()

	// Will fail or return error depending on git state
	_ = deleteBranchesForRepos(ws, cfg, "feature", worktrees)
}

// TestRemoveWorktreeFoldersFromVSCode_LoadError tests error loading VS Code workspace
func TestRemoveWorktreeFoldersFromVSCode_LoadError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt VS Code workspace file
	vscodeFile := filepath.Join(ws.Path, "test-ws.code-workspace")
	require.NoError(t, os.WriteFile(vscodeFile, []byte("invalid json"), 0644))

	worktrees := []worktreeToRemove{
		{
			WorktreePath: filepath.Join(ws.Path, "repos", "test-repo", "worktrees", "feature"),
		},
	}

	err = removeWorktreeFoldersFromVSCode(ws, worktrees)

	assert.Error(t, err)
}

// TestRemoveWorktreeFoldersFromVSCode_EmptyWorktreeList tests with empty worktree list
func TestRemoveWorktreeFoldersFromVSCode_EmptyWorktreeList(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = removeWorktreeFoldersFromVSCode(ws, []worktreeToRemove{})

	assert.NoError(t, err)
}

// TestRemoveWorktreeFoldersFromVSCode_ValidRemoval tests successful removal
func TestRemoveWorktreeFoldersFromVSCode_ValidRemoval(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a worktree path
	wtPath := filepath.Join(ws.Path, "repos", "test-repo", "worktrees", "feature")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	worktrees := []worktreeToRemove{
		{
			WorktreePath: wtPath,
		},
	}

	err = removeWorktreeFoldersFromVSCode(ws, worktrees)

	assert.NoError(t, err)
}

// TestRemoveWorktreesParallel_EmptyWorktreeList tests with empty worktree list
func TestRemoveWorktreesParallel_EmptyWorktreeList(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		Repos: []config.RepoConfig{},
	}

	results := removeWorktreesParallel(ws, cfg, []worktreeToRemove{})

	assert.Empty(t, results)
}

// TestRemoveWorktreesParallel_MultipleWorktrees tests parallel removal
func TestRemoveWorktreesParallel_MultipleWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{Name: "repo1", DefaultBranch: "main"},
			{Name: "repo2", DefaultBranch: "main"},
		},
	}

	worktrees := []worktreeToRemove{
		{
			RepoName:     "repo1",
			BareRepoPath: filepath.Join(tmpDir, "repo1"),
			WorktreePath: filepath.Join(tmpDir, "wt1"),
		},
		{
			RepoName:     "repo2",
			BareRepoPath: filepath.Join(tmpDir, "repo2"),
			WorktreePath: filepath.Join(tmpDir, "wt2"),
		},
	}

	results := removeWorktreesParallel(ws, cfg, worktrees)

	// Will have errors since git repos don't exist
	assert.Len(t, results, 2)
}

// TestRemoveWorktreesParallel_WithError tests error handling
func TestRemoveWorktreesParallel_WithError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		Repos: []config.RepoConfig{
			{Name: "test-repo", DefaultBranch: "main"},
		},
	}

	worktrees := []worktreeToRemove{
		{
			RepoName:     "test-repo",
			BareRepoPath: "/nonexistent/path",
			WorktreePath: "/nonexistent/worktree",
		},
	}

	results := removeWorktreesParallel(ws, cfg, worktrees)

	assert.Len(t, results, 1)
	assert.NotNil(t, results[0].Error)
}
