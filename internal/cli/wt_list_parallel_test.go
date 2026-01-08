package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
)

func TestDetectStatusParallel(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test worktrees
	wt1Path := filepath.Join(tmpDir, "wt1")
	wt2Path := filepath.Join(tmpDir, "wt2")

	// Create wt1 with changes
	if err := os.MkdirAll(filepath.Join(wt1Path, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt1Path, "test.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create wt2 without .git (will error)
	if err := os.MkdirAll(wt2Path, 0755); err != nil {
		t.Fatal(err)
	}

	worktrees := []worktreeInfo{
		{Path: wt1Path, Branch: "main"},
		{Path: wt2Path, Branch: "dev"},
		{Path: "/nonexistent", Branch: "feature"},
	}

	result := detectStatusParallel(worktrees)

	if len(result) != 3 {
		t.Errorf("Expected 3 worktrees, got %d", len(result))
	}

	// All should have some status set
	for i, wt := range result {
		if wt.Status == "" {
			t.Errorf("Worktree %d has empty status", i)
		}
	}
}

func TestDetectStatusParallel_EmptyList(t *testing.T) {
	worktrees := []worktreeInfo{}
	result := detectStatusParallel(worktrees)

	if len(result) != 0 {
		t.Errorf("Expected empty list, got %d worktrees", len(result))
	}
}

func TestDetectStatusParallel_SingleWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	wtPath := filepath.Join(tmpDir, "wt")

	// Create valid git worktree structure
	if err := os.MkdirAll(filepath.Join(wtPath, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	worktrees := []worktreeInfo{
		{Path: wtPath, Branch: "main"},
	}

	result := detectStatusParallel(worktrees)

	if len(result) != 1 {
		t.Errorf("Expected 1 worktree, got %d", len(result))
	}
	if result[0].Status == "" {
		t.Error("Expected status to be set")
	}
}

func TestDiscoverWorktrees_Error(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create workspace without repos directory
	ws, err := workspace.New("test-ws", tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := ws.Create(false); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(ws.Path); err != nil {
		t.Fatal(err)
	}

	// Create empty config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{Name: "test-ws"},
		Repos:     []config.RepoConfig{},
	}

	// This should return empty list (no error for no repos)
	worktrees, err := discoverWorktrees(ws, cfg, "")
	if err != nil {
		t.Error("Expected no error with empty config")
	}
	if len(worktrees) != 0 {
		t.Error("Expected empty worktree list")
	}
}

func TestRunList_DiscoveryError(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	defer func() { listJSONFlag = false }()

	// Create directory that's not a workspace
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	listJSONFlag = false
	err := runList(nil, []string{})
	if err == nil {
		t.Error("Expected error when running list outside workspace")
	}
}
