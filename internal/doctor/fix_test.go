package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFixer(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	fixer := NewFixer(ws)
	assert.NotNil(t, fixer)
	assert.Equal(t, ws, fixer.Workspace)
}

func TestFixer_Fix_NotFixable(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	fixer := NewFixer(ws)

	// Test with passing result
	result := CheckResult{
		Name:    "Test Check",
		Status:  StatusPass,
		Message: "Everything OK",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	assert.Equal(t, result, fixed) // Should return unchanged

	// Test with non-fixable result
	result = CheckResult{
		Name:    "Test Check",
		Status:  StatusFail,
		Message: "Error",
		Fixable: false,
	}

	fixed = fixer.Fix(result)
	assert.Equal(t, result, fixed) // Should return unchanged
}

func TestFixer_Fix_UnknownCheck(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	fixer := NewFixer(ws)

	result := CheckResult{
		Name:    "Unknown Check Name",
		Status:  StatusFail,
		Message: "Something wrong",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	assert.Equal(t, result, fixed) // Should return unchanged for unknown check
}

func TestFixer_FixStateFile(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create config with a repo
	cfg := config.DefaultConfig(ws.Name)
	cfg.Repos = []config.RepoConfig{
		{Name: "test-repo", URL: "https://github.com/org/test-repo.git", DefaultBranch: "main"},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Create bare repo directory
	repoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Create worktree directories
	err = os.MkdirAll(ws.WorktreePath("test-repo", "main"), 0755)
	require.NoError(t, err)

	// Delete state file
	os.Remove(ws.StatePath())

	fixer := NewFixer(ws)

	result := CheckResult{
		Name:    "State file valid",
		Status:  StatusFail,
		Message: "State file is missing",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	assert.Equal(t, StatusPass, fixed.Status)
	assert.Contains(t, fixed.Message, "regenerated")

	// Verify state file was created
	assert.FileExists(t, ws.StatePath())
}

func TestFixer_FixWorkspaceStructure(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove repos directory
	reposPath := filepath.Join(ws.Path, workspace.ReposDir)
	os.RemoveAll(reposPath)

	fixer := NewFixer(ws)

	result := CheckResult{
		Name:    "Workspace structure",
		Status:  StatusFail,
		Message: "Missing repos directory",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	assert.Equal(t, StatusPass, fixed.Status)

	// Verify repos directory was created
	info, err := os.Stat(reposPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFixer_FixOrphanedRepos(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create orphaned repo directory (not in config)
	orphanedPath := filepath.Join(ws.Path, workspace.ReposDir, "orphaned-repo")
	err = os.MkdirAll(filepath.Join(orphanedPath, ".bare"), 0755)
	require.NoError(t, err)

	fixer := NewFixer(ws)

	result := CheckResult{
		Name:    "Orphaned repositories",
		Status:  StatusWarn,
		Message: "Found orphaned repos",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	assert.Equal(t, StatusPass, fixed.Status)

	// Verify orphaned repo was removed
	_, err = os.Stat(orphanedPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFixer_FixOrphanedWorktrees(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repo to state
	repo := &workspace.Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{"feature"}, // Reference to non-existent worktree
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Create bare repo
	err = os.MkdirAll(ws.BareRepoPath("test-repo"), 0755)
	require.NoError(t, err)

	// Create a worktree directory that's not in state
	orphanedWT := ws.WorktreePath("test-repo", "orphaned")
	err = os.MkdirAll(orphanedWT, 0755)
	require.NoError(t, err)

	fixer := NewFixer(ws)

	result := CheckResult{
		Name:    "Orphaned worktrees",
		Status:  StatusWarn,
		Message: "Found orphaned worktrees",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	assert.Equal(t, StatusPass, fixed.Status)
}

func TestFixer_FixConfigStateConsistency(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create config with repo
	cfg := config.DefaultConfig(ws.Name)
	cfg.Repos = []config.RepoConfig{
		{Name: "test-repo", URL: "https://github.com/org/test-repo.git", DefaultBranch: "main"},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// State is empty (inconsistent)

	fixer := NewFixer(ws)

	result := CheckResult{
		Name:    "Config/state consistency",
		Status:  StatusWarn,
		Message: "Config and state are inconsistent",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	// Should pass or warn depending on implementation
	assert.NotEqual(t, StatusFail, fixed.Status)
}

func TestFixer_FixWorkspaceFileConsistency(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	fixer := NewFixer(ws)

	result := CheckResult{
		Name:    "Workspace file consistency",
		Status:  StatusWarn,
		Message: "VS Code workspace file inconsistent",
		Fixable: true,
	}

	fixed := fixer.Fix(result)
	assert.NotEqual(t, StatusFail, fixed.Status)
}
