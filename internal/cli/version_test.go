package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/foundagent/foundagent/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommandExists(t *testing.T) {
	// Verify version command is registered
	assert.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)
}

func TestVersionCommandFlags(t *testing.T) {
	// Verify all flags exist
	fullFlag := versionCmd.Flags().Lookup("full")
	assert.NotNil(t, fullFlag)
	assert.Equal(t, "bool", fullFlag.Value.Type())

	jsonFlag := versionCmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "bool", jsonFlag.Value.Type())

	checkFlag := versionCmd.Flags().Lookup("check")
	assert.NotNil(t, checkFlag)
	assert.Equal(t, "bool", checkFlag.Value.Type())
}

func TestRootVersionFlag(t *testing.T) {
	// Verify --version flag exists on root command
	versionFlag := rootCmd.PersistentFlags().Lookup("version")
	assert.NotNil(t, versionFlag)
	assert.Equal(t, "bool", versionFlag.Value.Type())
}

func TestVersionCommand_Default(t *testing.T) {
	// Reset flags
	versionFull = false
	versionJSON = false
	versionCheck = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run version command
	err := runVersion(versionCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed and show version
	assert.NoError(t, err)
	assert.Contains(t, output, version.String())
}

func TestVersionCommand_Full(t *testing.T) {
	// Reset flags
	versionFull = true
	versionJSON = false
	versionCheck = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run version command
	err := runVersion(versionCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed and show full version
	assert.NoError(t, err)
	assert.Equal(t, version.Full()+"\n", output)
}

func TestVersionCommand_JSON(t *testing.T) {
	// Reset flags
	versionFull = false
	versionJSON = true
	versionCheck = false

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run version command
	err := runVersion(versionCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed with JSON output
	assert.NoError(t, err)

	// Parse JSON to verify structure
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Verify expected fields
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "commit")
}

func TestVersionCommand_Check(t *testing.T) {
	// Reset flags
	versionFull = false
	versionJSON = false
	versionCheck = true

	// Capture both stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run version command (might fail to reach GitHub, but should not error)
	err := runVersion(versionCmd, []string{})

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	_, _ = stdoutBuf.ReadFrom(rOut)
	_, _ = stderrBuf.ReadFrom(rErr)

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	// Should succeed (update check failures are warnings only)
	assert.NoError(t, err)

	// Output should contain version or update info or warning
	hasVersion := strings.Contains(stdout, version.String())
	hasUpdateCheck := strings.Contains(stdout, "up to date") ||
		strings.Contains(stdout, "Update available") ||
		strings.Contains(stderr, "Failed to check for updates")

	assert.True(t, hasVersion || hasUpdateCheck,
		"Expected version info or update check result, got stdout: %q, stderr: %q", stdout, stderr)
}

func TestOutputVersionJSON(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run outputVersionJSON
	err := outputVersionJSON()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed
	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "commit")
	assert.Contains(t, result, "build_date")
}
