package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCreate_ConfigLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace but corrupt the config
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Write invalid config
	configPath := ws.Path + "/.foundagent/foundagent.yaml"
	err = os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	cmd := createCmd
	err = cmd.RunE(cmd, []string{"feature-123"})

	assert.Error(t, err)
}

func TestRunCreate_ConfigLoadErrorJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { createJSON = false }()

	// Create workspace but corrupt the config
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Write invalid config
	configPath := ws.Path + "/.foundagent/foundagent.yaml"
	err = os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	createJSON = true
	cmd := createCmd
	err = cmd.RunE(cmd, []string{"feature-123"})

	assert.Error(t, err)
}

func TestRunCreate_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with a repo config
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a config with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Create bare repo directory
	bareRepoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Will fail on validation (no real git repo)
	cmd := createCmd
	_ = cmd.RunE(cmd, []string{"feature-123"})
}

func TestRunCreate_ValidationErrorJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { createJSON = false }()

	// Create workspace with a repo config
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a config with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Create bare repo directory
	bareRepoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	createJSON = true
	cmd := createCmd
	_ = cmd.RunE(cmd, []string{"feature-123"})
}

func TestCreateWorktreeForRepo_NonExistentRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Test with non-existent repo
	repo := config.RepoConfig{
		Name:          "nonexistent",
		URL:           "https://github.com/org/repo.git",
		DefaultBranch: "main",
	}

	result := createWorktreeForRepo(ws, repo, "feature", "", false)

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

func TestCreateWorktreeForRepo_WithSourceBranch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create bare repo directory (but not a real git repo)
	repo := config.RepoConfig{
		Name:          "test-repo",
		URL:           "https://github.com/org/repo.git",
		DefaultBranch: "main",
	}
	bareRepoPath := ws.BareRepoPath(repo.Name)
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	// Will fail on git operations but tests the code path
	result := createWorktreeForRepo(ws, repo, "feature", "main", false)

	assert.Equal(t, "error", result.Status)
}

func TestCreateWorktreeForRepo_DefaultBranchDetection(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Test with no default branch in config (will try to detect)
	result := createWorktreeForRepo(ws, config.RepoConfig{
		Name: "test-repo",
		// No DefaultBranch set
	}, "feature", "", false)

	// Should error - either from branch detection or worktree creation
	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

func TestCreateWorktreeForRepo_ForceRemoveWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create bare repo directory
	bareRepoPath := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(bareRepoPath, 0755))

	// Create existing worktree path
	wtPath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	result := createWorktreeForRepo(ws, config.RepoConfig{
		Name:          "test-repo",
		DefaultBranch: "main",
	}, "feature", "", true)

	// Will error during worktree operations but tests force path
	assert.Equal(t, "error", result.Status)
}

func TestCreateWorktreeForRepo_SourceBranchExplicit(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Test with explicit source branch (overrides default)
	result := createWorktreeForRepo(ws, config.RepoConfig{
		Name:          "test-repo",
		DefaultBranch: "main",
	}, "feature", "develop", false)

	assert.Equal(t, "error", result.Status)
	assert.Equal(t, "test-repo", result.RepoName)
	assert.Equal(t, "feature", result.Branch)
}
