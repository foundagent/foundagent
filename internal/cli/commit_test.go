package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCommitTestWorkspace creates a workspace with repos for commit testing
func setupCommitTestWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()
	dir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-ws", dir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	return ws
}

// addTestRepoToWorkspace adds a test repository to the workspace
func addTestRepoToWorkspace(t *testing.T, ws *workspace.Workspace, name string) string {
	t.Helper()

	// Create repo directory structure
	repoDir := filepath.Join(ws.Path, workspace.ReposDir, name)
	bareDir := filepath.Join(repoDir, workspace.BareDir)
	worktreesDir := filepath.Join(repoDir, workspace.WorktreesDir)
	mainWorktree := filepath.Join(worktreesDir, "main")

	require.NoError(t, os.MkdirAll(bareDir, 0755))
	require.NoError(t, os.MkdirAll(worktreesDir, 0755))

	// Initialize bare repo
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = bareDir
	require.NoError(t, cmd.Run())

	// Create worktree with initial commit
	cmd = exec.Command("git", "init")
	cmd.Dir = mainWorktree
	if err := os.MkdirAll(mainWorktree, 0755); err == nil {
		require.NoError(t, cmd.Run())
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	// Create initial commit
	require.NoError(t, os.WriteFile(filepath.Join(mainWorktree, "README.md"), []byte("# "+name), 0644))
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = mainWorktree
	require.NoError(t, cmd.Run())

	// Update workspace state
	state, err := ws.LoadState()
	require.NoError(t, err)
	if state.Repositories == nil {
		state.Repositories = make(map[string]*workspace.Repository)
	}
	state.Repositories[name] = &workspace.Repository{
		Name: name,
		URL:  "https://github.com/test/" + name,
	}
	state.CurrentBranch = "main"
	require.NoError(t, ws.SaveState(state))

	return mainWorktree
}

func TestCommitCommand_NoArgs(t *testing.T) {
	// Reset flags
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = false
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	addTestRepoToWorkspace(t, ws, "api")

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Test without message
	err := runCommit(commitCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit message cannot be empty")
}

func TestCommitCommand_WithMessage(t *testing.T) {
	// Reset flags
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = false
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Stage a change
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
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

	err := runCommit(commitCmd, []string{"Test commit message"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "committed")
	assert.Contains(t, output, "api")
}

func TestCommitCommand_NothingToCommit(t *testing.T) {
	// Reset flags
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = false
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Make sure worktree is clean - add another commit so there's nothing new
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = worktree
	output, _ := cmd.Output()
	// If there are uncommitted changes from setup, commit them
	if len(output) > 0 {
		cmd = exec.Command("git", "add", "-A")
		cmd.Dir = worktree
		cmd.Run()
		cmd = exec.Command("git", "commit", "-m", "cleanup")
		cmd.Dir = worktree
		cmd.Run()
	}

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCommit(commitCmd, []string{"Test message"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	outputStr := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, outputStr, "skipped")
	assert.Contains(t, outputStr, "Nothing to commit")
}

func TestCommitCommand_JSONOutput(t *testing.T) {
	// Reset flags
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Stage a change
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
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

	err := runCommit(commitCmd, []string{"Test commit"})

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
	assert.Contains(t, result, "message")
	assert.Equal(t, "Test commit", result["message"])
}

func TestCommitCommand_MessageFlag(t *testing.T) {
	// Reset flags and set -m flag
	commitMessage = "Message from flag"
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Stage a change
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
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

	// No positional arg, should use -m flag
	err := runCommit(commitCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	assert.Equal(t, "Message from flag", result["message"])
}

func TestCommitCommand_PositionalOverridesFlag(t *testing.T) {
	// Both positional and -m flag set
	commitMessage = "Message from flag"
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Stage a change
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
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

	// Positional should win
	err := runCommit(commitCmd, []string{"Positional message"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	assert.Equal(t, "Positional message", result["message"])
}

func TestCommitCommand_MultipleRepos(t *testing.T) {
	// Reset flags
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree1 := addTestRepoToWorkspace(t, ws, "api")
	worktree2 := addTestRepoToWorkspace(t, ws, "web")

	// Stage changes in both repos
	require.NoError(t, os.WriteFile(filepath.Join(worktree1, "api.txt"), []byte("api"), 0644))
	cmd := exec.Command("git", "add", "api.txt")
	cmd.Dir = worktree1
	require.NoError(t, cmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(worktree2, "web.txt"), []byte("web"), 0644))
	cmd = exec.Command("git", "add", "web.txt")
	cmd.Dir = worktree2
	require.NoError(t, cmd.Run())

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCommit(commitCmd, []string{"Multi-repo commit"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	repos := result["repos"].([]interface{})
	assert.Len(t, repos, 2)

	summary := result["summary"].(map[string]interface{})
	assert.Equal(t, float64(2), summary["committed"])
}

func TestCommitCommand_DryRun(t *testing.T) {
	// Test dry-run mode - should not create commits
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = true
	commitRepos = nil
	commitJSON = false
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Stage a change
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new"), 0644))
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = worktree
	require.NoError(t, cmd.Run())

	// Get the HEAD SHA before dry-run
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = worktree
	beforeSHA, err := cmd.Output()
	require.NoError(t, err)

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCommit(commitCmd, []string{"Dry run test"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "DRY RUN")

	// Verify HEAD SHA unchanged (no commit created)
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = worktree
	afterSHA, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, string(beforeSHA), string(afterSHA))

	// Verify staged changes still exist
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = worktree
	err = cmd.Run()
	assert.Error(t, err) // Should have exit code 1 (staged changes present)
}

func TestCommitCommand_RepoFilter(t *testing.T) {
	// Test --repo flag to filter to specific repos
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = []string{"api"}
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree1 := addTestRepoToWorkspace(t, ws, "api")
	worktree2 := addTestRepoToWorkspace(t, ws, "web")

	// Stage changes in both repos
	require.NoError(t, os.WriteFile(filepath.Join(worktree1, "api.txt"), []byte("api"), 0644))
	cmd := exec.Command("git", "add", "api.txt")
	cmd.Dir = worktree1
	require.NoError(t, cmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(worktree2, "web.txt"), []byte("web"), 0644))
	cmd = exec.Command("git", "add", "web.txt")
	cmd.Dir = worktree2
	require.NoError(t, cmd.Run())

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCommit(commitCmd, []string{"API only commit"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	// Should only have 1 repo (api), web should be filtered out
	repos := result["repos"].([]interface{})
	assert.Len(t, repos, 1)
	assert.Equal(t, "api", repos[0].(map[string]interface{})["name"])
}

func TestCommitCommand_RepoFilterInvalid(t *testing.T) {
	// Test --repo flag with invalid repo name
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = []string{"nonexistent"}
	commitJSON = false
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	addTestRepoToWorkspace(t, ws, "api")

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// This should error because "nonexistent" repo doesn't exist
	err := runCommit(commitCmd, []string{"Invalid repo test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestCommitCommand_AllFlag(t *testing.T) {
	// Test -a flag to stage and commit tracked files
	commitMessage = ""
	commitAll = true
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Modify existing tracked file without staging
	require.NoError(t, os.WriteFile(filepath.Join(worktree, "README.md"), []byte("# Modified"), 0644))

	// Verify nothing is staged
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = worktree
	err := cmd.Run()
	require.NoError(t, err) // Exit 0 means no staged changes

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCommit(commitCmd, []string{"Stage and commit"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	summary := result["summary"].(map[string]interface{})
	assert.Equal(t, float64(1), summary["committed"])
}

func TestCommitCommand_AmendCLI(t *testing.T) {
	// Test --amend flag via CLI
	commitMessage = ""
	commitAll = false
	commitAmend = true
	commitDryRun = false
	commitRepos = nil
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	worktree := addTestRepoToWorkspace(t, ws, "api")

	// Get the original commit message
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = worktree
	origMsg, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "Initial commit", strings.TrimSpace(string(origMsg)))

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCommit(commitCmd, []string{"Amended message"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Parse JSON
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	summary := result["summary"].(map[string]interface{})
	assert.Equal(t, float64(1), summary["committed"])

	// Verify the message changed
	cmd = exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = worktree
	newMsg, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "Amended message", strings.TrimSpace(string(newMsg)))
}

func TestCommitCommand_JSONErrorOutput(t *testing.T) {
	// Test JSON output when there's an error (nothing to commit)
	commitMessage = ""
	commitAll = false
	commitAmend = false
	commitDryRun = false
	commitRepos = nil
	commitJSON = true
	commitVerbose = false
	commitAllowDetached = false

	ws := setupCommitTestWorkspace(t)
	addTestRepoToWorkspace(t, ws, "api")
	// No staged changes

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(ws.Path)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCommit(commitCmd, []string{"Nothing to commit"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err) // No error, just skip

	// Parse JSON - should still be valid
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))

	// Should show skipped repo
	summary := result["summary"].(map[string]interface{})
	assert.Equal(t, float64(0), summary["committed"])
	assert.Equal(t, float64(1), summary["skipped"])
}
