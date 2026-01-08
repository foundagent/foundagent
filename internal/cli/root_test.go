package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	// Verify root command exists
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "fa", rootCmd.Use)
}

func TestRootCommand_WithVersionFlag(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags
	showVersion = true

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run root command with version flag
	err := rootCmd.RunE(rootCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should succeed and show version
	assert.NoError(t, err)
	assert.Contains(t, output, "foundagent")
}

func TestRootCommand_NoArgs(t *testing.T) {
	// Reset flags
	showVersion = false

	// Run root command without args (should show help)
	err := rootCmd.RunE(rootCmd, []string{})

	// Should return nil (help is shown)
	assert.NoError(t, err)
}
