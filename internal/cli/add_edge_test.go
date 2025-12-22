package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRepository_NameInference(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Test with URL but no name - should fail validation on invalid URL
	repo := repoToAdd{URL: "not-a-url", Name: ""}
	result := addRepository(ws, repo)

	assert.Equal(t, "error", result.Status)
}

func TestAddRepository_ForceWithExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	err = ws.Create(false)
	require.NoError(t, err)

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/org/test.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	err = ws.AddRepository(repo)
	require.NoError(t, err)

	// Create the bare repo directory to simulate existing installation
	err = os.MkdirAll(repo.BareRepoPath, 0755)
	require.NoError(t, err)

	// Create a file in it to test removal
	testFile := repo.BareRepoPath + "/test.txt"
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Try to add with force (but invalid URL so it won't clone)
	addForce = true
	defer func() { addForce = false }()

	repoToAdd := repoToAdd{URL: "invalid-url", Name: "test-repo"}
	result := addRepository(ws, repoToAdd)

	// Should fail on URL validation before it tries to remove
	assert.Equal(t, "error", result.Status)
}
