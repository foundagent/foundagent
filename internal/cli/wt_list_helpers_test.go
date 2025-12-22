package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectWorktreeStatus_PathNotFound(t *testing.T) {
	status, desc := detectWorktreeStatus("/nonexistent/path")
	assert.Equal(t, "error", status)
	assert.Contains(t, desc, "not found")
}

func TestDetectWorktreeStatus_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a non-git directory
	testPath := filepath.Join(tmpDir, "not-a-git-repo")
	err := os.MkdirAll(testPath, 0755)
	assert.NoError(t, err)
	
	status, desc := detectWorktreeStatus(testPath)
	assert.Equal(t, "error", status)
	assert.Contains(t, desc, "failed to check status")
}
