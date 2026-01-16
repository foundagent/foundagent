package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepoWithRemote creates a test git repo with a remote
func setupTestRepoWithRemote(t *testing.T) (localDir string) {
	t.Helper()

	// Create bare remote repo
	remoteDir := t.TempDir()
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = remoteDir
	require.NoError(t, cmd.Run(), "failed to init bare repo")

	// Create local repo
	localDir = t.TempDir()
	cmd = exec.Command("git", "init")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run(), "failed to init local repo")

	// Configure user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())

	// Add remote
	cmd = exec.Command("git", "remote", "add", "origin", remoteDir)
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())

	// Create initial commit
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "README.md"), []byte("# Test"), 0644))
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())

	// Push to remote and set upstream
	cmd = exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())

	return localDir
}

func TestHasUnpushedCommits(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		{
			name:     "no unpushed commits",
			setup:    func(t *testing.T, dir string) {},
			expected: false,
		},
		{
			name: "has unpushed commits",
			setup: func(t *testing.T, dir string) {
				// Create a new commit
				require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644))
				cmd := exec.Command("git", "add", "new.txt")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
				cmd = exec.Command("git", "commit", "-m", "New commit")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localDir := setupTestRepoWithRemote(t)
			tt.setup(t, localDir)

			result, err := HasUnpushedCommits(localDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUnpushedCount(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, dir string)
		expectedCount int
	}{
		{
			name:          "no unpushed commits",
			setup:         func(t *testing.T, dir string) {},
			expectedCount: 0,
		},
		{
			name: "one unpushed commit",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644))
				cmd := exec.Command("git", "add", "new.txt")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
				cmd = exec.Command("git", "commit", "-m", "Commit 1")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			expectedCount: 1,
		},
		{
			name: "multiple unpushed commits",
			setup: func(t *testing.T, dir string) {
				for i := 1; i <= 3; i++ {
					filename := filepath.Join(dir, "file"+string(rune('0'+i))+".txt")
					require.NoError(t, os.WriteFile(filename, []byte("content"), 0644))
					cmd := exec.Command("git", "add", ".")
					cmd.Dir = dir
					require.NoError(t, cmd.Run())
					cmd = exec.Command("git", "commit", "-m", "Commit")
					cmd.Dir = dir
					require.NoError(t, cmd.Run())
				}
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localDir := setupTestRepoWithRemote(t)
			tt.setup(t, localDir)

			count, err := GetUnpushedCount(localDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestGetUnpushedCommits(t *testing.T) {
	localDir := setupTestRepoWithRemote(t)

	// Initially no unpushed commits
	commits, err := GetUnpushedCommits(localDir)
	require.NoError(t, err)
	assert.Empty(t, commits)

	// Add some commits
	for i := 1; i <= 2; i++ {
		filename := filepath.Join(localDir, "file"+string(rune('0'+i))+".txt")
		require.NoError(t, os.WriteFile(filename, []byte("content"), 0644))
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = localDir
		require.NoError(t, cmd.Run())
		cmd = exec.Command("git", "commit", "-m", "Commit "+string(rune('0'+i)))
		cmd.Dir = localDir
		require.NoError(t, cmd.Run())
	}

	// Get unpushed commits
	commits, err = GetUnpushedCommits(localDir)
	require.NoError(t, err)
	assert.Len(t, commits, 2)
	// Most recent first
	assert.Contains(t, commits[0], "Commit 2")
	assert.Contains(t, commits[1], "Commit 1")
}

func TestPushWithOptions(t *testing.T) {
	localDir := setupTestRepoWithRemote(t)

	// Create a commit to push
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "New commit")
	cmd.Dir = localDir
	require.NoError(t, cmd.Run())

	// Verify we have unpushed commits
	hasUnpushed, err := HasUnpushedCommits(localDir)
	require.NoError(t, err)
	assert.True(t, hasUnpushed)

	// Push
	err = PushWithOptions(localDir, false)
	require.NoError(t, err)

	// Verify no more unpushed commits
	hasUnpushed, err = HasUnpushedCommits(localDir)
	require.NoError(t, err)
	assert.False(t, hasUnpushed)
}

func TestGetPushRefspec(t *testing.T) {
	localDir := setupTestRepoWithRemote(t)

	refspec, err := GetPushRefspec(localDir)
	require.NoError(t, err)

	// Should be something like "main -> origin/main" or "master -> origin/master"
	assert.Contains(t, refspec, "-> origin/")
}

func TestHasUpstreamConfigured(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
	}{
		{
			name: "has upstream",
			setup: func(t *testing.T) string {
				return setupTestRepoWithRemote(t)
			},
			expected: true,
		},
		{
			name: "no upstream",
			setup: func(t *testing.T) string {
				dir := setupTestRepo(t)
				return dir
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			result, err := HasUpstreamConfigured(dir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasUnpushedCommitsNoRemote(t *testing.T) {
	// Test with a repo that has no remote
	dir := setupTestRepo(t)

	// Should return false, not error
	result, err := HasUnpushedCommits(dir)
	require.NoError(t, err)
	assert.False(t, result)
}

func TestGetUnpushedCountNoRemote(t *testing.T) {
	// Test with a repo that has no remote
	dir := setupTestRepo(t)

	// Should return 0, not error
	count, err := GetUnpushedCount(dir)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
