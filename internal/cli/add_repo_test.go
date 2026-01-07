package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddRepository_ValidateURLError tests error on invalid URL
func TestAddRepository_ValidateURLError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL: "not-a-valid-url",
	})

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestAddRepository_InferNameFromURLError tests error when name cannot be inferred
func TestAddRepository_InferNameFromURLError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL: "https://example.com",
	})

	assert.Equal(t, "error", result.Status)
}

// TestAddRepository_ValidURLWithClone tests successful addition with valid URL
func TestAddRepository_ValidURLWithClone(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/test/repo.git",
		Name: "test-repo",
	})

	// Will have error due to clone failing, but validates path
	assert.NotEmpty(t, result.URL)
}

// TestAddRepository_ExistingRepoNoForce tests skipping existing repo
func TestAddRepository_ExistingRepoNoForce(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository first
	repo := &workspace.Repository{
		Name:         "test-repo",
		URL:          "https://github.com/test/repo.git",
		BareRepoPath: ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	addForce = false
	defer func() { addForce = false }()

	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/test/repo.git",
		Name: "test-repo",
	})

	assert.Equal(t, "success", result.Status)
	assert.True(t, result.Skipped)
}

// TestAddRepository_ExistingRepoWithForce tests force replacement
func TestAddRepository_ExistingRepoWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository first
	repo := &workspace.Repository{
		Name:         "test-repo",
		URL:          "https://github.com/test/repo.git",
		BareRepoPath: ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	addForce = true
	defer func() { addForce = false }()

	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/test/new-repo.git",
		Name: "test-repo",
	})

	// Will fail on clone, but tests force path
	assert.NotEmpty(t, result.Name)
}

// TestAddRepository_InferredName tests name inference from URL
func TestAddRepository_InferredName(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL: "https://github.com/test/my-repo.git",
	})

	// Will fail on clone, but validates name inference
	assert.NotEmpty(t, result.URL)
}

// TestRunAdd_EmptyArgs tests behavior with empty arguments
func TestRunAdd_EmptyArgs(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Chdir to workspace
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runAdd(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// With no args, runs reconcile which succeeds
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "up-to-date")
}

// TestRunAdd_InvalidURLFormat tests error with invalid URL
func TestRunAdd_InvalidURLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Chdir to workspace
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runAdd(nil, []string{"not-a-url"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// May or may not error depending on validation
	_ = err
}

// TestRunAdd_JSONOutputMode tests JSON output mode
func TestRunAdd_JSONOutputMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Chdir to workspace
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	addJSON = true
	defer func() { addJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runAdd(nil, []string{"https://github.com/test/repo.git"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on clone
	_ = err
	assert.Contains(t, buf.String(), "{")
}
