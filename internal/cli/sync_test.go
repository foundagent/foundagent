package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestSyncCommandExists(t *testing.T) {
	// Verify sync command is registered
	assert.NotNil(t, syncCmd)
	assert.Equal(t, "sync [branch]", syncCmd.Use)
}

func TestSyncCommandFlags(t *testing.T) {
	// Verify all flags exist
	pullFlag := syncCmd.Flags().Lookup("pull")
	assert.NotNil(t, pullFlag)
	assert.Equal(t, "bool", pullFlag.Value.Type())

	pushFlag := syncCmd.Flags().Lookup("push")
	assert.NotNil(t, pushFlag)
	assert.Equal(t, "bool", pushFlag.Value.Type())

	stashFlag := syncCmd.Flags().Lookup("stash")
	assert.NotNil(t, stashFlag)
	assert.Equal(t, "bool", stashFlag.Value.Type())

	jsonFlag := syncCmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "bool", jsonFlag.Value.Type())

	verboseFlag := syncCmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "bool", verboseFlag.Value.Type())
}

func TestRunSync_ConflictingFlags(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original directory
	origDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(origDir)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(ws.Path)
	assert.NoError(t, err)

	// Test validation - cannot use --push and --pull together
	syncPush = true
	syncPull = true
	defer func() {
		syncPush = false
		syncPull = false
	}()

	err = runSync(syncCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --push and --pull together")
}

func TestRunSync_StashWithoutPull(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original directory
	origDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(origDir)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(ws.Path)
	assert.NoError(t, err)

	// Test validation - --stash requires --pull
	syncStash = true
	syncPull = false
	defer func() {
		syncStash = false
	}()

	err = runSync(syncCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--stash requires --pull")
}

func TestRunSync_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	err := os.Chdir(tmpDir)
	assert.NoError(t, err)
	
	err = runSync(syncCmd, []string{})
	assert.Error(t, err)
}

func TestRunSyncFetch_NoWorkspace(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(origDir)

	// Change to temp dir with no workspace
	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	assert.NoError(t, err)

	// Reset all flags
	syncPush = false
	syncPull = false
	syncStash = false
	syncJSON = false
	syncVerbose = false

	// Should error with no workspace
	err = runSync(syncCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")
}
