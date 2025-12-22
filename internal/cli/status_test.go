package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCommand_EmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	statusJSON = false
	statusVerbose = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run status command
	err = runStatus(statusCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed with empty workspace
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestStatusCommand_WithRepository(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository to state
	repo := &workspace.Repository{
		Name:      "test-repo",
		URL:       "https://github.com/org/test-repo.git",
		Worktrees: []string{},
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	statusJSON = false
	statusVerbose = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run status command
	err = runStatus(statusCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed and show the repository
	assert.NoError(t, err)
	assert.Contains(t, output, "test-repo")
}

func TestStatusCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	statusJSON = true
	statusVerbose = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run status command
	err = runStatus(statusCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should produce JSON output
	assert.NoError(t, err)
	assert.Contains(t, output, "{")
	assert.Contains(t, output, "}")
}

func TestStatusCommand_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	statusJSON = false
	statusVerbose = true

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run status command
	err = runStatus(statusCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	// Should succeed with verbose output
	assert.NoError(t, err)
}

func TestStatusCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (not a workspace)
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	statusJSON = false
	statusVerbose = false

	// Run status command
	err = runStatus(statusCmd, []string{})

	// Should error because not in a workspace
	assert.Error(t, err)
}

func TestOutputStatusHuman_NoRepos(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-workspace",
		WorkspacePath: "/tmp/test",
		Summary: workspace.StatusSummary{
			TotalRepos:     0,
			TotalWorktrees: 0,
			TotalBranches:  0,
		},
		Repos:     []workspace.RepoStatus{},
		Worktrees: []workspace.WorktreeStatus{},
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "No repositories configured")
	assert.Contains(t, output, "fa add")
}

func TestOutputStatusHuman_WithUncommittedChanges(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-workspace",
		WorkspacePath: "/tmp/test",
		Summary: workspace.StatusSummary{
			TotalRepos:            1,
			TotalWorktrees:        1,
			TotalBranches:         1,
			HasUncommittedChanges: true,
			DirtyWorktrees:        1,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/org/test.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{},
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Uncommitted changes")
}

func TestOutputStatusHuman_ConfigOutOfSync(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-workspace",
		WorkspacePath: "/tmp/test",
		Summary: workspace.StatusSummary{
			TotalRepos:      2,
			TotalWorktrees:  0,
			TotalBranches:   0,
			ConfigInSync:    false,
			ReposNotCloned:  1,
		},
		Repos: []workspace.RepoStatus{
			{Name: "cloned-repo", URL: "https://github.com/org/test.git", IsCloned: true},
			{Name: "not-cloned", URL: "https://github.com/org/test2.git", IsCloned: false},
		},
		Worktrees: []workspace.WorktreeStatus{},
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "not cloned")
	assert.Contains(t, output, "cloned-repo")
	assert.Contains(t, output, "not-cloned")
}

func TestOutputStatusHuman_VerboseMode(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-workspace",
		WorkspacePath: "/tmp/test",
		Summary: workspace.StatusSummary{
			TotalRepos:     1,
			TotalWorktrees: 0,
			TotalBranches:  0,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/org/test.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{},
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, true)

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "URL: https://github.com/org/test.git")
}
