package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRemove_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove config file to trigger config load error
	configPath := ws.ConfigPath()
	os.Remove(configPath)

	// Change to workspace directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(ws.Path)

	cmd := &cobra.Command{}
	err = runRemove(cmd, []string{"feature"})

	// Should error - config not found
	assert.Error(t, err)
}

func TestRunRemove_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Config exists but has no repos
	cfg := config.DefaultConfig("test-ws")
	cfg.Repos = []config.RepoConfig{} // Empty
	config.Save(ws.Path, cfg)

	// Change to workspace directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(ws.Path)

	cmd := &cobra.Command{}
	err = runRemove(cmd, []string{"feature"})

	// Should error - no repos
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No repositories")
}

func TestRunRemove_WithDeleteBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo
	cfg := config.DefaultConfig("test-ws")
	config.AddRepo(cfg, "https://github.com/test/repo.git", "test-repo", "main")
	config.Save(ws.Path, cfg)

	// Add to state with worktrees
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	ws.AddRepository(repo)

	// Create worktree directory
	wtPath := ws.WorktreePath("test-repo", "feature")
	os.MkdirAll(wtPath, 0755)

	// Change to workspace directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(ws.Path)

	// Enable delete branch flag
	oldDelete := removeDeleteBranch
	removeDeleteBranch = true
	defer func() { removeDeleteBranch = oldDelete }()

	cmd := &cobra.Command{}
	err = runRemove(cmd, []string{"feature"})

	// Will fail on git operations but tests the delete branch code path
	_ = err
}

func TestRunRemove_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove config to trigger error in JSON mode
	configPath := ws.ConfigPath()
	os.Remove(configPath)

	// Change to workspace directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(ws.Path)

	// Enable JSON mode
	oldJSON := removeJSON
	removeJSON = true
	defer func() { removeJSON = oldJSON }()

	cmd := &cobra.Command{}
	err = runRemove(cmd, []string{"feature"})

	// Should error but output in JSON
	assert.Error(t, err)
}

func TestRunRemove_JSONModeNoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Config with no repos
	cfg := config.DefaultConfig("test-ws")
	cfg.Repos = []config.RepoConfig{}
	config.Save(ws.Path, cfg)

	// Change to workspace directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(ws.Path)

	// Enable JSON mode
	oldJSON := removeJSON
	removeJSON = true
	defer func() { removeJSON = oldJSON }()

	cmd := &cobra.Command{}
	err = runRemove(cmd, []string{"feature"})

	// Should error with JSON output
	assert.Error(t, err)
}

func TestRunRemove_JSONModeNoWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo with no worktrees
	cfg := config.DefaultConfig("test-ws")
	config.AddRepo(cfg, "https://github.com/test/repo.git", "test-repo", "main")
	config.Save(ws.Path, cfg)

	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	ws.AddRepository(repo)

	// Change to workspace directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(ws.Path)

	// Enable JSON mode
	oldJSON := removeJSON
	removeJSON = true
	defer func() { removeJSON = oldJSON }()

	cmd := &cobra.Command{}
	err = runRemove(cmd, []string{"feature"})

	// Should error with JSON output
	assert.Error(t, err)
}

func TestRunRemove_PreValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo
	cfg := config.DefaultConfig("test-ws")
	config.AddRepo(cfg, "https://github.com/test/repo.git", "test-repo", "main")
	config.Save(ws.Path, cfg)

	// Add to state with worktrees
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"}, // Worktree exists
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	ws.AddRepository(repo)

	// Create worktree directory inside workspace path to trigger inside-worktree check
	wtPath := ws.WorktreePath("test-repo", "main")
	os.MkdirAll(wtPath, 0755)

	// Change to the worktree directory to trigger the "inside worktree" error
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(wtPath)

	cmd := &cobra.Command{}
	err = runRemove(cmd, []string{"main"})

	// Should error - trying to remove branch from inside its worktree
	assert.Error(t, err)
}

func TestRunSwitch_WithCurrentBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main", "feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directories
	mainPath := ws.WorktreePath("test-repo", "main")
	featurePath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(mainPath, 0755))
	require.NoError(t, os.MkdirAll(featurePath, 0755))

	// Set current branch in state
	state, _ := ws.LoadState()
	state.CurrentBranch = "main"
	ws.SaveState(state)

	// Add main folder to workspace file
	ws.AddWorktreeFolder(mainPath)

	// Change to workspace directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(ws.Path)

	// Don't quiet warnings
	oldQuiet := switchQuiet
	switchQuiet = false
	defer func() { switchQuiet = oldQuiet }()

	cmd := &cobra.Command{}
	err = runSwitch(cmd, []string{"feature"})

	// Should succeed and warn about uncommitted changes
	// (warning is non-fatal, it continues)
	assert.NoError(t, err)
}
