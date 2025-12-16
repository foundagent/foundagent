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
