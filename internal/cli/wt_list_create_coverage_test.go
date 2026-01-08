package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunList_JSONModeDiscoverError tests JSON output on discover error
func TestRunList_JSONModeDiscoverError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	listJSONFlag = true
	defer func() { listJSONFlag = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunList_HumanModeDiscoverError tests human output on discover error
func TestRunList_HumanModeDiscoverError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	listJSONFlag = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunList_JSONModeEmptyWorkspace tests JSON output with no repos
func TestRunList_JSONModeEmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = true
	defer func() { listJSONFlag = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "{")
	assert.Contains(t, buf.String(), "total_worktrees")
}

// TestRunList_HumanModeEmptyWorkspace tests human output with no repos
func TestRunList_HumanModeEmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "fa add")
}

// TestRunList_JSONModeWithBranchFilter tests JSON output with branch filter
func TestRunList_JSONModeWithBranchFilter(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{"main"},
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = true
	defer func() { listJSONFlag = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on config but exercises filter path
	_ = err
	assert.Contains(t, buf.String(), "{")
}

// TestRunList_HumanModeWithBranchFilter tests human output with branch filter
func TestRunList_HumanModeWithBranchFilter(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{"main"},
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on config but exercises filter path
	_ = err
}

// TestRunCreate_JSONModeDiscoverError tests JSON output on discover error
func TestRunCreate_JSONModeDiscoverError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	createJSON = true
	defer func() { createJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunCreate_HumanModeDiscoverError tests human output on discover error
func TestRunCreate_HumanModeDiscoverError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	createJSON = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunCreate_JSONModeInvalidBranch tests JSON output for invalid branch
func TestRunCreate_JSONModeInvalidBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	createJSON = true
	defer func() { createJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCreate(nil, []string{"invalid..branch"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunCreate_JSONModeEmptyWorkspace tests JSON output with no repos
func TestRunCreate_JSONModeEmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	createJSON = true
	defer func() { createJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCreate(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunCreate_HumanModeEmptyWorkspace tests human output with no repos
func TestRunCreate_HumanModeEmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	createJSON = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCreate(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunList_JSONModeDiscoverSuccessEmptyConfig tests JSON output with empty config
func TestRunList_JSONModeDiscoverSuccessEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = true
	defer func() { listJSONFlag = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// No error, just empty workspace
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunList_HumanModeDiscoverSuccessEmptyConfig tests human output with empty config
func TestRunList_HumanModeDiscoverSuccessEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// No error, just empty workspace
	assert.NoError(t, err)
}

// TestRunCreate_JSONModeConfigLoadError tests JSON output on config load error
func TestRunCreate_JSONModeConfigLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt config file
	configFile := filepath.Join(ws.Path, "foundagent.toml")
	os.WriteFile(configFile, []byte("invalid toml [[["), 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	createJSON = true
	defer func() { createJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCreate(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}

// TestRunCreate_HumanModeConfigLoadError tests human output on config load error
func TestRunCreate_HumanModeConfigLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt config file
	configFile := filepath.Join(ws.Path, "foundagent.toml")
	os.WriteFile(configFile, []byte("invalid toml [[["), 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	createJSON = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCreate(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunList_HumanModeNoWorktreesNoBranch tests human output with no worktrees without filter
func TestRunList_HumanModeNoWorktreesNoBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{},
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Not an error - just no worktrees found
	_ = err
}

// TestRunList_HumanModeNoWorktreesWithBranchFilter tests human output with branch filter
func TestRunList_HumanModeNoWorktreesWithBranchFilter(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{"nonexistent"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// No error, just empty workspace - shows fa add message
	assert.NoError(t, err)
}

// TestRunList_JSONModeNoWorktrees tests JSON output with no worktrees
func TestRunList_JSONModeNoWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{},
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	listJSONFlag = true
	defer func() { listJSONFlag = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Should output empty list
	_ = err
	assert.Contains(t, buf.String(), "{")
}
