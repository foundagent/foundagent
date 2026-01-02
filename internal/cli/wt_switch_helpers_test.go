package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAvailableBranches_NoBranches(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// No worktrees exist
	err = listAvailableBranches(ws)
	// Should not error
	assert.NoError(t, err)
}

func TestListAvailableBranches_WithBranches(t *testing.T) {
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

	err = listAvailableBranches(ws)
	// Should not error
	assert.NoError(t, err)
}

func TestListAvailableBranches_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with branches
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	mainPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(mainPath, 0755))

	// Test in JSON mode
	oldJSON := switchJSON
	switchJSON = true
	defer func() { switchJSON = oldJSON }()

	err = listAvailableBranches(ws)
	// Should not error
	assert.NoError(t, err)
}

func TestCreateWorktreesForBranch_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = createWorktreesForBranch(ws, "feature")
	// Should error - no repositories
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No repositories")
}

func TestCreateWorktreesForBranch_InvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository but don't create the actual git repo
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	err = createWorktreesForBranch(ws, "feature")
	// Should error - can't create worktree without real git repo
	assert.Error(t, err)
}

func TestCreateWorktreesForBranch_WithFromBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Set source branch
	oldFrom := switchFrom
	switchFrom = "develop"
	defer func() { switchFrom = oldFrom }()

	err = createWorktreesForBranch(ws, "feature")
	// Should error - can't create worktree without real git repo
	assert.Error(t, err)
}

func TestCreateMissingWorktrees_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = createMissingWorktrees(ws, "feature", []string{})
	// Should not error with empty list
	assert.NoError(t, err)
}

func TestCreateMissingWorktrees_InvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	err = createMissingWorktrees(ws, "feature", []string{"test-repo"})
	// Should error - can't create worktree without real git repo
	assert.Error(t, err)
}

func TestCreateMissingWorktrees_WithFromBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Set source branch
	oldFrom := switchFrom
	switchFrom = "develop"
	defer func() { switchFrom = oldFrom }()

	err = createMissingWorktrees(ws, "feature", []string{"test-repo"})
	// Should error - can't create worktree without real git repo
	assert.Error(t, err)
}

func TestCreateMissingWorktrees_EmptyDefaultBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository without a default branch
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	err = createMissingWorktrees(ws, "feature", []string{"test-repo"})
	// Should error - can't create worktree without real git repo
	// But the function should default to "main" for sourceBranch
	assert.Error(t, err)
}

func TestWarnUncommittedChanges_NoWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	err = warnUncommittedChanges(ws, "main")
	// Should not error - no worktrees to check
	assert.NoError(t, err)
}

func TestWarnUncommittedChanges_WithNonGitDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create a worktree directory but not a git repo
	worktreePath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(worktreePath, 0755))

	err = warnUncommittedChanges(ws, "main")
	// Should not error - git check will fail but it's handled
	assert.NoError(t, err)
}

func TestOutputJSON_ValidData(t *testing.T) {
	data := map[string]interface{}{
		"test": "value",
		"num":  42,
	}

	err := outputJSON(data)
	assert.NoError(t, err)
}

func TestOutputJSON_EmptyData(t *testing.T) {
	err := outputJSON(map[string]interface{}{})
	assert.NoError(t, err)
}
