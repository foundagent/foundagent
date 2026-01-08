package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckForUpdates_Success(t *testing.T) {
	// Just test that checkForUpdates runs without crashing
	// It will check GitHub API which may succeed or fail depending on network
	err := checkForUpdates()
	// Should always return nil (errors are logged, not returned)
	assert.NoError(t, err)
}

func TestRunSyncFetch_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	syncJSON = false
	syncVerbose = false
	defer func() {
		syncJSON = false
		syncVerbose = false
	}()

	err = runSyncFetch(ws)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "No repositories")
}

func TestRunSyncFetch_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	syncJSON = true
	defer func() { syncJSON = false }()

	err = runSyncFetch(ws)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.NoError(t, err)
}

func TestRunSyncPull_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	syncJSON = false
	syncVerbose = false
	syncStash = false
	defer func() {
		syncJSON = false
		syncVerbose = false
		syncStash = false
	}()

	err = runSyncPull(ws, "main")

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "No repositories")
}

func TestRunSyncPull_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	syncJSON = true
	defer func() { syncJSON = false }()

	err = runSyncPull(ws, "main")

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.NoError(t, err)
}

func TestOutputSyncJSON_EmptyResults(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err := outputSyncJSON([]workspace.SyncResult{})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "{")
}
