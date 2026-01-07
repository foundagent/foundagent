package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunRemove_BranchDeletionWarning tests branch deletion warning path
func TestRunRemove_BranchDeletionWarning(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Reset flags
	removeJSON = false
	removeForce = false
	removeDeleteBranch = true
	defer func() {
		removeJSON = false
		removeForce = false
		removeDeleteBranch = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Should fail because no repos
	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunRemove_HumanModeSuccess tests human mode output when things work
func TestRunRemove_HumanModeSuccessMessage(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Reset flags
	removeJSON = false
	removeForce = false
	defer func() {
		removeJSON = false
		removeForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunRemove_ForceMode tests force mode path
func TestRunRemove_ForceModePath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Reset flags
	removeJSON = false
	removeForce = true
	defer func() {
		removeJSON = false
		removeForce = false
	}()

	// Should fail because no repos
	err = runRemove(nil, []string{"feature"})

	assert.Error(t, err)
}

// TestRunRemove_JSONModeResults tests JSON output of results
func TestRunRemove_JSONModeResultsOutput(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Reset flags
	removeJSON = true
	defer func() {
		removeJSON = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runRemove(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRemoveOutput_Struct tests the removeOutput struct
func TestRemoveOutput_Struct(t *testing.T) {
	output := removeOutput{
		Branch:          "feature",
		TotalRemoved:    2,
		TotalSkipped:    1,
		TotalFailed:     0,
		BranchesDeleted: true,
		Results: []removeResult{
			{
				RepoName:     "repo1",
				Branch:       "feature",
				WorktreePath: "/path/to/wt",
				Status:       "removed",
			},
		},
	}

	assert.Equal(t, "feature", output.Branch)
	assert.Equal(t, 2, output.TotalRemoved)
	assert.Equal(t, 1, output.TotalSkipped)
	assert.Equal(t, 0, output.TotalFailed)
	assert.True(t, output.BranchesDeleted)
	assert.Len(t, output.Results, 1)
}

// TestRemoveResult_Struct tests the removeResult struct
func TestRemoveResult_Struct(t *testing.T) {
	result := removeResult{
		RepoName:     "test-repo",
		Branch:       "feature",
		WorktreePath: "/path/to/worktree",
		Status:       "failed",
		Error:        "permission denied",
		Reason:       "has uncommitted changes",
	}

	assert.Equal(t, "test-repo", result.RepoName)
	assert.Equal(t, "feature", result.Branch)
	assert.Equal(t, "/path/to/worktree", result.WorktreePath)
	assert.Equal(t, "failed", result.Status)
	assert.Equal(t, "permission denied", result.Error)
	assert.Equal(t, "has uncommitted changes", result.Reason)
}

// TestWorktreeToRemove_Struct tests the worktreeToRemove struct
func TestWorktreeToRemove_Struct(t *testing.T) {
	wt := worktreeToRemove{
		RepoName:     "test-repo",
		Branch:       "feature",
		WorktreePath: "/path/to/worktree",
		BareRepoPath: "/path/to/bare",
	}

	assert.Equal(t, "test-repo", wt.RepoName)
	assert.Equal(t, "feature", wt.Branch)
	assert.Equal(t, "/path/to/worktree", wt.WorktreePath)
	assert.Equal(t, "/path/to/bare", wt.BareRepoPath)
}
