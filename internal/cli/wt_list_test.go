package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestMarkCurrentWorktree(t *testing.T) {
	tests := []struct {
		name        string
		worktrees   []worktreeInfo
		currentPath string
		expected    int // index of worktree that should be current
	}{
		{
			name: "exact match",
			worktrees: []worktreeInfo{
				{Branch: "main", Repo: "repo1", Path: "/tmp/ws/repos/worktrees/repo1/main"},
				{Branch: "main", Repo: "repo2", Path: "/tmp/ws/repos/worktrees/repo2/main"},
			},
			currentPath: "/tmp/ws/repos/worktrees/repo1/main",
			expected:    0,
		},
		{
			name: "subdirectory match",
			worktrees: []worktreeInfo{
				{Branch: "main", Repo: "repo1", Path: "/tmp/ws/repos/worktrees/repo1/main"},
				{Branch: "main", Repo: "repo2", Path: "/tmp/ws/repos/worktrees/repo2/main"},
			},
			currentPath: "/tmp/ws/repos/worktrees/repo2/main/subdir",
			expected:    1,
		},
		{
			name: "no match",
			worktrees: []worktreeInfo{
				{Branch: "main", Repo: "repo1", Path: "/tmp/ws/repos/worktrees/repo1/main"},
				{Branch: "main", Repo: "repo2", Path: "/tmp/ws/repos/worktrees/repo2/main"},
			},
			currentPath: "/some/other/path",
			expected:    -1, // none should be current
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markCurrentWorktree(tt.worktrees, tt.currentPath)
			if tt.expected == -1 {
				// No worktree should be marked as current
				for _, wt := range result {
					assert.False(t, wt.IsCurrent, "no worktree should be marked current")
				}
			} else {
				// Specific worktree should be marked as current
				for i, wt := range result {
					if i == tt.expected {
						assert.True(t, wt.IsCurrent, "worktree at index %d should be current", i)
					} else {
						assert.False(t, wt.IsCurrent, "worktree at index %d should not be current", i)
					}
				}
			}
		})
	}
}

func TestBuildListOutput(t *testing.T) {
	worktrees := []worktreeInfo{
		{Branch: "main", Repo: "repo1", Path: "/tmp/ws/repos/worktrees/repo1/main", Status: "clean"},
		{Branch: "main", Repo: "repo2", Path: "/tmp/ws/repos/worktrees/repo2/main", Status: "clean"},
		{Branch: "feature", Repo: "repo1", Path: "/tmp/ws/repos/worktrees/repo1/feature", Status: "modified"},
	}

	output := buildListOutput("test-workspace", worktrees)

	assert.Equal(t, "test-workspace", output.WorkspaceName)
	assert.Equal(t, 3, output.TotalWorktrees)
	assert.Equal(t, 2, output.TotalBranches) // main and feature
	assert.Len(t, output.Worktrees, 3)
}

func TestGetWorktreesForRepo(t *testing.T) {
	// Create temp workspace
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Test with no worktrees
	worktrees, err := workspace.GetWorktreesForRepo(ws.Path, "nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, worktrees)
}

func TestDetectWorktreeStatus(t *testing.T) {
	// Test with non-existent path
	status, desc := detectWorktreeStatus("/nonexistent/path")
	assert.Equal(t, "error", status)
	assert.Equal(t, "worktree path not found", desc)
}

func TestWtListCommand_EmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	listJSONFlag = false

	// Capture output (both stdout and stderr)
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run list command
	err = runList(listCmd, []string{})

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	_, _ = stdoutBuf.ReadFrom(rOut)
	_, _ = stderrBuf.ReadFrom(rErr)
	
	output := stdoutBuf.String() + stderrBuf.String()

	// Should succeed with "No repositories" or "add repositories" message
	assert.NoError(t, err)
	hasExpectedMessage := assert.Contains(t, output, "No repositories") ||
		assert.Contains(t, output, "add repositories")
	assert.True(t, hasExpectedMessage)
}

func TestWtListCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	listJSONFlag = true

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run list command
	err = runList(listCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed with JSON output
	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	jsonErr := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, jsonErr)
	assert.Contains(t, result, "workspace_name")
	assert.Contains(t, result, "worktrees")
}

func TestWtListCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to non-workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(tmpDir)

	// Reset flags
	listJSONFlag = false

	// Run list command (should fail)
	err := runList(listCmd, []string{})

	// Should fail with workspace not found error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")
}
