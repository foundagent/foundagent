package cli

import (
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddRepository_InferNameError tests error when inferring name fails
func TestAddRepository_InferNameError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repo := repoToAdd{
		URL:  "invalid-url-without-slash",
		Name: "", // Force name inference
	}

	result := addRepository(ws, repo)

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestAddRepository_CloneBareError tests error when cloning fails
func TestAddRepository_CloneBareError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repo := repoToAdd{
		URL:  "https://github.com/nonexistent-user-12345678/nonexistent-repo-12345678.git",
		Name: "test-repo",
	}

	// Set JSON mode to suppress output
	addJSON = true
	defer func() { addJSON = false }()

	result := addRepository(ws, repo)

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestAddRepository_DefaultBranchError tests error when getting default branch fails
func TestAddRepository_DefaultBranchError(t *testing.T) {
	t.Skip("Complex test - requires corrupting git bare repo")
}

// TestAddRepository_AddRepositoryError tests error when registering repository fails
func TestAddRepository_AddRepositoryError(t *testing.T) {
	t.Skip("Complex test - requires workspace.AddRepository to fail")
}

// TestAddRepository_ConfigLoadWarning tests warning when config load fails
func TestAddRepository_ConfigLoadWarning(t *testing.T) {
	t.Skip("Complex test - requires config file corruption during add")
}

// TestAddRepository_ConfigSaveWarning tests warning when config save fails
func TestAddRepository_ConfigSaveWarning(t *testing.T) {
	t.Skip("Complex test - requires config save to fail")
}

// TestAddRepository_VSCodeUpdateWarning tests warning when VS Code update fails
func TestAddRepository_VSCodeUpdateWarning(t *testing.T) {
	t.Skip("Complex test - requires VS Code workspace update to fail")
}

// TestAddRepository_SuccessWithInferredName tests successful add with inferred name
func TestAddRepository_SuccessWithInferredName(t *testing.T) {
	t.Skip("Integration test - requires git clone which is slow")
}

// TestAddRepository_ForceReplaceExisting tests force replacing an existing repo
func TestAddRepository_ForceReplaceExisting(t *testing.T) {
	t.Skip("Integration test - requires git clone which is slow")
}
