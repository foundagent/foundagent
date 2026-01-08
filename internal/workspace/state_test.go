package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadState(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Load state should succeed
	state, err := ws.LoadState()
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.Empty(t, state.CurrentBranch)
	assert.Empty(t, state.Repositories)
}

func TestLoadState_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	// Don't create the workspace

	// Load state should fail
	_, err = ws.LoadState()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "State file not found")
}

func TestLoadState_CorruptedJSON(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Corrupt the state file
	statePath := ws.StatePath()
	err = os.WriteFile(statePath, []byte("{ invalid json"), 0644)
	require.NoError(t, err)

	// Load state should fail with parse error
	_, err = ws.LoadState()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to parse state file")
}

func TestSaveState(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Create state with data
	state := &State{
		CurrentBranch: "feature-branch",
		Repositories: map[string]*Repository{
			"repo1": {
				Name:          "repo1",
				URL:           "https://github.com/org/repo1.git",
				DefaultBranch: "main",
				BareRepoPath:  ws.BareRepoPath("repo1"),
			},
		},
	}

	// Save state
	err = ws.SaveState(state)
	require.NoError(t, err)

	// Load it back
	loadedState, err := ws.LoadState()
	require.NoError(t, err)
	assert.Equal(t, "feature-branch", loadedState.CurrentBranch)
	assert.Len(t, loadedState.Repositories, 1)
	assert.Equal(t, "repo1", loadedState.Repositories["repo1"].Name)
}

func TestSaveState_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create workspace directory and .foundagent subdirectory
	err = os.MkdirAll(filepath.Join(ws.Path, ".foundagent"), 0755)
	require.NoError(t, err)

	// Save state should work even if state file doesn't exist yet
	state := &State{}
	err = ws.SaveState(state)
	require.NoError(t, err)

	// Verify file exists
	statePath := ws.StatePath()
	_, err = os.Stat(statePath)
	assert.NoError(t, err)
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Load config should succeed
	cfg, err := ws.LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-ws", cfg.Name)
}

func TestLoadConfig_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	// Don't create the workspace

	// Load config should fail
	_, err = ws.LoadConfig()
	assert.Error(t, err)
}

func TestSaveConfig(t *testing.T) {
	t.Skip("SaveConfig is deprecated and returns an error")
}

func TestStatePath(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	expectedPath := filepath.Join(tmpDir, "test-ws", ".foundagent", "state.json")
	assert.Equal(t, expectedPath, ws.StatePath())
}

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	expectedPath := filepath.Join(tmpDir, "test-ws", ".foundagent.yaml")
	assert.Equal(t, expectedPath, ws.ConfigPath())
}

func TestCreateState_InvalidData(t *testing.T) {
	// This test is for code completeness - the state should always marshal successfully
	// but we test the basic flow
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create the workspace directory
	err = os.MkdirAll(ws.Path, 0755)
	require.NoError(t, err)

	// Create foundagent directory
	err = os.MkdirAll(filepath.Join(ws.Path, ".foundagent"), 0755)
	require.NoError(t, err)

	// Call createState
	err = ws.createState()
	require.NoError(t, err)

	// Verify state file exists and has valid JSON
	data, err := os.ReadFile(ws.StatePath())
	require.NoError(t, err)

	var state State
	err = json.Unmarshal(data, &state)
	require.NoError(t, err)
}

func TestStateRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Create complex state
	originalState := &State{
		CurrentBranch: "feature-123",
		Repositories: map[string]*Repository{
			"repo1": {
				Name:          "repo1",
				URL:           "https://github.com/org/repo1.git",
				DefaultBranch: "main",
				BareRepoPath:  ws.BareRepoPath("repo1"),
			},
			"repo2": {
				Name:          "repo2",
				URL:           "https://github.com/org/repo2.git",
				DefaultBranch: "master",
				BareRepoPath:  ws.BareRepoPath("repo2"),
			},
		},
	}

	// Save
	err = ws.SaveState(originalState)
	require.NoError(t, err)

	// Load
	loadedState, err := ws.LoadState()
	require.NoError(t, err)

	// Verify
	assert.Equal(t, originalState.CurrentBranch, loadedState.CurrentBranch)
	assert.Len(t, loadedState.Repositories, 2)
	assert.Equal(t, "repo1", loadedState.Repositories["repo1"].Name)
	assert.Equal(t, "repo2", loadedState.Repositories["repo2"].Name)
	assert.Equal(t, "https://github.com/org/repo1.git", loadedState.Repositories["repo1"].URL)
	assert.Equal(t, "https://github.com/org/repo2.git", loadedState.Repositories["repo2"].URL)
}
