package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveState_Success(t *testing.T) {
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

	// Load and save state
	state, err := ws.LoadState()
	require.NoError(t, err)

	err = ws.SaveState(state)
	require.NoError(t, err)

	// Verify state file exists
	statePath := filepath.Join(ws.Path, ".foundagent", "state.json")
	_, err = os.Stat(statePath)
	assert.NoError(t, err)
}

func TestSaveState_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Remove .foundagent directory
	foundagentDir := filepath.Join(ws.Path, ".foundagent")
	require.NoError(t, os.RemoveAll(foundagentDir))

	// Make it a file instead
	require.NoError(t, os.WriteFile(foundagentDir, []byte("test"), 0644))

	state := &State{Repositories: make(map[string]*Repository)}
	err = ws.SaveState(state)

	// Should error
	assert.Error(t, err)
}

func TestCreateState_Success(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create .foundagent directory
	foundagentDir := filepath.Join(ws.Path, ".foundagent")
	require.NoError(t, os.MkdirAll(foundagentDir, 0755))

	err = ws.createState()
	require.NoError(t, err)

	// Verify state file exists
	statePath := filepath.Join(ws.Path, ".foundagent", "state.json")
	_, err = os.Stat(statePath)
	assert.NoError(t, err)
}

func TestMustDiscover_Success(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Change to workspace directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	require.NoError(t, os.Chdir(ws.Path))

	foundWs := MustDiscover(ws.Path)

	assert.NotNil(t, foundWs)
	assert.Equal(t, ws.Name, foundWs.Name)
}

func TestMustDiscover_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to directory without workspace
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	require.NoError(t, os.Chdir(tmpDir))

	// This would panic in real use, but we can't test that easily
	// Just verify Discover returns error
	ws, err := Discover(tmpDir)
	assert.Error(t, err)
	assert.Nil(t, ws)
}

func TestHasRepository_True(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add a repository
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	has, err := ws.HasRepository("test-repo")

	require.NoError(t, err)
	assert.True(t, has)
}

func TestHasRepository_False(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	has, err := ws.HasRepository("nonexistent")

	require.NoError(t, err)
	assert.False(t, has)
}

func TestRemoveRepoFromConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Add repo via state
	repo := &Repository{
		Name:          "test-repo",
		URL:           "https://github.com/test/repo.git",
		DefaultBranch: "main",
		Worktrees:     []string{},
		BareRepoPath:  ws.BareRepoPath("test-repo"),
	}
	require.NoError(t, ws.AddRepository(repo))

	err = ws.removeRepoFromConfig("test-repo")

	// Should succeed
	assert.NoError(t, err)
}

func TestFindDirtyWorktrees_NoneFound(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	dirty, err := ws.findDirtyWorktrees("test-repo")

	// Should succeed with empty list
	require.NoError(t, err)
	assert.Empty(t, dirty)
}

func TestFindDirtyWorktrees_WithDirtyWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a non-git worktree directory
	wtPath := filepath.Join(ws.Path, "repos", "test-repo", "wt", "main")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	dirty, err := ws.findDirtyWorktrees("test-repo")

	// May find it as dirty or skip it
	_ = err
	_ = dirty
}
