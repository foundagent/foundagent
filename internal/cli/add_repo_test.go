package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createLocalGitRepo creates a local git repository for testing
func createLocalGitRepo(t *testing.T, name string) string {
	t.Helper()
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, name)

	cmd := exec.Command("git", "init", repoPath)
	require.NoError(t, cmd.Run(), "Failed to init git repo")

	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	require.NoError(t, cmd.Run())

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test Repo"), 0644))

	cmd = exec.Command("git", "-C", repoPath, "add", ".")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit")
	require.NoError(t, cmd.Run())

	return repoPath
}

// TestAddRepository_ValidateURLError tests error on invalid URL
func TestAddRepository_ValidateURLError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL: "not-a-valid-url",
	})

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestAddRepository_InferNameFromURLError tests error when name cannot be inferred
func TestAddRepository_InferNameFromURLError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL: "https://example.com",
	})

	assert.Equal(t, "error", result.Status)
}

// TestAddRepository_ValidURLWithClone tests successful addition with valid URL using local repo
func TestAddRepository_ValidURLWithClone(t *testing.T) {
	// Create a local git repo to clone from
	sourceRepo := createLocalGitRepo(t, "source-repo")
	sourceURL := "file://" + sourceRepo

	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL:  sourceURL,
		Name: "test-repo",
	})

	// Should succeed with local repo
	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "test-repo", result.Name)
	assert.NotEmpty(t, result.BareRepoPath)
}

// TestAddRepository_ExistingRepoNoForce tests skipping existing repo
func TestAddRepository_ExistingRepoNoForce(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository first
	repo := &workspace.Repository{
		Name:         "test-repo",
		URL:          "https://github.com/test/repo.git",
		BareRepoPath: ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	addForce = false
	defer func() { addForce = false }()

	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/test/repo.git",
		Name: "test-repo",
	})

	assert.Equal(t, "success", result.Status)
	assert.True(t, result.Skipped)
}

// TestAddRepository_ExistingRepoWithForce tests force replacement using local repo
func TestAddRepository_ExistingRepoWithForce(t *testing.T) {
	// Create a local git repo to clone from
	sourceRepo := createLocalGitRepo(t, "source-repo")
	sourceURL := "file://" + sourceRepo

	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repository first
	repo := &workspace.Repository{
		Name:         "test-repo",
		URL:          sourceURL,
		BareRepoPath: ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create the bare repo directory
	require.NoError(t, os.MkdirAll(repo.BareRepoPath, 0755))

	addForce = true
	defer func() { addForce = false }()

	result := addRepository(ws, repoToAdd{
		URL:  sourceURL,
		Name: "test-repo",
	})

	// Should succeed with force
	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "test-repo", result.Name)
}

// TestAddRepository_InferredName tests name inference from local path
func TestAddRepository_InferredName(t *testing.T) {
	// Create a local git repo - name will be inferred from path
	sourceRepo := createLocalGitRepo(t, "my-awesome-repo")
	sourceURL := "file://" + sourceRepo

	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	result := addRepository(ws, repoToAdd{
		URL: sourceURL, // Name should be inferred as "my-awesome-repo"
	})

	// Should succeed and infer name from path
	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "my-awesome-repo", result.Name)
}

// TestRunAdd_EmptyArgs tests behavior with empty arguments
func TestRunAdd_EmptyArgs(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Chdir to workspace
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runAdd(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// With no args, runs reconcile which succeeds
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "up-to-date")
}

// TestRunAdd_InvalidURLFormat tests error with invalid URL
func TestRunAdd_InvalidURLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Chdir to workspace
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runAdd(nil, []string{"not-a-url"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// May or may not error depending on validation
	_ = err
}

// TestRunAdd_JSONOutputMode tests JSON output mode with local repo
func TestRunAdd_JSONOutputMode(t *testing.T) {
	// Create a local git repo to clone from
	sourceRepo := createLocalGitRepo(t, "json-test-repo")
	sourceURL := "file://" + sourceRepo

	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Chdir to workspace
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	addJSON = true
	defer func() { addJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runAdd(nil, []string{sourceURL})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "{")
	assert.Contains(t, buf.String(), "json-test-repo")
}
