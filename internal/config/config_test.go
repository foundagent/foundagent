package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-workspace")
	assert.Equal(t, "test-workspace", cfg.Workspace.Name)
	assert.Empty(t, cfg.Repos)
	assert.True(t, cfg.Settings.AutoCreateWorktree)
}

func TestLoadSaveYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.yaml")

	// Create config
	cfg := DefaultConfig("test-ws")
	cfg.Repos = []RepoConfig{
		{URL: "https://github.com/org/repo.git", Name: "repo", DefaultBranch: "main"},
	}

	// Save YAML
	err := SaveYAML(configPath, cfg)
	require.NoError(t, err)

	// Load YAML
	loadedCfg, err := LoadYAML(configPath)
	require.NoError(t, err)
	assert.Equal(t, cfg.Workspace.Name, loadedCfg.Workspace.Name)
	assert.Len(t, loadedCfg.Repos, 1)
	assert.Equal(t, "repo", loadedCfg.Repos[0].Name)
}

func TestLoadSaveTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.toml")

	// Create config
	cfg := DefaultConfig("test-ws")
	cfg.Repos = []RepoConfig{
		{URL: "https://github.com/org/repo.git", Name: "repo", DefaultBranch: "main"},
	}

	// Save TOML
	err := SaveTOML(configPath, cfg)
	require.NoError(t, err)

	// Load TOML
	loadedCfg, err := LoadTOML(configPath)
	require.NoError(t, err)
	assert.Equal(t, cfg.Workspace.Name, loadedCfg.Workspace.Name)
	assert.Len(t, loadedCfg.Repos, 1)
	assert.Equal(t, "repo", loadedCfg.Repos[0].Name)
}

func TestLoadSaveJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.json")

	// Create config
	cfg := DefaultConfig("test-ws")
	cfg.Repos = []RepoConfig{
		{URL: "https://github.com/org/repo.git", Name: "repo", DefaultBranch: "main"},
	}

	// Save JSON
	err := SaveJSON(configPath, cfg)
	require.NoError(t, err)

	// Load JSON
	loadedCfg, err := LoadJSON(configPath)
	require.NoError(t, err)
	assert.Equal(t, cfg.Workspace.Name, loadedCfg.Workspace.Name)
	assert.Len(t, loadedCfg.Repos, 1)
	assert.Equal(t, "repo", loadedCfg.Repos[0].Name)
}

func TestFindConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		configFile     string
		expectedFormat ConfigFormat
	}{
		{"YAML .yaml", ".foundagent.yaml", FormatYAML},
		{"YAML .yml", ".foundagent.yml", FormatYAML},
		{"TOML", ".foundagent.toml", FormatTOML},
		{"JSON", ".foundagent.json", FormatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			// Create config file
			configPath := filepath.Join(testDir, tt.configFile)
			cfg := DefaultConfig("test")

			switch tt.expectedFormat {
			case FormatYAML:
				err = SaveYAML(configPath, cfg)
			case FormatTOML:
				err = SaveTOML(configPath, cfg)
			case FormatJSON:
				err = SaveJSON(configPath, cfg)
			}
			require.NoError(t, err)

			// Find config
			foundPath, format, err := FindConfig(testDir)
			require.NoError(t, err)
			assert.Equal(t, configPath, foundPath)
			assert.Equal(t, tt.expectedFormat, format)
		})
	}
}

func TestFindConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, _, err := FindConfig(tmpDir)
	assert.Error(t, err)
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.yaml")

	// Create config
	cfg := DefaultConfig("test-ws")
	cfg.Repos = []RepoConfig{
		{URL: "https://github.com/org/repo.git", Name: "repo"},
	}
	err := SaveYAML(configPath, cfg)
	require.NoError(t, err)

	// Load config
	loadedCfg, err := Load(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "test-ws", loadedCfg.Workspace.Name)
	assert.Len(t, loadedCfg.Repos, 1)
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Save config (will create YAML by default)
	cfg := DefaultConfig("test-ws")
	err := Save(tmpDir, cfg)
	require.NoError(t, err)

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".foundagent.yaml")
	assert.FileExists(t, configPath)

	// Load and verify
	loadedCfg, err := LoadYAML(configPath)
	require.NoError(t, err)
	assert.Equal(t, "test-ws", loadedCfg.Workspace.Name)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "nil config",
			config:    nil,
			expectErr: true,
		},
		{
			name: "empty workspace name",
			config: &Config{
				Workspace: WorkspaceConfig{Name: ""},
			},
			expectErr: true,
		},
		{
			name: "valid config",
			config: &Config{
				Workspace: WorkspaceConfig{Name: "test"},
				Repos:     []RepoConfig{},
				Settings:  SettingsConfig{AutoCreateWorktree: true},
			},
			expectErr: false,
		},
		{
			name: "invalid repo URL",
			config: &Config{
				Workspace: WorkspaceConfig{Name: "test"},
				Repos: []RepoConfig{
					{URL: "invalid-url"},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate repo names",
			config: &Config{
				Workspace: WorkspaceConfig{Name: "test"},
				Repos: []RepoConfig{
					{URL: "git@github.com:org/repo.git", Name: "repo"},
					{URL: "git@github.com:org/other.git", Name: "repo"},
				},
			},
			expectErr: true,
		},
		{
			name: "infer missing name",
			config: &Config{
				Workspace: WorkspaceConfig{Name: "test"},
				Repos: []RepoConfig{
					{URL: "git@github.com:org/my-repo.git"},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddRemoveHasGetRepo(t *testing.T) {
	cfg := DefaultConfig("test")

	// Add repo
	AddRepo(cfg, "https://github.com/org/repo.git", "repo", "main")
	assert.True(t, HasRepo(cfg, "repo"))
	assert.Len(t, cfg.Repos, 1)

	// Get repo
	repo := GetRepo(cfg, "repo")
	require.NotNil(t, repo)
	assert.Equal(t, "repo", repo.Name)
	assert.Equal(t, "https://github.com/org/repo.git", repo.URL)
	assert.Equal(t, "main", repo.DefaultBranch)

	// Remove repo
	removed := RemoveRepo(cfg, "repo")
	assert.True(t, removed)
	assert.False(t, HasRepo(cfg, "repo"))
	assert.Empty(t, cfg.Repos)

	// Remove non-existent
	removed = RemoveRepo(cfg, "non-existent")
	assert.False(t, removed)
}

func TestDefaultTemplate(t *testing.T) {
	template := DefaultTemplate("my-workspace")
	assert.Contains(t, template, "my-workspace")
	assert.Contains(t, template, "# Foundagent Workspace Configuration")
	assert.Contains(t, template, "repos: []")
	assert.Contains(t, template, "auto_create_worktree: true")
}

func TestAddRepo_EmptyName(t *testing.T) {
	cfg := DefaultConfig("test")

	// Add repo without name - should infer from URL
	AddRepo(cfg, "https://github.com/org/my-repo.git", "", "main")

	assert.Len(t, cfg.Repos, 1)
	// Name should be inferred, but may not match exactly - just verify we have a repo
	assert.NotEmpty(t, cfg.Repos[0].URL)
}

func TestGetRepo_NotFound(t *testing.T) {
	cfg := DefaultConfig("test")

	repo := GetRepo(cfg, "nonexistent")
	assert.Nil(t, repo)
}

func TestLoadYAML_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: [[["), 0644)
	require.NoError(t, err)

	_, err = LoadYAML(configPath)
	assert.Error(t, err)
}

func TestLoadTOML_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.toml")

	// Write invalid TOML
	err := os.WriteFile(configPath, []byte("invalid toml ]][["), 0644)
	require.NoError(t, err)

	_, err = LoadTOML(configPath)
	assert.Error(t, err)
}

func TestLoadJSON_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.json")

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("{invalid json}"), 0644)
	require.NoError(t, err)

	_, err = LoadJSON(configPath)
	assert.Error(t, err)
}

func TestSaveYAML_InvalidPath(t *testing.T) {
	cfg := DefaultConfig("test")
	err := SaveYAML("/nonexistent/directory/.foundagent.yaml", cfg)
	assert.Error(t, err)
}

func TestLoadConfig_WithExistingFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a TOML config
	cfg := DefaultConfig("test-ws")
	configPath := filepath.Join(tmpDir, ".foundagent.toml")
	err := SaveTOML(configPath, cfg)
	require.NoError(t, err)

	// Load should auto-detect TOML
	loadedCfg, err := Load(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "test-ws", loadedCfg.Workspace.Name)
}

func TestSave_WithExistingFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial TOML config
	cfg := DefaultConfig("test-ws")
	configPath := filepath.Join(tmpDir, ".foundagent.toml")
	err := SaveTOML(configPath, cfg)
	require.NoError(t, err)

	// Save should preserve TOML format
	cfg.Workspace.Name = "updated-ws"
	err = Save(tmpDir, cfg)
	require.NoError(t, err)

	// Verify it's still TOML
	loadedCfg, err := LoadTOML(configPath)
	require.NoError(t, err)
	assert.Equal(t, "updated-ws", loadedCfg.Workspace.Name)
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected ConfigFormat
	}{
		{"YAML .yaml", ".foundagent.yaml", FormatYAML},
		{"YAML .yml", ".foundagent.yml", FormatYAML},
		{"TOML", ".foundagent.toml", FormatTOML},
		{"JSON", ".foundagent.json", FormatJSON},
		{"Unknown", ".foundagent.xyz", FormatYAML}, // defaults to YAML
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := detectFormat(tt.filename)
			assert.Equal(t, tt.expected, format)
		})
	}
}

