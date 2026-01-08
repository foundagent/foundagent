package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRunInit_JSONModeSuccess tests JSON output on success
func TestRunInit_JSONModeSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	initJSON = true
	initForce = false
	defer func() {
		initJSON = false
		initForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(nil, []string{"test-ws"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test-ws")
	assert.Contains(t, buf.String(), "{")
}

// TestRunInit_JSONModeError tests JSON output on error
func TestRunInit_JSONModeError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	initJSON = true
	initForce = false
	defer func() {
		initJSON = false
		initForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Invalid workspace name
	err := runInit(nil, []string{""})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunInit_ForceFlag tests force flag for reinitialization
func TestRunInit_ForceFlag(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	initJSON = false
	initForce = false
	defer func() {
		initJSON = false
		initForce = false
	}()

	// First create
	err := runInit(nil, []string{"test-ws"})
	assert.NoError(t, err)

	// Try to create again without force - should fail
	err = runInit(nil, []string{"test-ws"})
	assert.Error(t, err)

	// Now with force - should succeed
	initForce = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runInit(nil, []string{"test-ws"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "reinitialized")
}

// TestRunInit_JSONModeForce tests JSON output with force flag
func TestRunInit_JSONModeForce(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	initJSON = true
	initForce = false
	defer func() {
		initJSON = false
		initForce = false
	}()

	// First create
	err := runInit(nil, []string{"test-ws"})
	assert.NoError(t, err)

	// Reinitialize with force
	initForce = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runInit(nil, []string{"test-ws"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "reinitialized")
}

// TestRunInit_JSONModeInvalidName tests JSON error for invalid names
func TestRunInit_JSONModeInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	initJSON = true
	initForce = false
	defer func() {
		initJSON = false
		initForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test with path separator in name
	err := runInit(nil, []string{"invalid/name"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunInit_HumanModeInvalidName tests human output for invalid names
func TestRunInit_HumanModeInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	initJSON = false
	initForce = false
	defer func() {
		initJSON = false
		initForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test with reserved name
	err := runInit(nil, []string{"."})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}
