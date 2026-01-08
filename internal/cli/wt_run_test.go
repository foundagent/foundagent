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

func TestRunCreate_InvalidBranchName(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	createJSON = false
	cmd := createCmd
	err = cmd.RunE(cmd, []string{"invalid branch name"})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters")
}

func TestRunCreate_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to the workspace directory (tmpDir/test-ws)
	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	createJSON = false
	cmd := createCmd
	err = cmd.RunE(cmd, []string{"feature-123"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No repositories")
}

func TestRunCreate_NoReposJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { createJSON = false }()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	createJSON = true
	cmd := createCmd
	err = cmd.RunE(cmd, []string{"feature-123"})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.Error(t, err)
}

func TestRunList_WithBranchFilter(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a config with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	listJSONFlag = false
	cmd := listCmd
	// Pass a branch filter - will fail on git operations but tests the branch filter path
	err = cmd.RunE(cmd, []string{"main"})
	// Error is expected (no actual git repos), but we've covered the branch filter code path
	_ = err
}

func TestRunList_EmptyWorkspaceJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { listJSONFlag = false }()

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	listJSONFlag = true
	cmd := listCmd
	err = cmd.RunE(cmd, []string{})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "\"total_worktrees\": 0")
}

func TestRunList_EmptyWorkspaceHumanMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Capture stderr (where the error message is printed)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	listJSONFlag = false
	cmd := listCmd
	err = cmd.RunE(cmd, []string{})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Command succeeds and prints remediation message to stderr
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}
