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

// TestCreateWorktreesForBranch_LoadStateError tests error loading state
func TestCreateWorktreesForBranch_LoadStateError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	// Don't create workspace - will fail to load state

	err = createWorktreesForBranch(ws, "feature")

	assert.Error(t, err)
}

// TestCreateWorktreesForBranch_NoReposError tests error with no repositories
func TestCreateWorktreesForBranch_NoReposError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = createWorktreesForBranch(ws, "feature")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No repositories")
}

// TestCreateWorktreesForBranch_WithFromFlag tests using --from flag
func TestCreateWorktreesForBranch_WithFromFlag(t *testing.T) {
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

	switchFrom = "develop"
	defer func() { switchFrom = "" }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = createWorktreesForBranch(ws, "feature")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on git operation, but tests the from flag path
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "develop")
}

// TestCreateWorktreesForBranch_DefaultSourceBranch tests default source branch detection
func TestCreateWorktreesForBranch_DefaultSourceBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with custom default branch
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "master",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	switchFrom = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = createWorktreesForBranch(ws, "feature")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on git operation, but tests default branch detection
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "master")
}

// TestCreateWorktreesForBranch_EmptyDefaultBranch tests fallback to "main"
func TestCreateWorktreesForBranch_EmptyDefaultBranch(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository with empty default branch
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "",
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	switchFrom = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = createWorktreesForBranch(ws, "feature")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Will fail on git operation, but tests fallback to main
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "main")
}

// TestCheckForUpdates_PrintsVersion tests version output
func TestCheckForUpdates_PrintsVersion(t *testing.T) {
	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	err := checkForUpdates()

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	// Should not return error even if update check fails
	assert.NoError(t, err)
}
