package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunList_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := listCmd
	err = cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

func TestRunList_NoWorkspace_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { listJSONFlag = false }()

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	listJSONFlag = true
	cmd := listCmd
	err = cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}
