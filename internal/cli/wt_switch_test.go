package cli

import (
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
