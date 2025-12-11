package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
