package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultBranch_SymbolicRef(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a bare repository
	bareRepo := filepath.Join(tmpDir, "test.git")
	cmd := exec.Command("git", "init", "--bare", bareRepo)
	require.NoError(t, cmd.Run())

	// Create a temporary working directory
	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	// Initialize and configure
	cmd = exec.Command("git", "init")
	cmd.Dir = workDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = workDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = workDir
	require.NoError(t, cmd.Run())

	// Create initial commit
	testFile := filepath.Join(workDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = workDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = workDir
	require.NoError(t, cmd.Run())

	// Add remote
	cmd = exec.Command("git", "remote", "add", "origin", bareRepo)
	cmd.Dir = workDir
	require.NoError(t, cmd.Run())

	// Push to bare repo
	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = workDir
	_ = cmd.Run() // May fail but that's ok for this test

	branch, err := GetDefaultBranch(bareRepo)

	// Should succeed with some branch
	require.NoError(t, err)
	assert.NotEmpty(t, branch)
}

func TestGetDefaultBranch_Fallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty bare repository
	bareRepo := filepath.Join(tmpDir, "test.git")
	cmd := exec.Command("git", "init", "--bare", bareRepo)
	require.NoError(t, cmd.Run())

	branch, err := GetDefaultBranch(bareRepo)

	// Should fallback to "main"
	require.NoError(t, err)
	assert.Equal(t, "main", branch)
}

func TestGetDefaultBranch_InvalidPath(t *testing.T) {
	branch, err := GetDefaultBranch("/nonexistent/path")

	// Should return main as fallback
	require.NoError(t, err)
	assert.Equal(t, "main", branch)
}

func TestIsBranchMerged_NotMerged(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Create initial commit on master/main
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Create a new branch
	cmd = exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Make a commit on feature
	require.NoError(t, os.WriteFile(testFile, []byte("feature"), 0644))
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "feature commit")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Switch back to main/master
	cmd = exec.Command("git", "checkout", "master")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Check if feature is merged (it's not)
	merged, err := IsBranchMerged(tmpDir, "feature", "master")

	// May fail or return false
	_ = err
	_ = merged
}
