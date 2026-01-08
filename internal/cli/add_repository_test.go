package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRepository_CompletelyInvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Test with completely invalid URL
	result := addRepository(ws, repoToAdd{
		URL:  "not a valid url at all",
		Name: "test",
	})

	assert.Equal(t, "error", result.Status)
	assert.Contains(t, result.Error, "Invalid")
}

func TestAddRepository_EmptyNameInference(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Test with URL that can't infer a name (no path component)
	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com",
		Name: "",
	})

	assert.Equal(t, "error", result.Status)
}

func TestAddRepository_SkipExisting(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { addForce = false }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Create a fake existing repo directory
	repoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Add to state
	state, _ := ws.LoadState()
	if state.Repositories == nil {
		state.Repositories = make(map[string]*workspace.Repository)
	}
	state.Repositories["test-repo"] = &workspace.Repository{
		Name: "test-repo",
		URL:  "https://github.com/org/test-repo.git",
	}
	_ = ws.SaveState(state)

	addForce = false
	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/org/test-repo.git",
		Name: "test-repo",
	})

	assert.Equal(t, "success", result.Status)
	assert.True(t, result.Skipped)
}

func TestAddRepository_CloneFailure(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { addJSON = true }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Try to clone from a non-existent URL
	addJSON = true
	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/nonexistent/repo-that-does-not-exist-12345.git",
		Name: "test-repo",
	})

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

func TestAddRepository_ForceRemoveExisting(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() {
		addForce = false
		addJSON = true
	}()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Create a fake existing repo directory with a file
	repoPath := ws.BareRepoPath("test-repo")
	err = os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)
	err = os.WriteFile(repoPath+"/testfile", []byte("test"), 0644)
	require.NoError(t, err)

	// Add to state
	state, _ := ws.LoadState()
	if state.Repositories == nil {
		state.Repositories = make(map[string]*workspace.Repository)
	}
	state.Repositories["test-repo"] = &workspace.Repository{
		Name: "test-repo",
		URL:  "https://github.com/org/old-repo.git",
	}
	_ = ws.SaveState(state)

	// Try with force and non-existent URL (will fail clone but tests force path)
	addForce = true
	addJSON = true
	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/nonexistent/new-repo.git",
		Name: "test-repo",
	})

	// Should have removed the old directory
	_, err = os.Stat(repoPath + "/testfile")
	assert.True(t, os.IsNotExist(err))

	// Will fail on clone, but tests the force removal path
	assert.Equal(t, "error", result.Status)
}
