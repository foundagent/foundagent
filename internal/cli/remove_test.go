package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveCommandExists(t *testing.T) {
	// Verify remove command is registered
	assert.NotNil(t, repoRemoveCmd)
	assert.Equal(t, "remove <repo>...", repoRemoveCmd.Use)

	// Verify alias
	assert.Contains(t, repoRemoveCmd.Aliases, "rm")
}

func TestRemoveCommandFlags(t *testing.T) {
	// Verify all flags exist
	forceFlag := repoRemoveCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "bool", forceFlag.Value.Type())

	configOnlyFlag := repoRemoveCmd.Flags().Lookup("config-only")
	assert.NotNil(t, configOnlyFlag)
	assert.Equal(t, "bool", configOnlyFlag.Value.Type())

	jsonFlag := repoRemoveCmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "bool", jsonFlag.Value.Type())
}
