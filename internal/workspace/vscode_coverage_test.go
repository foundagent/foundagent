package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceWorktreeFolders_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	err = ws.ReplaceWorktreeFolders("feature")
	
	// Should error - no repositories
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No repositories configured")
}

func TestReplaceWorktreeFolders_Success(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main", "feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directories
	mainPath := ws.WorktreePath("test-repo", "main")
	featurePath := ws.WorktreePath("test-repo", "feature")
	require.NoError(t, os.MkdirAll(mainPath, 0755))
	require.NoError(t, os.MkdirAll(featurePath, 0755))

	// Add main folder to workspace
	vscodeData, _ := ws.LoadVSCodeWorkspace()
	relMainPath, _ := filepath.Rel(ws.Path, mainPath)
	vscodeData.Folders = append(vscodeData.Folders, VSCodeFolder{Path: relMainPath})
	ws.SaveVSCodeWorkspace(vscodeData)

	// Replace with feature branch
	err = ws.ReplaceWorktreeFolders("feature")
	require.NoError(t, err)

	// Verify folders were replaced
	reloaded, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)

	// Should have feature folder, not main
	hasFeature := false
	hasMain := false
	for _, folder := range reloaded.Folders {
		if filepath.Base(folder.Path) == "feature" {
			hasFeature = true
		}
		if filepath.Base(folder.Path) == "main" {
			hasMain = true
		}
	}

	assert.True(t, hasFeature, "Should have feature folder")
	assert.False(t, hasMain, "Should not have main folder")
}

func TestReplaceWorktreeFolders_PreservesNonWorktreeFolders(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"main"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Create worktree directory
	mainPath := ws.WorktreePath("test-repo", "main")
	require.NoError(t, os.MkdirAll(mainPath, 0755))

	// Add workspace root and worktree folder
	vscodeData, _ := ws.LoadVSCodeWorkspace()
	vscodeData.Folders = []VSCodeFolder{
		{Path: "."}, // Workspace root - should be preserved
		{Path: "repos/test-repo/wt/main"},
	}
	ws.SaveVSCodeWorkspace(vscodeData)

	// Replace folders
	err = ws.ReplaceWorktreeFolders("main")
	require.NoError(t, err)

	// Verify root folder is preserved
	reloaded, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)

	hasRoot := false
	for _, folder := range reloaded.Folders {
		if folder.Path == "." {
			hasRoot = true
		}
	}

	assert.True(t, hasRoot, "Should preserve workspace root folder")
}

func TestReplaceWorktreeFolders_NonexistentWorktrees(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository but don't create worktrees
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{"feature"},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	// Replace with feature branch (worktree doesn't exist)
	err = ws.ReplaceWorktreeFolders("feature")
	
	// Should succeed but not add any folders
	require.NoError(t, err)

	reloaded, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)

	// Should have no worktree folders
	for _, folder := range reloaded.Folders {
		assert.NotContains(t, folder.Path, "repos/")
	}
}

func TestReplaceWorktreeFolders_MultipleRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add two repositories
	repo1 := &Repository{
		Name:          "repo1",
		URL:           "https://github.com/test/repo1.git",
		DefaultBranch: "main",
		Worktrees:     []string{"feature"},
		BareRepoPath:  ws.BareRepoPath("repo1"),
	}
	repo2 := &Repository{
		Name:          "repo2",
		URL:           "https://github.com/test/repo2.git",
		DefaultBranch: "main",
		Worktrees:     []string{"feature"},
		BareRepoPath:  ws.BareRepoPath("repo2"),
	}
	require.NoError(t, ws.AddRepository(repo1))
	require.NoError(t, ws.AddRepository(repo2))

	// Create worktree directories
	feature1Path := ws.WorktreePath("repo1", "feature")
	feature2Path := ws.WorktreePath("repo2", "feature")
	require.NoError(t, os.MkdirAll(feature1Path, 0755))
	require.NoError(t, os.MkdirAll(feature2Path, 0755))

	// Replace folders
	err = ws.ReplaceWorktreeFolders("feature")
	require.NoError(t, err)

	// Verify both worktrees were added
	reloaded, err := ws.LoadVSCodeWorkspace()
	require.NoError(t, err)

	foundRepo1 := false
	foundRepo2 := false
	for _, folder := range reloaded.Folders {
		if filepath.Base(folder.Path) == "feature" && filepath.Base(filepath.Dir(filepath.Dir(folder.Path))) == "repo1" {
			foundRepo1 = true
		}
		if filepath.Base(folder.Path) == "feature" && filepath.Base(filepath.Dir(filepath.Dir(folder.Path))) == "repo2" {
			foundRepo2 = true
		}
	}

	assert.True(t, foundRepo1, "Should have repo1 worktree")
	assert.True(t, foundRepo2, "Should have repo2 worktree")
}

func TestLoadVSCodeWorkspace_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Write invalid JSON
	wsPath := ws.VSCodeWorkspacePath()
	err = os.WriteFile(wsPath, []byte("invalid json {{{"), 0644)
	require.NoError(t, err)

	_, err = ws.LoadVSCodeWorkspace()
	
	// Should error on invalid JSON
	assert.Error(t, err)
}

func TestLoadVSCodeWorkspace_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	
	// Don't create workspace - file won't exist

	_, err = ws.LoadVSCodeWorkspace()
	
	// Should error on missing file
	assert.Error(t, err)
}
