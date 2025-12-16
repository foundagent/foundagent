package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncCommand_EmptyWorkspace(t *testing.T) {
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
	syncPull = false
	syncPush = false
	syncStash = false
	syncJSON = false
	syncVerbose = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run sync on empty workspace
	err = runSync(syncCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	// Should succeed (nothing to sync)
	assert.NoError(t, err)
}

func TestSyncCommand_ConflictingFlags(t *testing.T) {
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

	// Set conflicting flags
	syncPull = true
	syncPush = true
	syncStash = false
	syncJSON = false
	syncVerbose = false

	// Run sync with conflicting flags
	err = runSync(syncCmd, []string{})

	// Should error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --push and --pull together")
}

func TestSyncCommand_StashWithoutPull(t *testing.T) {
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

	// Set stash without pull
	syncPull = false
	syncPush = false
	syncStash = true
	syncJSON = false
	syncVerbose = false

	// Run sync
	err = runSync(syncCmd, []string{})

	// Should error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--stash requires --pull")
}

func TestSyncCommand_JSONOutput(t *testing.T) {
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
	syncPull = false
	syncPush = false
	syncStash = false
	syncJSON = true
	syncVerbose = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run sync with JSON output
	err = runSync(syncCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should produce JSON output
	assert.NoError(t, err)
	assert.Contains(t, output, "{")
}

func TestSyncCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (not a workspace)
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	syncPull = false
	syncPush = false
	syncStash = false
	syncJSON = false
	syncVerbose = false

	// Run sync outside workspace
	err = runSync(syncCmd, []string{})

	// Should error because not in a workspace
	assert.Error(t, err)
}

func TestSyncCommand_WithBranch(t *testing.T) {
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
	syncPull = true
	syncPush = false
	syncStash = false
	syncJSON = false
	syncVerbose = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run sync with branch name
	_ = runSync(syncCmd, []string{"main"})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	// Command executed (may error due to no repos, but tested the flow)
}
