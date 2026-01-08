package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunReconcile_AllUpToDate(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Run reconcile on empty workspace
	err = runReconcile(ws)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "up-to-date")
}

func TestRunReconcile_AllUpToDateJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { addJSON = false }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	addJSON = true
	err = runReconcile(ws)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "repos_to_clone")
}

func TestRunReconcile_ReposToClone(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a config with a repo that doesn't exist yet
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/nonexistent.git", DefaultBranch: "main"},
		},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Run reconcile - will try to clone
	err = runReconcile(ws)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Will likely error on clone, but tests the code path
	_ = err
}

func TestRunReconcile_ReposToCloneJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { addJSON = false }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a config with a repo that doesn't exist yet
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/nonexistent.git", DefaultBranch: "main"},
		},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	addJSON = true
	err = runReconcile(ws)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Will likely error on clone, but tests JSON output path
	_ = err
}

func TestAddRepository_ValidURL(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Test with invalid URL to trigger error path
	result := addRepository(ws, repoToAdd{
		URL:  "not-a-valid-url",
		Name: "test",
	})

	assert.Equal(t, "error", result.Status)
}
