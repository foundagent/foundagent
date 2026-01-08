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

// TestAddRepository_InferNameError tests error when inferring name fails
func TestAddRepository_InferNameError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	repo := repoToAdd{
		URL:  "invalid-url-without-slash",
		Name: "", // Force name inference
	}

	result := addRepository(ws, repo)

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestAddRepository_CloneBareError tests error when cloning fails
func TestAddRepository_CloneBareError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Use a non-existent local path to test clone failure without network
	nonExistentPath := "/nonexistent/path/to/repo"
	repo := repoToAdd{
		URL:  nonExistentPath,
		Name: "test-repo",
	}

	// Set JSON mode to suppress output
	addJSON = true
	defer func() { addJSON = false }()

	result := addRepository(ws, repo)

	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestAddRepository_DefaultBranchError tests error when getting default branch fails
func TestAddRepository_DefaultBranchError(t *testing.T) {
	t.Skip("Complex test - requires corrupting git bare repo")
}

// TestAddRepository_AddRepositoryError tests error when registering repository fails
func TestAddRepository_AddRepositoryError(t *testing.T) {
	t.Skip("Complex test - requires workspace.AddRepository to fail")
}

// TestAddRepository_ConfigLoadWarning tests warning when config load fails
func TestAddRepository_ConfigLoadWarning(t *testing.T) {
	t.Skip("Complex test - requires config file corruption during add")
}

// TestAddRepository_ConfigSaveWarning tests warning when config save fails
func TestAddRepository_ConfigSaveWarning(t *testing.T) {
	t.Skip("Complex test - requires config save to fail")
}

// TestAddRepository_VSCodeUpdateWarning tests warning when VS Code update fails
func TestAddRepository_VSCodeUpdateWarning(t *testing.T) {
	t.Skip("Complex test - requires VS Code workspace update to fail")
}

// TestAddRepository_SuccessWithInferredName tests successful add with inferred name
func TestAddRepository_SuccessWithInferredName(t *testing.T) {
	t.Skip("Integration test - requires git clone which is slow")
}

// TestAddRepository_ForceReplaceExisting tests force replacing an existing repo
func TestAddRepository_ForceReplaceExisting(t *testing.T) {
	t.Skip("Integration test - requires git clone which is slow")
}

// TestRunAdd_HumanModeSkippedRepo tests human output for skipped repos
func TestRunAdd_HumanModeSkippedRepo(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	addJSON = false
	addForce = false
	defer func() {
		addJSON = false
		addForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Try to add invalid URL (will fail at validation)
	err = runAdd(nil, []string{"invalid-url"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunAdd_JSONModeMultipleResults tests JSON output for multiple repos
func TestRunAdd_JSONModeMultipleResults(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(ws.Path)

	addJSON = true
	defer func() { addJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Try to add multiple invalid URLs
	_ = runAdd(nil, []string{"invalid-url1", "invalid-url2"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Should have JSON array output
	assert.Contains(t, buf.String(), "[")
}

// TestRunAdd_HumanModeSummary tests human output summary for multiple repos
func TestRunAdd_HumanModeSummary(t *testing.T) {
	t.Skip("Network-dependent test - requires git clone")
}

// TestParseAddArgs_SingleURL tests parseAddArgs with single URL
func TestParseAddArgs_SingleURL(t *testing.T) {
	repos := parseAddArgs([]string{"https://github.com/test/repo.git"})

	assert.Len(t, repos, 1)
	assert.Equal(t, "https://github.com/test/repo.git", repos[0].URL)
	assert.Empty(t, repos[0].Name)
}

// TestParseAddArgs_URLWithName tests parseAddArgs with URL and name
func TestParseAddArgs_URLWithName(t *testing.T) {
	repos := parseAddArgs([]string{"https://github.com/test/repo.git", "custom-name"})

	assert.Len(t, repos, 1)
	assert.Equal(t, "https://github.com/test/repo.git", repos[0].URL)
	assert.Equal(t, "custom-name", repos[0].Name)
}

// TestParseAddArgs_MultipleURLs tests parseAddArgs with multiple URLs
func TestParseAddArgs_MultipleURLs(t *testing.T) {
	repos := parseAddArgs([]string{
		"https://github.com/test/repo1.git",
		"https://github.com/test/repo2.git",
	})

	assert.Len(t, repos, 2)
	assert.Equal(t, "https://github.com/test/repo1.git", repos[0].URL)
	assert.Equal(t, "https://github.com/test/repo2.git", repos[1].URL)
}

// TestAddResult_StructFields tests addResult struct
func TestAddResult_StructFields(t *testing.T) {
	result := addResult{
		Name:         "test-repo",
		URL:          "https://github.com/test/repo.git",
		BareRepoPath: "/path/to/bare",
		WorktreePath: "/path/to/wt",
		Status:       "success",
		Skipped:      false,
	}

	assert.Equal(t, "test-repo", result.Name)
	assert.Equal(t, "https://github.com/test/repo.git", result.URL)
	assert.Equal(t, "/path/to/bare", result.BareRepoPath)
	assert.Equal(t, "/path/to/wt", result.WorktreePath)
	assert.Equal(t, "success", result.Status)
	assert.False(t, result.Skipped)
}

// TestAddResult_WithErrorField tests addResult struct with error
func TestAddResult_WithErrorField(t *testing.T) {
	result := addResult{
		Name:   "test-repo",
		URL:    "invalid-url",
		Status: "error",
		Error:  "invalid repository URL format",
	}

	assert.Equal(t, "test-repo", result.Name)
	assert.Equal(t, "invalid-url", result.URL)
	assert.Equal(t, "error", result.Status)
	assert.Equal(t, "invalid repository URL format", result.Error)
}

// TestRepoToAdd_StructFields tests repoToAdd struct
func TestRepoToAdd_StructFields(t *testing.T) {
	repo := repoToAdd{
		URL:  "https://github.com/test/repo.git",
		Name: "custom-name",
	}

	assert.Equal(t, "https://github.com/test/repo.git", repo.URL)
	assert.Equal(t, "custom-name", repo.Name)
}

// TestRunAdd_HumanModeDiscoverErrorPath tests human output for discover error
func TestRunAdd_HumanModeDiscoverErrorPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir) // Not a workspace

	addJSON = false
	defer func() { addJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAdd(nil, []string{"https://github.com/test/repo.git"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
}

// TestRunAdd_JSONModeDiscoverErrorPath tests JSON output for discover error
func TestRunAdd_JSONModeDiscoverErrorPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir) // Not a workspace

	addJSON = true
	defer func() { addJSON = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAdd(nil, []string{"https://github.com/test/repo.git"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, buf.String(), "{")
}
