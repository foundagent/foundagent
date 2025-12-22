package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveConfig_Deprecated(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := New("test-save", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// SaveConfig should return deprecation error
	err = ws.SaveConfig(&Config{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deprecated")
}
