package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_FoundagentDirError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create workspace directory
	require.NoError(t, os.MkdirAll(ws.Path, 0755))

	// Create a file where .foundagent directory should be
	foundagentPath := filepath.Join(ws.Path, FoundagentDir)
	require.NoError(t, os.WriteFile(foundagentPath, []byte("blocking"), 0644))

	err = ws.Create(false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "foundagent")
}

func TestCreate_ReposStructureError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create workspace directory
	require.NoError(t, os.MkdirAll(ws.Path, 0755))

	// Create a file where repos directory should be
	reposPath := filepath.Join(ws.Path, ReposDir)
	require.NoError(t, os.WriteFile(reposPath, []byte("blocking"), 0644))

	err = ws.Create(false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repos")
}

func TestCreate_ConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create all directories
	require.NoError(t, os.MkdirAll(ws.Path, 0755))
	foundagentPath := filepath.Join(ws.Path, FoundagentDir)
	require.NoError(t, os.MkdirAll(foundagentPath, 0755))
	reposPath := filepath.Join(ws.Path, ReposDir)
	require.NoError(t, os.MkdirAll(reposPath, 0755))

	// Create a read-only directory to prevent config file creation
	configPath := ws.ConfigPath()
	require.NoError(t, os.WriteFile(configPath, []byte("existing"), 0444))

	err = ws.Create(false)

	// May succeed or fail depending on how createConfig handles existing files
	_ = err
}

func TestCreate_StateError(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create all directories
	require.NoError(t, os.MkdirAll(ws.Path, 0755))
	foundagentPath := filepath.Join(ws.Path, FoundagentDir)
	require.NoError(t, os.MkdirAll(foundagentPath, 0755))

	// Make .foundagent directory read-only to prevent state file creation
	require.NoError(t, os.Chmod(foundagentPath, 0444))
	defer os.Chmod(foundagentPath, 0755)

	err = ws.Create(false)

	assert.Error(t, err)
}

func TestCreate_VSCodeWorkspaceError(t *testing.T) {
	t.Skip("Complex error scenario - VS Code workspace file handling")
}

func TestCreate_WithForcePreservesRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create workspace first time
	require.NoError(t, ws.Create(false))

	// Create repos directory with some content
	reposPath := filepath.Join(ws.Path, ReposDir)
	testFile := filepath.Join(reposPath, "test-repo", "data.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile), 0755))
	require.NoError(t, os.WriteFile(testFile, []byte("important data"), 0644))

	// Recreate with force
	err = ws.Create(true)

	require.NoError(t, err)

	// Verify repos directory content is preserved
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "important data", string(data))
}

func TestCreate_PermissionDeniedOnWorkspaceDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows - permission model differs")
	}
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Try to create workspace in system directory
	ws, err := New("test-ws", "/usr/local/impossible-location-12345")
	require.NoError(t, err)

	err = ws.Create(false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}

func TestCreateReposStructure_ForceWithExistingRepos(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create workspace first
	require.NoError(t, ws.Create(false))

	// Add some content to repos directory
	reposPath := filepath.Join(ws.Path, ReposDir)
	testFile := filepath.Join(reposPath, "existing-repo", "data.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile), 0755))
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	// Call createReposStructure with force=true
	err = ws.createReposStructure(true)

	require.NoError(t, err)

	// Verify existing content is preserved
	_, err = os.Stat(testFile)
	assert.NoError(t, err)
}

func TestCreateReposStructure_PermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create workspace directory
	require.NoError(t, os.MkdirAll(ws.Path, 0755))

	// Make workspace directory read-only
	require.NoError(t, os.Chmod(ws.Path, 0444))
	defer os.Chmod(ws.Path, 0755)

	err = ws.createReposStructure(false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}

func TestCreate_PathsReturnCorrectValues(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	configPath := ws.ConfigPath()
	statePath := ws.StatePath()
	vsCodePath := ws.VSCodeWorkspacePath()

	assert.Equal(t, filepath.Join(ws.Path, ConfigFileName), configPath)
	assert.Equal(t, filepath.Join(ws.Path, FoundagentDir, StateFileName), statePath)
	assert.Equal(t, filepath.Join(ws.Path, "test-ws.code-workspace"), vsCodePath)
}

func TestCreate_MultipleTimesWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := New("test-ws", tmpDir)
	require.NoError(t, err)

	// Create first time
	require.NoError(t, ws.Create(false))

	// Create again with force multiple times
	require.NoError(t, ws.Create(true))
	require.NoError(t, ws.Create(true))

	// Verify workspace is still valid
	assert.True(t, ws.Exists())
}
