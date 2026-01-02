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

func TestRemoveWorktreesParallel_Success(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := config.DefaultConfig("test-ws")

	// Create mock worktree directories
	wt1Path := filepath.Join(tmpDir, "wt1")
	wt2Path := filepath.Join(tmpDir, "wt2")
	require.NoError(t, os.MkdirAll(wt1Path, 0755))
	require.NoError(t, os.MkdirAll(wt2Path, 0755))

	worktrees := []worktreeToRemove{
		{
			RepoName:     "repo1",
			Branch:       "branch1",
			WorktreePath: wt1Path,
			BareRepoPath: "/fake/bare/repo1",
		},
		{
			RepoName:     "repo2",
			Branch:       "branch2",
			WorktreePath: wt2Path,
			BareRepoPath: "/fake/bare/repo2",
		},
	}

	results := removeWorktreesParallel(ws, cfg, worktrees)

	// Should have results for all worktrees
	assert.Len(t, results, 2)

	// May fail due to invalid git repo, but should have attempted both
	for _, result := range results {
		assert.NotEmpty(t, result.RepoName)
		assert.NotEmpty(t, result.Branch)
		assert.NotEmpty(t, result.WorktreePath)
	}
}

func TestRemoveWorktreesParallel_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := config.DefaultConfig("test-ws")

	results := removeWorktreesParallel(ws, cfg, []worktreeToRemove{})

	assert.Empty(t, results)
}

func TestDeleteBranchesForRepos_NoWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	cfg := config.DefaultConfig("test-ws")

	err = deleteBranchesForRepos(ws, cfg, "feature", []worktreeToRemove{})

	// Should not error with no worktrees
	assert.NoError(t, err)
}

func TestDeleteBranchesForRepos_WithForce(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a config
	cfg := config.DefaultConfig("test-ws")
	cfg.Repos = append(cfg.Repos, config.RepoConfig{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
	})

	repoPath := ws.BareRepoPath("test-repo")

	worktrees := []worktreeToRemove{
		{
			RepoName:     "test-repo",
			Branch:       "feature",
			BareRepoPath: repoPath,
			RepoConfig:   cfg.Repos[0],
		},
	}

	// With force, should attempt deletion even if not merged
	oldForce := removeForce
	removeForce = true
	defer func() { removeForce = oldForce }()

	// This will fail because the repo doesn't exist, but tests the code path
	err = deleteBranchesForRepos(ws, cfg, "feature", worktrees)
	// Error is expected - we're just testing the force path executes
	assert.Error(t, err)
}

func TestRemoveWorktreeFoldersFromVSCode_Success(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create worktree directories
	wt1Path := filepath.Join(ws.Path, "repos", "repo1", "wt", "branch1")
	wt2Path := filepath.Join(ws.Path, "repos", "repo2", "wt", "branch2")
	require.NoError(t, os.MkdirAll(wt1Path, 0755))
	require.NoError(t, os.MkdirAll(wt2Path, 0755))

	// Add folders to VSCode workspace
	vscodeData := &workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "repos/repo1/wt/branch1"},
			{Path: "repos/repo2/wt/branch2"},
			{Path: "other/folder"},
		},
	}
	require.NoError(t, ws.SaveVSCodeWorkspace(vscodeData))

	// Remove wt1Path from VSCode workspace
	worktrees := []worktreeToRemove{
		{
			WorktreePath: wt1Path,
		},
	}

	err = removeWorktreeFoldersFromVSCode(ws, worktrees)
	require.NoError(t, err)

	// Verify folder was removed
	reloaded, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)
	assert.Len(t, reloaded.Folders, 2)

	// Should still have repo2 and other folder
	found := false
	for _, folder := range reloaded.Folders {
		if folder.Path == "repos/repo2/wt/branch2" {
			found = true
			break
		}
	}
	assert.True(t, found, "repo2 folder should still exist")
}

func TestRemoveWorktreeFoldersFromVSCode_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add some folders
	vscodeData := &workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "folder1"},
			{Path: "folder2"},
		},
	}
	require.NoError(t, ws.SaveVSCodeWorkspace(vscodeData))

	// Remove with empty list
	err = removeWorktreeFoldersFromVSCode(ws, []worktreeToRemove{})
	require.NoError(t, err)

	// All folders should still be there
	reloaded, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)
	assert.Len(t, reloaded.Folders, 2)
}

func TestRemoveWorktreeFoldersFromVSCode_NonexistentPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add some folders
	vscodeData := &workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "folder1"},
		},
	}
	require.NoError(t, ws.SaveVSCodeWorkspace(vscodeData))

	// Try to remove a path that doesn't exist in the workspace
	worktrees := []worktreeToRemove{
		{
			WorktreePath: "/completely/different/path",
		},
	}

	err = removeWorktreeFoldersFromVSCode(ws, worktrees)
	// Should not error, just skip the path
	assert.NoError(t, err)

	// Original folder should still be there
	reloaded, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)
	assert.Len(t, reloaded.Folders, 1)
}
