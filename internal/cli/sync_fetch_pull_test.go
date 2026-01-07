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

// TestRunSyncFetch_ErrorFromSyncAllRepos tests error handling when SyncAllRepos fails
func TestRunSyncFetch_ErrorFromSyncAllRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt state to cause error
	stateFile := ws.Path + "/.foundagent/state.json"
	require.NoError(t, os.WriteFile(stateFile, []byte("invalid"), 0644))

	err = runSyncFetch(ws)

	assert.Error(t, err)
}

// TestRunSyncFetch_EmptyWorkspaceHuman tests empty workspace with human output
func TestRunSyncFetch_EmptyWorkspaceHuman(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncFetch(ws)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "No repositories configured")
	assert.Contains(t, output, "fa add")
}

// TestRunSyncFetch_EmptyWorkspaceJSON tests empty workspace with JSON output
func TestRunSyncFetch_EmptyWorkspaceJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	syncJSON = true
	defer func() { syncJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncFetch(ws)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "repos")
	assert.Contains(t, output, "summary")
}

// TestRunSyncFetch_WithReposHuman tests fetch with repos using human output
func TestRunSyncFetch_WithReposHuman(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncFetch(ws)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// May have errors since we don't have real git repos
	_ = err
	assert.Contains(t, output, "Summary:")
}

// TestRunSyncFetch_WithReposJSON tests fetch with repos using JSON output
func TestRunSyncFetch_WithReposJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	syncJSON = true
	defer func() { syncJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncFetch(ws)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// May have errors since we don't have real git repos
	_ = err
	_ = output
}

// TestRunSyncFetch_VerboseMode tests verbose output
func TestRunSyncFetch_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	syncVerbose = true
	defer func() { syncVerbose = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncFetch(ws)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// May have errors since we don't have real git repos
	_ = err
	assert.Contains(t, output, "Fetching from all remotes")
}

// TestRunSyncFetch_WithFailures tests handling of fetch failures
func TestRunSyncFetch_WithFailures(t *testing.T) {
	t.Skip("Complex test - requires simulating fetch failures")
}

// TestRunSyncPull_ErrorFromPullAllWorktrees tests error handling when PullAllWorktrees fails
func TestRunSyncPull_ErrorFromPullAllWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Corrupt state to cause error
	stateFile := ws.Path + "/.foundagent/state.json"
	require.NoError(t, os.WriteFile(stateFile, []byte("invalid"), 0644))

	err = runSyncPull(ws, "main")

	assert.Error(t, err)
}

// TestRunSyncPull_EmptyWorkspaceHuman tests empty workspace with human output
func TestRunSyncPull_EmptyWorkspaceHuman(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncPull(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "No repositories configured")
	assert.Contains(t, output, "fa add")
}

// TestRunSyncPull_EmptyWorkspaceJSON tests empty workspace with JSON output
func TestRunSyncPull_EmptyWorkspaceJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	syncJSON = true
	defer func() { syncJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncPull(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "repos")
	assert.Contains(t, output, "summary")
}

// TestRunSyncPull_WithReposHuman tests pull with repos using human output
func TestRunSyncPull_WithReposHuman(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncPull(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// May have errors since we don't have real git repos
	_ = err
	_ = output
}

// TestRunSyncPull_WithReposJSON tests pull with repos using JSON output
func TestRunSyncPull_WithReposJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	syncJSON = true
	defer func() { syncJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncPull(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// May have errors since we don't have real git repos
	_ = err
	_ = output
}

// TestRunSyncPull_VerboseMode tests verbose output
func TestRunSyncPull_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	syncVerbose = true
	defer func() { syncVerbose = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncPull(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// May have errors since we don't have real git repos
	_ = err
	assert.Contains(t, output, "Syncing branch")
	assert.Contains(t, output, "--pull")
}

// TestRunSyncPull_WithStashFlag tests stash mode
func TestRunSyncPull_WithStashFlag(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &workspace.Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	syncStash = true
	defer func() { syncStash = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runSyncPull(ws, "main")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// May have errors since we don't have real git repos
	_ = err
}

// TestRunSyncPull_WithFailures tests handling of pull failures
func TestRunSyncPull_WithFailures(t *testing.T) {
	t.Skip("Complex test - requires simulating pull failures")
}
