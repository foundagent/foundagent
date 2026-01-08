package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/require"
)

func TestRunSyncPush_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { syncJSON = false; syncVerbose = false }()

	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	syncJSON = false
	err = runSyncPush(ws)

	// No repos is not an error
	if err != nil {
		t.Errorf("Expected no error with empty workspace, got: %v", err)
	}
}

func TestRunSyncPush_NoReposJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { syncJSON = false; syncVerbose = false }()

	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	syncJSON = true
	err = runSyncPush(ws)

	// JSON mode should also succeed
	if err != nil {
		t.Errorf("Expected no error with empty workspace in JSON mode, got: %v", err)
	}
}

func TestRunSyncPush_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { syncJSON = false; syncVerbose = false }()

	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	syncVerbose = true
	err = runSyncPush(ws)

	// Verbose mode should work
	if err != nil {
		t.Errorf("Expected no error in verbose mode, got: %v", err)
	}
}
