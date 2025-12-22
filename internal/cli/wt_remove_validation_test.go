package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRemove_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	
	cmd := removeCmd
	err = cmd.RunE(cmd, []string{"main"})
	assert.Error(t, err)
}

func TestRunRemove_NoWorkspace_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { removeJSON = false }()
	
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	
	removeJSON = true
	cmd := removeCmd
	err = cmd.RunE(cmd, []string{"main"})
	assert.Error(t, err)
}
