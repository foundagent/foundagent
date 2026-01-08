package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwitchCommandExists(t *testing.T) {
	// Test that switch command is registered
	cmd := worktreeCmd
	switchCmd := cmd.Commands()

	found := false
	for _, c := range switchCmd {
		if c.Name() == "switch" {
			found = true
			break
		}
	}

	assert.True(t, found, "switch command should be registered")
}

func TestSwitchCommandFlags(t *testing.T) {
	// Test that all required flags are registered
	flags := []string{"create", "from", "quiet", "json"}

	for _, flagName := range flags {
		flag := switchCmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "flag %s should exist", flagName)
	}
}

func TestWtSwitchCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to non-workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(tmpDir)

	// Reset flags
	switchCreate = false
	switchFrom = ""
	switchQuiet = false
	switchJSON = false

	// Run switch command
	err := runSwitch(switchCmd, []string{"feature-test"})

	// Should fail with workspace not found error
	assert.Error(t, err)
}

func TestWtSwitchCommand_NoArgument(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	switchCreate = false
	switchFrom = ""
	switchQuiet = false
	switchJSON = false

	// Run switch without argument should fail (worktree doesn't exist)
	err = runSwitch(switchCmd, []string{})

	// Should fail
	assert.Error(t, err)
}

func TestWtSwitchCommand_InvalidBranchName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	switchCreate = false
	switchFrom = ""
	switchQuiet = false
	switchJSON = false

	// Run with invalid branch name
	err = runSwitch(switchCmd, []string{"../invalid"})

	// Should fail with validation error
	assert.Error(t, err)
}

func TestWtSwitchCommand_FromFlagWithoutCreate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	switchCreate = false
	switchFrom = "main"
	switchQuiet = false
	switchJSON = false

	// Run switch with --from but without --create
	err = runSwitch(switchCmd, []string{"feature-test"})

	// Should fail with flag validation error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from")
}

func TestWtSwitchCommand_NoRepositories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Reset flags
	switchCreate = false
	switchFrom = ""
	switchQuiet = false
	switchJSON = false

	// Run switch command (should fail - no repos)
	err = runSwitch(switchCmd, []string{"feature-test"})

	// Should fail with "no repositories" error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repositories")
}

func TestOutputJSON(t *testing.T) {
	data := map[string]interface{}{
		"switched_to":     "feature-123",
		"previous_branch": "main",
		"workspace_file":  "/path/to/workspace.code-workspace",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(data)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Should succeed
	assert.NoError(t, err)

	// Parse JSON to verify it's valid
	var result map[string]interface{}
	jsonErr := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, jsonErr)
	assert.Equal(t, "feature-123", result["switched_to"])
	assert.Equal(t, "main", result["previous_branch"])
}

func TestWarnUncommittedChanges(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-warn", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Create empty state (no repos)
	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{},
	}
	err = ws.SaveState(state)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Should not error with empty state
	err = warnUncommittedChanges(ws, "feature-123")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Should succeed
	assert.NoError(t, err)
}

func TestRunSwitch_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := switchCmd
	err = cmd.RunE(cmd, []string{"main"})
	assert.Error(t, err)
}

func TestRunSwitch_NoArgs(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))
	require.NoError(t, ws.SaveState(&workspace.State{Repositories: map[string]*workspace.Repository{}}))

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := switchCmd
	err = cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}
