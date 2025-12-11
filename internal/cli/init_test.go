package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		flags        map[string]string
		setupFunc    func(t *testing.T, dir string)
		expectError  bool
		validateFunc func(t *testing.T, dir string, output string)
	}{
		{
			name: "create new workspace successfully",
			args: []string{"test-workspace"},
			validateFunc: func(t *testing.T, dir string, output string) {
				wsPath := filepath.Join(dir, "test-workspace")

				// Check directory exists
				assert.DirExists(t, wsPath)

				// Check .foundagent directory
				assert.DirExists(t, filepath.Join(wsPath, ".foundagent"))

				// Check config file
				configPath := filepath.Join(wsPath, ".foundagent.yaml")
				assert.FileExists(t, configPath)

				// Check state file
				statePath := filepath.Join(wsPath, ".foundagent", "state.json")
				assert.FileExists(t, statePath)

				// Verify state is valid JSON
				stateData, err := os.ReadFile(statePath)
				require.NoError(t, err)
				var state map[string]interface{}
				require.NoError(t, json.Unmarshal(stateData, &state))

				// Check repos structure
				// Note: Individual repo directories (repos/<repo-name>/.bare/ and worktrees/)
				// are created when repos are added, not during init
				assert.DirExists(t, filepath.Join(wsPath, "repos"))

				// Check VS Code workspace file
				vscPath := filepath.Join(wsPath, "test-workspace.code-workspace")
				assert.FileExists(t, vscPath)

				// Verify VS Code workspace is valid JSON
				vscData, err := os.ReadFile(vscPath)
				require.NoError(t, err)
				var vsc map[string]interface{}
				require.NoError(t, json.Unmarshal(vscData, &vsc))
				assert.Contains(t, vsc, "folders")
			},
		},
		{
			name:        "error on empty name",
			args:        []string{""},
			expectError: true,
		},
		{
			name:        "error on invalid character",
			args:        []string{"test/workspace"},
			expectError: true,
		},
		{
			name:        "error on dot",
			args:        []string{"."},
			expectError: true,
		},
		{
			name:        "error on double-dot",
			args:        []string{".."},
			expectError: true,
		},
		{
			name: "error on existing workspace without force",
			args: []string{"existing"},
			setupFunc: func(t *testing.T, dir string) {
				wsPath := filepath.Join(dir, "existing")
				require.NoError(t, os.MkdirAll(filepath.Join(wsPath, ".foundagent"), 0755))
			},
			expectError: true,
		},
		{
			name: "reinitialize with force flag",
			args: []string{"existing"},
			flags: map[string]string{
				"force": "true",
			},
			setupFunc: func(t *testing.T, dir string) {
				wsPath := filepath.Join(dir, "existing")
				require.NoError(t, os.MkdirAll(filepath.Join(wsPath, ".foundagent"), 0755))
				// Create a repos directory to ensure it's preserved
				reposPath := filepath.Join(wsPath, "repos")
				require.NoError(t, os.MkdirAll(reposPath, 0755))
				testFile := filepath.Join(reposPath, "test.txt")
				require.NoError(t, os.WriteFile(testFile, []byte("preserve me"), 0644))
			},
			validateFunc: func(t *testing.T, dir string, output string) {
				wsPath := filepath.Join(dir, "existing")

				// Check that workspace was recreated
				assert.DirExists(t, filepath.Join(wsPath, ".foundagent"))
				assert.FileExists(t, filepath.Join(wsPath, ".foundagent.yaml"))

				// Check that repos directory and its content were preserved
				testFile := filepath.Join(wsPath, "repos", "test.txt")
				assert.FileExists(t, testFile)
				data, err := os.ReadFile(testFile)
				require.NoError(t, err)
				assert.Equal(t, "preserve me", string(data))
			},
		},
		{
			name: "json output format",
			args: []string{"json-test"},
			flags: map[string]string{
				"json": "true",
			},
			validateFunc: func(t *testing.T, dir string, output string) {
				var result map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(output), &result))

				assert.Equal(t, "success", result["status"])
				assert.Contains(t, result, "data")

				data := result["data"].(map[string]interface{})
				assert.Equal(t, "json-test", data["name"])
				assert.Equal(t, "created", data["action"])
				assert.Contains(t, data["path"], "json-test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Change to temp directory
			oldCwd, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldCwd) }()

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// Reset command flags
			initForce = false
			initJSON = false

			// Set flags
			if tt.flags != nil {
				if tt.flags["force"] == "true" {
					initForce = true
				}
				if tt.flags["json"] == "true" {
					initJSON = true
				}
			}

			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			err = runInit(initCmd, tt.args)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Check error expectation
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Run validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, tmpDir, output)
			}
		})
	}
}

func TestInitCommandHelp(t *testing.T) {
	// Verify that help text contains examples
	assert.Contains(t, initCmd.Example, "fa init my-project")
	assert.Contains(t, initCmd.Example, "--json")
	assert.Contains(t, initCmd.Example, "--force")
}
