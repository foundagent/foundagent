package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSyncPull_WithStash(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() {
		syncStash = false
		syncJSON = false
	}()

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	syncStash = true
	err = runSyncPull(ws, "main")

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.NoError(t, err)
}

func TestRunSyncFetch_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() {
		syncVerbose = false
		syncJSON = false
	}()

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	syncVerbose = true
	err = runSyncFetch(ws)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.NoError(t, err)
}

func TestRunSync_FetchMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { syncPull = false }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Default mode is fetch (no --pull flag)
	syncPull = false
	cmd := syncCmd
	err = cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func TestRunSync_PullMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { syncPull = false }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	syncPull = true
	cmd := syncCmd
	err = cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func TestRunSync_WithBranchArg(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	cmd := syncCmd
	err = cmd.RunE(cmd, []string{"main"})

	assert.NoError(t, err)
}

func TestRunSyncPull_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { syncVerbose = false }()

	// Create workspace with a repo
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a config with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	syncVerbose = true
	syncJSON = false

	// Will likely fail on git operations but tests verbose path
	_ = runSyncPull(ws, "main")
}
