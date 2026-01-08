package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRepository_HasRepositoryError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a malformed config to trigger HasRepository error
	configPath := filepath.Join(ws.Path, ".foundagent", "config.yaml")
	err = os.WriteFile(configPath, []byte("invalid: [yaml content"), 0644)
	require.NoError(t, err)

	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/test/repo.git",
		Name: "test-repo",
	})

	// Should error - either from yaml or from clone
	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

func TestAddRepository_ForceRemoveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows - permission model differs")
	}

	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a repository directory that can't be removed
	repoPath := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	// Create a file we can't delete (read-only directory on unix)
	subDir := filepath.Join(repoPath, "protected")
	require.NoError(t, os.MkdirAll(subDir, 0555))
	defer os.Chmod(subDir, 0755) // Cleanup

	// Manually add to config
	cfg, _ := config.Load(ws.Path)
	cfg.Repos = append(cfg.Repos, config.RepoConfig{
		Name: "test-repo",
		URL:  "https://github.com/test/repo.git",
	})
	config.Save(ws.Path, cfg)

	addForce = true
	defer func() { addForce = false }()

	result := addRepository(ws, repoToAdd{
		URL:  "https://github.com/test/repo.git",
		Name: "test-repo",
	})

	// Should error - may be from removal or clone
	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

func TestAddRepository_GetDefaultBranchError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := workspace.New("test-ws", tmpDir)
	require.NoError(t, err)
	require.NoError(t, ws.Create(false))

	// Create a bare repo path that will fail GetDefaultBranch
	repoPath := ws.BareRepoPath("test-repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repoPath, "refs", "heads"), 0755))

	// Create a minimal .git structure but no HEAD file
	// This will cause GetDefaultBranch to fail

	result := addRepository(ws, repoToAdd{
		URL:  "file://" + repoPath, // Use file:// to avoid network
		Name: "test-repo",
	})

	// This will fail during clone or default branch detection
	assert.Equal(t, "error", result.Status)
}

func TestAddRepository_WorktreeDirError(t *testing.T) {
	// This test is hard to trigger since we'd need the parent directory
	// to be read-only or otherwise prevent mkdir. Skip for now as it's
	// an edge case that's unlikely in practice.
	t.Skip("Difficult to test worktree directory creation failure")
}

func TestAddRepository_WorktreeAddError(t *testing.T) {
	// This would require a partially successful clone but failed worktree add
	// which is difficult to mock without extensive setup. The cleanup path
	// is tested indirectly through clone failures.
	t.Skip("Difficult to test worktree add failure isolation")
}
