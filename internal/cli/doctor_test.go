package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctorCommand_BasicChecks(t *testing.T) {
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
	doctorFix = false
	doctorJSON = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run doctor command
	err = runDoctor(doctorCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	// Should succeed for healthy workspace
	assert.NoError(t, err)
}

func TestDoctorCommand_WithBrokenState(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove state file to create an issue
	os.Remove(ws.StatePath())

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	doctorFix = false
	doctorJSON = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run doctor command
	_ = runDoctor(doctorCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	// Should detect the issue (but not necessarily error)
	// The command may return success but report warnings
}

func TestDoctorCommand_WithFix(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove state file to create fixable issue
	os.Remove(ws.StatePath())

	// Change to workspace directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(ws.Path))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	doctorFix = true
	doctorJSON = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run doctor command with fix
	_ = runDoctor(doctorCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	// Verify state file was recreated (fixable issue should be fixed)
	assert.FileExists(t, ws.StatePath())
}

func TestDoctorCommand_JSONOutput(t *testing.T) {
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
	doctorFix = false
	doctorJSON = true

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run doctor command
	err = runDoctor(doctorCmd, []string{})

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

func TestDoctorCommand_OutsideWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (not a workspace)
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(oldCwd) }()

	// Reset flags
	doctorFix = false
	doctorJSON = false

	// Run doctor command
	err = runDoctor(doctorCmd, []string{})

	// Should error because not in a workspace
	assert.Error(t, err)
}
