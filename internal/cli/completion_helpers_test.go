package cli

import (
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetWorktreeCompletions(t *testing.T) {
	// Test graceful degradation when not in workspace
	completions, directive := getWorktreeCompletions(nil, []string{}, "")

	assert.Empty(t, completions, "Should return empty completions when not in workspace")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive, "Should use NoFileComp directive")
}

func TestGetRepoCompletions(t *testing.T) {
	// Test graceful degradation when not in workspace
	completions, directive := getRepoCompletions(nil, []string{}, "")

	assert.Empty(t, completions, "Should return empty completions when not in workspace")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive, "Should use NoFileComp directive")
}

func TestGetBranchCompletions(t *testing.T) {
	// Test graceful degradation when not in workspace
	completions, directive := getBranchCompletions(nil, []string{}, "")

	assert.Empty(t, completions, "Should return empty completions when not in workspace")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive, "Should use NoFileComp directive")
}

// Note: Testing with actual workspace requires integration tests
// These unit tests verify graceful degradation behavior

func TestGetRepoCompletions_WithWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Add repositories
	repo1 := &workspace.Repository{
		Name:      "repo1",
		URL:       "https://github.com/test/repo1.git",
		Worktrees: []string{},
	}
	repo2 := &workspace.Repository{
		Name:      "repo2",
		URL:       "https://github.com/test/repo2.git",
		Worktrees: []string{},
	}
	assert.NoError(t, ws.AddRepository(repo1))
	assert.NoError(t, ws.AddRepository(repo2))

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Test completion
	completions, directive := getRepoCompletions(nil, []string{}, "")

	assert.Len(t, completions, 2, "Should return both repos")
	assert.Contains(t, completions, "repo1")
	assert.Contains(t, completions, "repo2")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestGetWorktreeCompletions_WithWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Add repository with worktrees
	repo := &workspace.Repository{
		Name:      "test-repo",
		URL:       "https://github.com/test/repo.git",
		Worktrees: []string{"main", "feature-1", "feature-2"},
	}
	assert.NoError(t, ws.AddRepository(repo))

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Test completion
	completions, directive := getWorktreeCompletions(nil, []string{}, "")

	assert.Len(t, completions, 3, "Should return all worktrees")
	assert.Contains(t, completions, "main")
	assert.Contains(t, completions, "feature-1")
	assert.Contains(t, completions, "feature-2")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestGetBranchCompletions_WithWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Add repository with worktrees (which represent branches)
	repo := &workspace.Repository{
		Name:      "test-repo",
		URL:       "https://github.com/test/repo.git",
		Worktrees: []string{"main", "develop"},
	}
	assert.NoError(t, ws.AddRepository(repo))

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Test completion
	completions, directive := getBranchCompletions(nil, []string{}, "")

	assert.Len(t, completions, 2, "Should return all branches")
	assert.Contains(t, completions, "main")
	assert.Contains(t, completions, "develop")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestGetRepoCompletions_WithPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	// Create workspace
	ws, err := workspace.New("test-workspace", tmpDir)
	assert.NoError(t, err)
	err = ws.Create(false)
	assert.NoError(t, err)

	// Add repositories
	repo1 := &workspace.Repository{
		Name:      "frontend-app",
		URL:       "https://github.com/test/frontend.git",
		Worktrees: []string{},
	}
	repo2 := &workspace.Repository{
		Name:      "backend-api",
		URL:       "https://github.com/test/backend.git",
		Worktrees: []string{},
	}
	assert.NoError(t, ws.AddRepository(repo1))
	assert.NoError(t, ws.AddRepository(repo2))

	// Change to workspace directory
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	_ = os.Chdir(ws.Path)

	// Test completion with prefix
	completions, directive := getRepoCompletions(nil, []string{}, "front")

	assert.Len(t, completions, 1, "Should return only matching repo")
	assert.Contains(t, completions, "frontend-app")
	assert.NotContains(t, completions, "backend-api")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}
