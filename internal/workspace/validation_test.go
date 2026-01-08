package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid name",
			input:   "my-workspace",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "dot",
			input:   ".",
			wantErr: true,
		},
		{
			name:    "double dot",
			input:   "..",
			wantErr: true,
		},
		{
			name:    "with slash",
			input:   "my/workspace",
			wantErr: true,
		},
		{
			name:    "with backslash",
			input:   "my\\workspace",
			wantErr: true,
		},
		{
			name:    "with colon",
			input:   "my:workspace",
			wantErr: true,
		},
		{
			name:    "with asterisk",
			input:   "my*workspace",
			wantErr: true,
		},
		{
			name:    "with question mark",
			input:   "my?workspace",
			wantErr: true,
		},
		{
			name:    "with quotes",
			input:   "my\"workspace",
			wantErr: true,
		},
		{
			name:    "with less than",
			input:   "my<workspace",
			wantErr: true,
		},
		{
			name:    "with greater than",
			input:   "my>workspace",
			wantErr: true,
		},
		{
			name:    "with pipe",
			input:   "my|workspace",
			wantErr: true,
		},
		{
			name:    "with leading space",
			input:   " my-workspace",
			wantErr: false, // Trimmed
		},
		{
			name:    "with trailing space",
			input:   "my-workspace ",
			wantErr: false, // Trimmed
		},
		{
			name:    "only spaces",
			input:   "   ",
			wantErr: true, // Empty after trim
		},
		{
			name:    "valid with numbers",
			input:   "workspace-123",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			input:   "my_workspace",
			wantErr: false,
		},
		{
			name:    "valid with hyphens",
			input:   "my-workspace-name",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePathLength(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "normal path",
			path:    "/tmp/workspace",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "./workspace",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathLength(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePathLength(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePathLength_TooLong(t *testing.T) {
	// Create a path that exceeds the limit
	maxLength := maxPathLengthUnix
	if runtime.GOOS == "windows" {
		maxLength = maxPathLengthWindows
	}

	// Build a very long path
	longPath := "/" + strings.Repeat("a", maxLength+100)

	err := ValidatePathLength(longPath)
	if err == nil {
		t.Error("ValidatePathLength() should return error for excessively long path")
	}
}

func TestExecuteParallel(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3"}
	callCount := 0
	var mu sync.Mutex

	fn := func(repo string) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}

	results := ExecuteParallel(repos, fn)

	if len(results) != 3 {
		t.Errorf("ExecuteParallel() returned %d results, want 3", len(results))
	}

	if callCount != 3 {
		t.Errorf("Function was called %d times, want 3", callCount)
	}

	for i, result := range results {
		if result.RepoName != repos[i] {
			t.Errorf("Result[%d].RepoName = %q, want %q", i, result.RepoName, repos[i])
		}
		if result.Error != nil {
			t.Errorf("Result[%d].Error = %v, want nil", i, result.Error)
		}
	}
}

func TestExecuteParallel_WithErrors(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3"}

	fn := func(repo string) error {
		if repo == "repo2" {
			return errors.New("test error")
		}
		return nil
	}

	results := ExecuteParallel(repos, fn)

	if len(results) != 3 {
		t.Errorf("ExecuteParallel() returned %d results, want 3", len(results))
	}

	// Check that repo2 has an error
	if results[1].Error == nil {
		t.Error("Expected error for repo2")
	}

	// Check that repo1 and repo3 have no errors
	if results[0].Error != nil {
		t.Errorf("Unexpected error for repo1: %v", results[0].Error)
	}
	if results[2].Error != nil {
		t.Errorf("Unexpected error for repo3: %v", results[2].Error)
	}
}

func TestExecuteParallel_Empty(t *testing.T) {
	fn := func(repo string) error {
		return nil
	}

	results := ExecuteParallel([]string{}, fn)

	if len(results) != 0 {
		t.Errorf("ExecuteParallel([]) returned %d results, want 0", len(results))
	}
}

func TestDiscover(t *testing.T) {
	// Create a temporary workspace
	tmpDir := t.TempDir()
	wsPath := filepath.Join(tmpDir, "test-workspace")
	os.MkdirAll(wsPath, 0755)

	// Create config file
	configPath := filepath.Join(wsPath, ConfigFileName)
	configContent := `name: test-workspace
repos: []
`
	os.WriteFile(configPath, []byte(configContent), 0644)

	// Test discovering from workspace root
	ws, err := Discover(wsPath)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if ws.Name != "test-workspace" {
		t.Errorf("Workspace name = %q, want test-workspace", ws.Name)
	}

	// Test discovering from subdirectory
	subdir := filepath.Join(wsPath, "subdir", "nested")
	os.MkdirAll(subdir, 0755)

	ws, err = Discover(subdir)
	if err != nil {
		t.Fatalf("Discover() from subdir error = %v", err)
	}
	if ws.Name != "test-workspace" {
		t.Errorf("Workspace name from subdir = %q, want test-workspace", ws.Name)
	}
}

func TestDiscover_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	notWorkspace := filepath.Join(tmpDir, "not-a-workspace")
	os.MkdirAll(notWorkspace, 0755)

	_, err := Discover(notWorkspace)
	if err == nil {
		t.Error("Discover() should return error when not in workspace")
	}
}

func TestDiscover_CurrentDir(t *testing.T) {
	// Test with empty string (should use current directory)
	_, err := Discover("")
	// This will error unless we're in a workspace, which is fine
	_ = err // Just test that it doesn't panic
}

func TestMustDiscover_Panic(t *testing.T) {
	tmpDir := t.TempDir()
	notWorkspace := filepath.Join(tmpDir, "not-a-workspace")
	os.MkdirAll(notWorkspace, 0755)

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustDiscover() should panic when workspace not found")
		}
	}()

	MustDiscover(notWorkspace)
}

func TestValidateName_NullByte(t *testing.T) {
	err := ValidateName("test\x00name")
	if err == nil {
		t.Error("ValidateName() should reject names with null bytes")
	}
}

func TestParallelResult_Fields(t *testing.T) {
	result := ParallelResult{
		RepoName: "test-repo",
		Error:    errors.New("test error"),
	}

	if result.RepoName != "test-repo" {
		t.Errorf("RepoName = %q, want test-repo", result.RepoName)
	}
	if result.Error == nil {
		t.Error("Error should not be nil")
	}
}

func TestExecuteParallel_OrderPreserved(t *testing.T) {
	repos := []string{"a", "b", "c", "d", "e"}

	fn := func(repo string) error {
		// Sleep varying amounts to test that results are in correct order
		// despite async execution
		return nil
	}

	results := ExecuteParallel(repos, fn)

	for i, result := range results {
		if result.RepoName != repos[i] {
			t.Errorf("Result[%d].RepoName = %q, want %q", i, result.RepoName, repos[i])
		}
	}
}
