package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		wsName      string
		basePath    string
		expectError bool
	}{
		{
			name:        "valid workspace name",
			wsName:      "test-workspace",
			basePath:    "/tmp",
			expectError: false,
		},
		{
			name:        "empty name",
			wsName:      "",
			basePath:    "/tmp",
			expectError: true,
		},
		{
			name:        "invalid character slash",
			wsName:      "test/workspace",
			basePath:    "/tmp",
			expectError: true,
		},
		{
			name:        "dot name",
			wsName:      ".",
			basePath:    "/tmp",
			expectError: true,
		},
		{
			name:        "double-dot name",
			wsName:      "..",
			basePath:    "/tmp",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, err := New(tt.wsName, tt.basePath)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ws)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ws)
				assert.Equal(t, tt.wsName, ws.Name)
			}
		})
	}
}

func TestWorkspaceCreate(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("create new workspace", func(t *testing.T) {
		ws, err := New("test-ws", tmpDir)
		require.NoError(t, err)

		err = ws.Create(false)
		require.NoError(t, err)

		// Verify structure
		assert.DirExists(t, ws.Path)
		assert.DirExists(t, filepath.Join(ws.Path, FoundagentDir))
		assert.FileExists(t, ws.ConfigPath())
		assert.FileExists(t, ws.StatePath())
		assert.DirExists(t, filepath.Join(ws.Path, ReposDir))
		// Note: Individual repo directories (repos/<repo-name>/.bare/ and worktrees/)
		// are created when repos are added, not during workspace creation
		assert.FileExists(t, ws.VSCodeWorkspacePath())
	})

	t.Run("error on existing workspace without force", func(t *testing.T) {
		ws, err := New("test-existing", tmpDir)
		require.NoError(t, err)

		// Create workspace first time
		err = ws.Create(false)
		require.NoError(t, err)

		// Try to create again without force
		err = ws.Create(false)
		assert.Error(t, err)
	})

	t.Run("reinitialize with force", func(t *testing.T) {
		ws, err := New("test-force", tmpDir)
		require.NoError(t, err)

		// Create workspace first time
		err = ws.Create(false)
		require.NoError(t, err)

		// Create a test file in repos
		testFile := filepath.Join(ws.Path, ReposDir, "test.txt")
		err = os.WriteFile(testFile, []byte("preserve"), 0644)
		require.NoError(t, err)

		// Reinitialize with force
		err = ws.Create(true)
		require.NoError(t, err)

		// Verify test file was preserved
		assert.FileExists(t, testFile)
		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "preserve", string(data))
	})
}

func TestWorkspaceExists(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-exists", tmpDir)
	require.NoError(t, err)

	// Should not exist initially
	assert.False(t, ws.Exists())

	// Create workspace
	err = ws.Create(false)
	require.NoError(t, err)

	// Should exist now
	assert.True(t, ws.Exists())
}

func TestConfigOperations(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-config", tmpDir)
	require.NoError(t, err)

	err = ws.Create(false)
	require.NoError(t, err)

	t.Run("load config", func(t *testing.T) {
		config, err := ws.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-config", config.Name)
		assert.Equal(t, []string{}, config.Repos)
	})

	// SaveConfig is deprecated - test using new config package instead
	t.Run("save config using config package", func(t *testing.T) {
		cfg, err := config.Load(ws.Path)
		require.NoError(t, err)

		config.AddRepo(cfg, "https://github.com/org/repo1.git", "repo1", "main")
		config.AddRepo(cfg, "https://github.com/org/repo2.git", "repo2", "master")

		err = config.Save(ws.Path, cfg)
		require.NoError(t, err)

		// Load and verify
		loadedConfig, err := config.Load(ws.Path)
		require.NoError(t, err)
		assert.Len(t, loadedConfig.Repos, 2)
	})
}

func TestStateOperations(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-state", tmpDir)
	require.NoError(t, err)

	err = ws.Create(false)
	require.NoError(t, err)

	t.Run("load state", func(t *testing.T) {
		state, err := ws.LoadState()
		require.NoError(t, err)
		assert.NotNil(t, state)
	})

	t.Run("save state", func(t *testing.T) {
		state := &State{}
		err := ws.SaveState(state)
		require.NoError(t, err)

		// Load and verify
		loadedState, err := ws.LoadState()
		require.NoError(t, err)
		assert.NotNil(t, loadedState)
	})
}

func TestVSCodeWorkspaceOperations(t *testing.T) {
	tmpDir := t.TempDir()

	ws, err := New("test-vscode", tmpDir)
	require.NoError(t, err)

	err = ws.Create(false)
	require.NoError(t, err)

	t.Run("load VS Code workspace", func(t *testing.T) {
		vscode, err := ws.LoadVSCodeWorkspace()
		require.NoError(t, err)
		assert.NotNil(t, vscode)
		assert.Len(t, vscode.Folders, 1)
		assert.Equal(t, ".", vscode.Folders[0].Path)
	})

	t.Run("save VS Code workspace", func(t *testing.T) {
		vscode := &VSCodeWorkspace{
			Folders: []VSCodeFolder{
				{Path: "."},
				{Path: "repos/worktrees/repo1/main"},
			},
		}
		err := ws.SaveVSCodeWorkspace(vscode)
		require.NoError(t, err)

		// Load and verify
		loadedVSCode, err := ws.LoadVSCodeWorkspace()
		require.NoError(t, err)
		assert.Len(t, loadedVSCode.Folders, 2)
		assert.Equal(t, "repos/worktrees/repo1/main", loadedVSCode.Folders[1].Path)
	})
}

func TestCreateReposStructure(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("creates repos directory", func(t *testing.T) {
		ws, err := New("test-repos", tmpDir)
		require.NoError(t, err)

		// Create repos structure
		err = ws.createReposStructure(false)
		require.NoError(t, err)

		// Verify repos directory exists
		reposPath := filepath.Join(ws.Path, ReposDir)
		assert.DirExists(t, reposPath)
	})

	t.Run("preserves existing repos directory in force mode", func(t *testing.T) {
		ws, err := New("test-repos-force", tmpDir)
		require.NoError(t, err)

		// Create repos directory with a test file
		reposPath := filepath.Join(ws.Path, ReposDir)
		err = os.MkdirAll(reposPath, 0755)
		require.NoError(t, err)

		testFile := filepath.Join(reposPath, "test.txt")
		err = os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Create repos structure with force
		err = ws.createReposStructure(true)
		require.NoError(t, err)

		// Verify test file still exists
		assert.FileExists(t, testFile)
	})
}

func TestNew_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("name with special characters", func(t *testing.T) {
		_, err := New("test-workspace_123", tmpDir)
		assert.NoError(t, err)
	})

	t.Run("very long path", func(t *testing.T) {
		longName := "workspace-with-a-very-long-name-that-exceeds-normal-limits"
		ws, err := New(longName, tmpDir)
		assert.NoError(t, err)
		assert.Equal(t, longName, ws.Name)
	})
}
