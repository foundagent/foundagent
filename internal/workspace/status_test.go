package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkspaceStatus(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{},
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Get status
	status, err := ws.GetWorkspaceStatus(false)
	require.NoError(t, err)

	assert.Equal(t, "test-workspace", status.WorkspaceName)
	assert.NotEmpty(t, status.WorkspacePath)
	assert.Len(t, status.Repos, 1)
	assert.Equal(t, "test-repo", status.Repos[0].Name)
}

func TestGetWorkspaceStatus_WithClonedRepo(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{},
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Create bare repo structure
	bareRepoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	// Create bare repo markers
	err = os.WriteFile(filepath.Join(bareRepoPath, "HEAD"), []byte("ref: refs/heads/main"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(bareRepoPath, "config"), []byte("[core]\n\tbare = true"), 0644)
	require.NoError(t, err)

	// Get status
	status, err := ws.GetWorkspaceStatus(false)
	require.NoError(t, err)

	assert.True(t, status.Repos[0].IsCloned)
}

func TestGetWorkspaceStatus_WithWorktrees(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{},
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Create worktree directories
	mainWT := ws.WorktreePath("test-repo", "main")
	err = os.MkdirAll(mainWT, 0755)
	require.NoError(t, err)

	// Get status (verbose to check worktrees)
	status, err := ws.GetWorkspaceStatus(true)
	require.NoError(t, err)

	assert.NotNil(t, status.Summary)
}

func TestIsBareCloneExists(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)

	// Non-existent repo
	exists := ws.isBareCloneExists(filepath.Join(tmpDir, "nonexistent"))
	assert.False(t, exists)

	// Create bare repo markers
	bareRepoPath := filepath.Join(tmpDir, "bare-repo")
	err = os.MkdirAll(bareRepoPath, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(bareRepoPath, "HEAD"), []byte("ref: refs/heads/main"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(bareRepoPath, "config"), []byte("[core]"), 0644)
	require.NoError(t, err)

	// Should exist now
	exists = ws.isBareCloneExists(bareRepoPath)
	assert.True(t, exists)
}

func TestCalculateSummary(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-workspace", tmpDir)
	require.NoError(t, err)

	repoStatuses := []RepoStatus{
		{Name: "repo1", IsCloned: true},
		{Name: "repo2", IsCloned: false},
	}

	worktreeStatuses := []WorktreeStatus{
		{Repo: "repo1", Branch: "main", Status: "clean"},
		{Repo: "repo1", Branch: "feature", Status: "modified"},
	}

	summary := ws.calculateSummary(repoStatuses, worktreeStatuses)

	assert.Equal(t, 2, summary.TotalRepos)
	assert.Equal(t, 1, summary.ReposNotCloned)
	assert.Equal(t, 2, summary.TotalWorktrees)
}
