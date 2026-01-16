package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a test git repository with an initial commit
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Initialize repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run(), "failed to init git repo")

	// Configure user for commits
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	// Create initial file and commit
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644))
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	return dir
}

func TestHasStagedChanges(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		{
			name:     "no staged changes",
			setup:    func(t *testing.T, dir string) {},
			expected: false,
		},
		{
			name: "has staged changes",
			setup: func(t *testing.T, dir string) {
				// Create and stage a new file
				require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644))
				cmd := exec.Command("git", "add", "new.txt")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			expected: true,
		},
		{
			name: "unstaged changes only",
			setup: func(t *testing.T, dir string) {
				// Modify existing file but don't stage
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Modified"), 0644))
			},
			expected: false,
		},
		{
			name: "staged modification",
			setup: func(t *testing.T, dir string) {
				// Modify and stage existing file
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Modified"), 0644))
				cmd := exec.Command("git", "add", "README.md")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestRepo(t)
			tt.setup(t, dir)

			result, err := HasStagedChanges(dir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommit(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		opts        CommitOptions
		expectError bool
		errorCode   string
	}{
		{
			name: "successful commit",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644))
				cmd := exec.Command("git", "add", "new.txt")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			opts:        CommitOptions{Message: "Add new file"},
			expectError: false,
		},
		{
			name: "empty message error",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644))
				cmd := exec.Command("git", "add", "new.txt")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			opts:        CommitOptions{Message: ""},
			expectError: true,
			errorCode:   "E501",
		},
		{
			name:        "nothing to commit",
			setup:       func(t *testing.T, dir string) {},
			opts:        CommitOptions{Message: "Empty commit"},
			expectError: true,
			errorCode:   "E503",
		},
		{
			name: "commit with -a flag",
			setup: func(t *testing.T, dir string) {
				// Modify tracked file without staging
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Modified"), 0644))
			},
			opts:        CommitOptions{Message: "Update readme", All: true},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestRepo(t)
			tt.setup(t, dir)

			sha, err := Commit(dir, tt.opts)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != "" {
					assert.Contains(t, err.Error(), tt.errorCode)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sha)
				assert.Len(t, sha, 7) // Short SHA
			}
		})
	}
}

func TestGetStagedFiles(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, dir string)
		expectedCount int
	}{
		{
			name:          "no staged files",
			setup:         func(t *testing.T, dir string) {},
			expectedCount: 0,
		},
		{
			name: "one staged file",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644))
				cmd := exec.Command("git", "add", "new.txt")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			expectedCount: 1,
		},
		{
			name: "multiple staged files",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644))
				cmd := exec.Command("git", "add", ".")
				cmd.Dir = dir
				require.NoError(t, cmd.Run())
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestRepo(t)
			tt.setup(t, dir)

			files, err := GetStagedFiles(dir)
			require.NoError(t, err)
			assert.Len(t, files, tt.expectedCount)
		})
	}
}

func TestGetCommitStats(t *testing.T) {
	dir := setupTestRepo(t)

	// Create and stage a file with content
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("line1\nline2\nline3\n"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	// Commit
	sha, err := Commit(dir, CommitOptions{Message: "Add file with lines"})
	require.NoError(t, err)

	// Get stats
	stats, err := GetCommitStats(dir, sha)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.FilesChanged)
	assert.Equal(t, 3, stats.Insertions)
	assert.Equal(t, 0, stats.Deletions)
}

func TestStageAllTracked(t *testing.T) {
	dir := setupTestRepo(t)

	// Modify existing file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Modified"), 0644))

	// Verify no staged changes initially
	hasStaged, err := HasStagedChanges(dir)
	require.NoError(t, err)
	assert.False(t, hasStaged)

	// Stage all tracked
	err = StageAllTracked(dir)
	require.NoError(t, err)

	// Verify changes are now staged
	hasStaged, err = HasStagedChanges(dir)
	require.NoError(t, err)
	assert.True(t, hasStaged)
}

func TestHasTrackedChanges(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		{
			name:     "no changes",
			setup:    func(t *testing.T, dir string) {},
			expected: false,
		},
		{
			name: "modified tracked file",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Modified"), 0644))
			},
			expected: true,
		},
		{
			name: "untracked file only",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("new"), 0644))
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestRepo(t)
			tt.setup(t, dir)

			result, err := HasTrackedChanges(dir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentBranch(t *testing.T) {
	dir := setupTestRepo(t)

	branch, err := GetCurrentBranch(dir)
	require.NoError(t, err)
	// Should be main or master depending on git defaults
	assert.True(t, branch == "main" || branch == "master")
}

func TestFormatDiffStat(t *testing.T) {
	tests := []struct {
		name     string
		stats    *CommitStats
		expected string
	}{
		{
			name:     "no changes",
			stats:    &CommitStats{},
			expected: "no changes",
		},
		{
			name:     "one file, insertions only",
			stats:    &CommitStats{FilesChanged: 1, Insertions: 10, Deletions: 0},
			expected: "1 file, +10",
		},
		{
			name:     "multiple files, both insertions and deletions",
			stats:    &CommitStats{FilesChanged: 3, Insertions: 45, Deletions: 12},
			expected: "3 files, +45, -12",
		},
		{
			name:     "deletions only",
			stats:    &CommitStats{FilesChanged: 1, Insertions: 0, Deletions: 5},
			expected: "1 file, -5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDiffStat(tt.stats)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommitAmend(t *testing.T) {
	dir := setupTestRepo(t)

	// Get initial commit message
	initialSHA, err := GetHeadSHA(dir)
	require.NoError(t, err)
	initialMsg, err := GetCommitMessage(dir, initialSHA)
	require.NoError(t, err)
	assert.Equal(t, "Initial commit", initialMsg)

	// Amend with new message
	_, err = Commit(dir, CommitOptions{Message: "Updated message", Amend: true})
	require.NoError(t, err)

	// Verify message changed
	newSHA, err := GetHeadSHA(dir)
	require.NoError(t, err)
	newMsg, err := GetCommitMessage(dir, newSHA)
	require.NoError(t, err)
	assert.Equal(t, "Updated message", newMsg)
}
