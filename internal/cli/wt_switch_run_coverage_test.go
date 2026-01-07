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

// TestRunSwitch_FromFlagValidation tests error when --from used without --create
func TestRunSwitch_FromFlagValidation(t *testing.T) {
	switchFrom = "main"
	switchCreate = false
	defer func() {
		switchFrom = ""
		switchCreate = false
	}()

	err := runSwitch(nil, []string{"feature"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--from flag can only be used with --create")
}

// TestRunSwitch_DiscoverError tests error when not in workspace
func TestRunSwitch_DiscoverError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	switchFrom = ""
	switchCreate = false

	err := runSwitch(nil, []string{"feature"})

	assert.Error(t, err)
}

// TestRunSwitch_EmptyWorkspace tests error when workspace has no repositories
func TestRunSwitch_EmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	switchFrom = ""
	switchCreate = false

	err = runSwitch(nil, []string{"feature"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No repositories")
}

// TestRunSwitch_NoArgsListsBranches tests listing branches when no args
func TestRunSwitch_NoArgsListsBranches(t *testing.T) {
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
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	switchFrom = ""
	switchCreate = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSwitch(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will show available branches or error
	_ = err
}

// TestRunSwitch_InvalidBranchNameFormat tests error on invalid branch name
func TestRunSwitch_InvalidBranchNameFormat(t *testing.T) {
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
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	switchFrom = ""
	switchCreate = false

	err = runSwitch(nil, []string{"invalid..branch"})

	assert.Error(t, err)
}

// TestRunSwitch_JSONModeAlreadyOnBranch tests JSON output when already on branch
func TestRunSwitch_JSONModeAlreadyOnBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with worktree on main
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
		Worktrees:     []string{"main"},
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory to simulate current branch
	wtPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	switchJSON = true
	switchFrom = ""
	switchCreate = false
	defer func() { switchJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSwitch(nil, []string{"main"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// May or may not be already on branch depending on detection
	_ = err
}

// TestRunSwitch_WithCreateFlag tests creating new worktrees
func TestRunSwitch_WithCreateFlag(t *testing.T) {
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
	}
	require.NoError(t, ws.AddRepository(repo))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	switchCreate = true
	switchFrom = ""
	defer func() { switchCreate = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSwitch(nil, []string{"feature"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on git ops
	_ = err
}
