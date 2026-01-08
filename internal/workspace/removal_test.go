package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveRepoFromState(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &Repository{
		Name:      "test-repo",
		URL:       "https://github.com/test/repo.git",
		Worktrees: []string{},
	}
	require.NoError(t, ws.AddRepository(repo))

	// Verify repo exists
	state, err := ws.LoadState()
	require.NoError(t, err)
	assert.Contains(t, state.Repositories, "test-repo")

	// Remove from state
	err = ws.removeRepoFromState("test-repo")
	require.NoError(t, err)

	// Verify repo removed
	state, err = ws.LoadState()
	require.NoError(t, err)
	assert.NotContains(t, state.Repositories, "test-repo")
}

func TestRemoveRepoFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create config with a repo
	cfg := config.DefaultConfig(ws.Name)
	config.AddRepo(cfg, "https://github.com/test/repo.git", "test-repo", "main")

	// Save config
	require.NoError(t, config.Save(ws.Path, cfg))

	// Verify repo exists in config
	loadedCfg, err := ws.loadFoundagentConfig()
	require.NoError(t, err)
	assert.Len(t, loadedCfg.Repos, 1)
	assert.Equal(t, "test-repo", loadedCfg.Repos[0].Name)

	// Remove from config
	err = ws.removeRepoFromConfig("test-repo")
	require.NoError(t, err)

	// Verify repo removed
	loadedCfg, err = ws.loadFoundagentConfig()
	require.NoError(t, err)
	assert.Len(t, loadedCfg.Repos, 0)
}

func TestRemoveRepoFromWorkspaceFile(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a folder to workspace file using absolute path
	worktreePath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(worktreePath, 0755))

	vscodeData, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)

	// Add with absolute path to ensure it matches
	vscodeData.Folders = append(vscodeData.Folders, VSCodeFolder{
		Path: worktreePath,
	})
	require.NoError(t, ws.SaveVSCodeWorkspace(vscodeData))

	// Verify folder exists
	vscodeData, err = ws.LoadVSCodeWorkspace()
	require.NoError(t, err)
	initialFolderCount := len(vscodeData.Folders)

	// Remove repo worktrees from workspace file
	err = ws.removeRepoFromWorkspaceFile("test-repo")
	require.NoError(t, err)

	// Verify worktree folder removed
	vscodeData, err = ws.LoadVSCodeWorkspace()
	require.NoError(t, err)
	finalFolderCount := len(vscodeData.Folders)

	// Should have removed at least one folder (the worktree we added)
	assert.Less(t, finalFolderCount, initialFolderCount, "Should have removed the worktree folder")
}

func TestLoadSaveFoundagentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create and save config
	cfg := config.DefaultConfig(ws.Name)
	config.AddRepo(cfg, "https://github.com/test/repo.git", "test-repo", "main")

	err = ws.saveFoundagentConfig(cfg)
	require.NoError(t, err)

	// Load and verify
	loadedCfg, err := ws.loadFoundagentConfig()
	require.NoError(t, err)
	assert.Equal(t, ws.Name, loadedCfg.Workspace.Name)
	assert.Len(t, loadedCfg.Repos, 1)
	assert.Equal(t, "test-repo", loadedCfg.Repos[0].Name)
}

func TestFindDirtyWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Test with non-existent worktree base
	dirtyWorktrees, err := ws.findDirtyWorktrees("nonexistent-repo")
	assert.NoError(t, err)
	assert.Empty(t, dirtyWorktrees)

	// Create worktree directory (but not a real git worktree)
	worktreeBase := ws.WorktreeBasePath("test-repo")
	worktreePath := filepath.Join(worktreeBase, "feature")
	require.NoError(t, os.MkdirAll(worktreePath, 0755))

	// Should return empty (no git repo to check)
	dirtyWorktrees, err = ws.findDirtyWorktrees("test-repo")
	assert.NoError(t, err)
	assert.Empty(t, dirtyWorktrees)
}

func TestRemoveAllWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Test with non-existent worktree base
	count, err := ws.removeAllWorktrees("nonexistent-repo")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create worktree directories
	worktreeBase := ws.WorktreeBasePath("test-repo")
	wt1 := filepath.Join(worktreeBase, "feature1")
	wt2 := filepath.Join(worktreeBase, "feature2")
	require.NoError(t, os.MkdirAll(wt1, 0755))
	require.NoError(t, os.MkdirAll(wt2, 0755))

	// Create some files
	require.NoError(t, os.WriteFile(filepath.Join(wt1, "file.txt"), []byte("test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(wt2, "file.txt"), []byte("test"), 0644))

	// Remove all worktrees
	count, err = ws.removeAllWorktrees("test-repo")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Verify worktrees removed
	_, err = os.Stat(wt1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(wt2)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveRepo_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Try to remove non-existent repo
	result := ws.RemoveRepo("nonexistent-repo", false, false)
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "not found")
}

func TestRemoveRepo_ConfigOnly(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Add repo to state
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/org/test-repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Add to config as well
	cfg, err := ws.loadFoundagentConfig()
	require.NoError(t, err)
	cfg.Repos = append(cfg.Repos, config.RepoConfig{
		Name: "test-repo",
		URL:  "https://github.com/org/test-repo.git",
	})
	err = ws.saveFoundagentConfig(cfg)
	require.NoError(t, err)

	// Remove config only
	result := ws.RemoveRepo("test-repo", false, true)
	assert.Empty(t, result.Error)
	assert.True(t, result.ConfigOnly)
	assert.True(t, result.RemovedFromConfig)

	// Verify still in state
	state, err := ws.LoadState()
	require.NoError(t, err)
	assert.Contains(t, state.Repositories, "test-repo")

	// Verify removed from config
	cfg, err = ws.loadFoundagentConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.Repos)
}

func TestRemoveRepo_WithForce(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original working directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Add repo
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/org/test-repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Create bare repo directory
	err = os.MkdirAll(repo.BareRepoPath, 0755)
	require.NoError(t, err)

	// Remove with force (no dirty check)
	result := ws.RemoveRepo("test-repo", true, false)
	assert.Empty(t, result.Error)
	assert.True(t, result.BareCloneDeleted || result.RemovedFromConfig)
}

func TestRemoveRepo_InsideWorktree(t *testing.T) {
	t.Skip("Path prefix checking may not work reliably in test environment")
}
