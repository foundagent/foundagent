package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
