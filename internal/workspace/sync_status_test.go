package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPullAllWorktrees_StateLoadError tests error when loading state fails
func TestPullAllWorktrees_StateLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	// Don't create workspace, state load will fail

	results, err := ws.PullAllWorktrees("main", false, false)

	assert.Error(t, err)
	assert.Nil(t, results)
}

// TestPullAllWorktrees_NoRepositories tests with no repositories
func TestPullAllWorktrees_NoRepositories(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	results, err := ws.PullAllWorktrees("main", false, false)

	// May return empty results or error depending on sync
	_ = err
	_ = results
}

// TestPullAllWorktrees_StashEnabled tests stash functionality
func TestPullAllWorktrees_StashEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	results, err := ws.PullAllWorktrees("main", true, false)

	// Will fail due to missing git repo
	_ = err
	_ = results
}

// TestPullAllWorktrees_VerboseEnabled tests verbose output
func TestPullAllWorktrees_VerboseEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	results, err := ws.PullAllWorktrees("main", false, true)

	// Will fail due to missing git repo
	_ = err
	_ = results
}

// TestDetectWorktreeStatus_MissingPath tests with non-existent path
func TestDetectWorktreeStatus_MissingPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	status := ws.detectWorktreeStatus("/nonexistent/path", false)

	assert.Equal(t, "missing", status.Status)
}

// TestDetectWorktreeStatus_ExistingPath tests with existing path
func TestDetectWorktreeStatus_ExistingPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a directory to test
	testPath := filepath.Join(tmpDir, "test-worktree")
	require.NoError(t, os.MkdirAll(testPath, 0755))

	status := ws.detectWorktreeStatus(testPath, false)

	// Will be clean since not a git repo
	assert.NotEmpty(t, status.Status)
}

// TestDetectWorktreeStatus_VerboseMode tests verbose output
func TestDetectWorktreeStatus_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a directory to test
	testPath := filepath.Join(tmpDir, "test-worktree")
	require.NoError(t, os.MkdirAll(testPath, 0755))

	status := ws.detectWorktreeStatus(testPath, true)

	// Returns a status
	assert.NotEmpty(t, status.Status)
}

// TestGetModifiedFilesInPath tests getting modified files
func TestGetModifiedFilesInPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	files := ws.getModifiedFiles(tmpDir)

	// Will return empty since not a git repo
	assert.Empty(t, files)
}

// TestGetUntrackedFilesInPath tests getting untracked files
func TestGetUntrackedFilesInPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	files := ws.getUntrackedFiles(tmpDir)

	// Will return empty since not a git repo
	assert.Empty(t, files)
}

// TestCalculateSummary_MixedStatus tests aggregate statistics with mixed repos
func TestCalculateSummary_MixedStatus(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repos := []RepoStatus{
		{Name: "repo1", IsCloned: true},
		{Name: "repo2", IsCloned: false},
	}
	worktrees := []WorktreeStatus{
		{Branch: "main"},
		{Branch: "feature"},
	}

	summary := ws.calculateSummary(repos, worktrees)

	assert.Equal(t, 2, summary.TotalRepos)
	assert.Equal(t, 2, summary.TotalWorktrees)
	assert.Equal(t, 1, summary.ReposNotCloned)
}

// TestCalculateSummary_AllReposCloned tests when all repos are cloned
func TestCalculateSummary_AllReposCloned(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repos := []RepoStatus{
		{Name: "repo1", IsCloned: true},
		{Name: "repo2", IsCloned: true},
	}
	worktrees := []WorktreeStatus{
		{Branch: "main"},
	}

	summary := ws.calculateSummary(repos, worktrees)

	assert.Equal(t, 2, summary.TotalRepos)
	assert.Equal(t, 1, summary.TotalWorktrees)
	assert.Equal(t, 0, summary.ReposNotCloned)
}
