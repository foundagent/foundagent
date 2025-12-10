package cli

import (
	"testing"

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