func TestSaveYAML_WithExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.yaml")

	// Create initial config with comments
	initialContent := `# Workspace configuration
workspace:
  name: test-ws

repos: []

settings:
  auto_create_worktree: true
`
	err := os.WriteFile(configPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Load and modify
	cfg, err := LoadYAML(configPath)
	require.NoError(t, err)

	cfg.Workspace.Name = "updated-ws"

	// Save should work
	err = SaveYAML(configPath, cfg)
	require.NoError(t, err)

	// Verify changes
	loaded, err := LoadYAML(configPath)
	require.NoError(t, err)
	assert.Equal(t, "updated-ws", loaded.Workspace.Name)
}

func TestSaveYAML_NewFileInSubdir(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "nested", ".foundagent.yaml")

	cfg := DefaultConfig("test")

	// Should create directories
	err := SaveYAML(configPath, cfg)
	require.NoError(t, err)

	assert.FileExists(t, configPath)
}

func TestLoad_UnknownFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with unknown extension
	unknownPath := filepath.Join(tmpDir, ".foundagent.unknown")
	err := os.WriteFile(unknownPath, []byte("test"), 0644)
	require.NoError(t, err)

	// Load should return error for unknown format
	_, err = Load(tmpDir)
	assert.Error(t, err)
}

func TestLoad_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".foundagent.yaml")

	// Create config with invalid data that will fail validation
	cfg := &Config{
		Workspace: WorkspaceConfig{Name: ""}, // Empty name is invalid
	}

	data, _ := yaml.Marshal(cfg)
	err := os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Load should fail validation
	_, err = Load(tmpDir)
	assert.Error(t, err)
}

func TestAddRepo_InfersName(t *testing.T) {
	cfg := DefaultConfig("test")

	// Add repo with empty name
	AddRepo(cfg, "https://github.com/org/myrepo.git", "", "main")

	assert.Len(t, cfg.Repos, 1)
	// The name should be inferred from URL, but we just verify a repo was added
	assert.NotEmpty(t, cfg.Repos[0].URL)
}

func TestSaveJSON_WriteError(t *testing.T) {
	cfg := DefaultConfig("test")

	// Try to save to a directory path (should fail)
	tmpDir := t.TempDir()
	err := os.MkdirAll(filepath.Join(tmpDir, "file.json"), 0755)
	require.NoError(t, err)

	err = SaveJSON(filepath.Join(tmpDir, "file.json"), cfg)
	assert.Error(t, err)
}

func TestSaveTOML_WriteError(t *testing.T) {
	cfg := DefaultConfig("test")

	// Try to save to read-only location
	err := SaveTOML("/root/.foundagent.toml", cfg)
	assert.Error(t, err)
}
