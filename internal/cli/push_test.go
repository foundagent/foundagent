package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPushTestWorkspaceWithRemote creates a workspace with repos that have remotes
func setupPushTestWorkspaceWithRemote(t *testing.T, repoName string) (*workspace.Workspace, string) {
	t.Helper()
	dir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", dir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create repo directory structure
	repoDir := filepath.Join(ws.Path, workspace.ReposDir, repoName)
	bareDir := filepath.Join(repoDir, workspace.BareDir)
	worktreesDir := filepath.Join(repoDir, workspace.WorktreesDir)
	mainWorktree := filepath.Join(worktreesDir, "main")

	require.NoError(t, os.MkdirAll(bareDir, 0755))
	require.NoError(t, os.MkdirAll(worktreesDir, 0755))

	// Create a separate remote bare repo
	remoteDir := t.TempDir()
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = remoteDir
	require.NoError(t, cmd.Run())

	// Initialize local worktree
	require.NoError(t, os.MkdirAll(mainWorktree, 0755))
	cmd = exec.Command("git", "init")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	// Add remote
	cmd = exec.Command("git", "remote", "add", "origin", remoteDir)
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	// Create initial commit
	require.NoError(t, os.WriteFile(filepath.Join(mainWorktree, "README.md"), []byte("# "+repoName), 0644))
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	// Push to remote and set upstream
	cmd = exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	// Update workspace state
	state, err := ws.LoadState()
	require.NoError(t, err)
	if state.Repositories == nil {
		state.Repositories = make(map[string]*workspace.Repository)
	}
	state.Repositories[repoName] = &workspace.Repository{
		Name: repoName,
		URL:  remoteDir,
	}
	state.CurrentBranch = "main"
	require.NoError(t, ws.SaveState(state))

	return ws, mainWorktree
}

func TestPushCommand_NothingToPush(t *testing.T) {
	// Reset flags
	pushDryRun = false
	pushRepos = nil
	pushJSON = false
	pushVerbose = false
	pushForce = false

	ws, _ := setupPushTestWorkspaceWithRemote(t, "api")

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPush(pushCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "skipped")
	assert.Contains(t, output, "Nothing to push")
}

func TestPushCommand_WithUnpushedCommits(t *testing.T) {
	// Reset flags
	pushDryRun = false
	pushRepos = nil
	pushJSON = false
	pushVerbose = false
	pushForce = false

	ws, worktree := setupPushTestWorkspaceWithRemote(t, "api")

	// Create a new commit (unpushed)
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = worktree
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "New commit")
	cmd.Dir = worktree
	require.NoError(t, cmd.Run())

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPush(pushCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "pushed")
	assert.Contains(t, output, "api")
}

func TestPushCommand_JSONOutput(t *testing.T) {
	// Reset flags
	pushDryRun = false
	pushRepos = nil
	pushJSON = true
	pushVerbose = false
	pushForce = false

	ws, _ := setupPushTestWorkspaceWithRemote(t, "api")

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPush(pushCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	assert.Contains(t, result, "repos")
	assert.Contains(t, result, "summary")
}

func TestPushCommand_DryRun(t *testing.T) {
	// Reset flags
	pushDryRun = true
	pushRepos = nil
	pushJSON = false
	pushVerbose = false
	pushForce = false

	ws, worktree := setupPushTestWorkspaceWithRemote(t, "api")

	// Create a new commit (unpushed)
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = worktree
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "New commit")
	cmd.Dir = worktree
	require.NoError(t, cmd.Run())

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPush(pushCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "DRY RUN")
	assert.Contains(t, output, "would-push")
}

func TestPushCommand_ForceWithJSON(t *testing.T) {
	// Reset flags - force + json should error
	pushDryRun = false
	pushRepos = nil
	pushJSON = true
	pushVerbose = false
	pushForce = true

	ws, _ := setupPushTestWorkspaceWithRemote(t, "api")

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	err := runPush(pushCmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "force push requires confirmation")
}
