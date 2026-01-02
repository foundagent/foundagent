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

func TestRunRemove_ConfigLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace but don't create config
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	// Remove the config file to trigger error
	_ = os.Remove(ws.Path + "/.foundagent/foundagent.yaml")

	cmd := removeCmd
	err = cmd.RunE(cmd, []string{"feature"})

	assert.Error(t, err)
}

func TestRunRemove_NoReposError(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with empty config
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

	cmd := removeCmd
	err = cmd.RunE(cmd, []string{"feature"})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.Error(t, err)
	assert.Contains(t, output, "No repositories")
}

func TestRunRemove_NoReposJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { removeJSON = false }()

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

	removeJSON = true
	cmd := removeCmd
	err = cmd.RunE(cmd, []string{"feature"})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.Error(t, err)
}

func TestRunRemove_NoWorktreesFound(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with a repo config but no actual git repo
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

	// Create the repo directory but not as a git repo
	_ = os.MkdirAll(ws.Path+"/test-repo", 0755)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	cmd := removeCmd
	err = cmd.RunE(cmd, []string{"nonexistent-branch"})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Should error - either no worktrees found or git error
	assert.Error(t, err)
}

func TestRunRemove_NoWorktreesFoundJSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { removeJSON = false }()

	// Create workspace with a repo config but no actual git repo
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

	// Create the repo directory but not as a git repo
	_ = os.MkdirAll(ws.Path+"/test-repo", 0755)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	removeJSON = true
	cmd := removeCmd
	err = cmd.RunE(cmd, []string{"nonexistent-branch"})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.Error(t, err)
}
