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

func TestRunSwitch_FromFlagWithoutCreate(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() {
		switchFrom = ""
		switchCreate = false
	}()

	// Create workspace
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	switchFrom = "main"
	switchCreate = false

	cmd := switchCmd
	err = cmd.RunE(cmd, []string{"feature"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--from flag can only be used with --create")
}

func TestRunSwitch_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with no repos
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	cmd := switchCmd
	err = cmd.RunE(cmd, []string{"feature"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No repositories")
}

func TestRunSwitch_NoArgsListBranches(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with a repo
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

	cmd := switchCmd
	// No args should list branches - will fail on git operations but tests the code path
	_ = cmd.RunE(cmd, []string{})
}

func TestRunSwitch_InvalidBranchName(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace with a repo
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a config and state with a repo
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos: []config.RepoConfig{
			{Name: "test-repo", URL: "https://github.com/org/repo.git", DefaultBranch: "main"},
		},
	}
	err = config.Save(ws.Path, cfg)
	require.NoError(t, err)

	// Add to state as well
	state, _ := ws.LoadState()
	if state.Repositories == nil {
		state.Repositories = make(map[string]*workspace.Repository)
	}
	state.Repositories["test-repo"] = &workspace.Repository{
		Name:         "test-repo",
		BareRepoPath: ws.Path + "/test-repo",
		URL:          "https://github.com/org/repo.git",
	}
	_ = ws.SaveState(state)

	err = os.Chdir(ws.Path)
	require.NoError(t, err)

	cmd := switchCmd
	err = cmd.RunE(cmd, []string{"invalid branch name"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters")
}

func TestRunSwitch_JSONMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { switchJSON = false }()

	// Create workspace with no repos
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

	switchJSON = true
	cmd := switchCmd
	err = cmd.RunE(cmd, []string{"feature"})

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Will error due to no repos, but tests JSON mode path
	assert.Error(t, err)
}
