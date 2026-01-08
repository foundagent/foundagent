package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupRemoteTestRepo(t *testing.T) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()

	// Create "remote" repo
	remoteRepo := filepath.Join(tmpDir, "remote.git")
	cmd := exec.Command("git", "init", "--bare", remoteRepo)
	require.NoError(t, cmd.Run(), "Failed to init bare repo")

	// Create local working repo
	workRepo := filepath.Join(tmpDir, "work")
	cmd = exec.Command("git", "init", workRepo)
	require.NoError(t, cmd.Run(), "Failed to init work repo")

	cmd = exec.Command("git", "-C", workRepo, "config", "user.email", "test@example.com")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", workRepo, "config", "user.name", "Test User")
	require.NoError(t, cmd.Run())

	// Create initial commit
	readmePath := filepath.Join(workRepo, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test"), 0644))

	cmd = exec.Command("git", "-C", workRepo, "add", ".")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", workRepo, "commit", "-m", "Initial commit")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", workRepo, "branch", "-M", "main")
	require.NoError(t, cmd.Run())

	// Push to remote
	cmd = exec.Command("git", "-C", workRepo, "remote", "add", "origin", remoteRepo)
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", workRepo, "push", "-u", "origin", "main")
	require.NoError(t, cmd.Run())

	return workRepo, remoteRepo
}

func TestGetDefaultBranch(t *testing.T) {
	_, remoteRepo := setupRemoteTestRepo(t)

	// The remote bare repo should have the default branch set
	// Use the remote repo directly instead of cloning
	branch, err := GetDefaultBranch(remoteRepo)
	if err != nil {
		t.Fatalf("GetDefaultBranch() error = %v", err)
	}

	// Should return "main" or similar default branch
	if branch != "main" && branch != "master" {
		t.Errorf("GetDefaultBranch() = %s, want main or master", branch)
	}
}

func TestListRemoteBranches(t *testing.T) {
	workRepo, _ := setupRemoteTestRepo(t)

	// Create additional branches
	cmd := exec.Command("git", "-C", workRepo, "checkout", "-b", "feature-1")
	require.NoError(t, cmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(workRepo, "f1.txt"), []byte("feature 1"), 0644))

	cmd = exec.Command("git", "-C", workRepo, "add", ".")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", workRepo, "commit", "-m", "Feature 1")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", workRepo, "push", "-u", "origin", "feature-1")
	require.NoError(t, cmd.Run())

	// Use the work repo with remote configured for listing
	branches, err := ListRemoteBranches(workRepo)
	if err != nil {
		t.Fatalf("ListRemoteBranches() error = %v", err)
	}

	// Should have at least main
	if len(branches) == 0 {
		t.Error("ListRemoteBranches() returned no branches")
	}

	// Check if main exists
	hasMain := false
	for _, branch := range branches {
		if branch == "main" {
			hasMain = true
			break
		}
	}
	if !hasMain {
		t.Errorf("ListRemoteBranches() missing main branch, got: %v", branches)
	}
}

func TestFetch(t *testing.T) {
	workRepo, _ := setupRemoteTestRepo(t)

	err := Fetch(workRepo)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
}

func TestFetch_InvalidRepo(t *testing.T) {
	err := Fetch("/nonexistent/path")
	if err == nil {
		t.Error("Fetch() should return error for invalid repo")
	}
}

func TestPull(t *testing.T) {
	workRepo, remoteRepo := setupRemoteTestRepo(t)

	// Create another clone to make changes
	tmpDir := t.TempDir()
	otherClone := filepath.Join(tmpDir, "other-clone")

	cmd := exec.Command("git", "clone", remoteRepo, otherClone)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to clone: %s", string(output))

	cmd = exec.Command("git", "-C", otherClone, "config", "user.email", "test@example.com")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", otherClone, "config", "user.name", "Test User")
	require.NoError(t, cmd.Run())

	// Make and push changes from other clone
	testFile := filepath.Join(otherClone, "new-file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("new content"), 0644))

	cmd = exec.Command("git", "-C", otherClone, "add", ".")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", otherClone, "commit", "-m", "Add new file")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to commit: %s", string(output))

	cmd = exec.Command("git", "-C", otherClone, "push")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to push from other clone: %s", string(output))

	// Fetch first to ensure we have the remote refs
	cmd = exec.Command("git", "-C", workRepo, "fetch", "origin")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to fetch: %s", string(output))

	// Pull in original repo
	err = Pull(workRepo)
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Verify new file exists
	pulledFile := filepath.Join(workRepo, "new-file.txt")
	if _, err := os.Stat(pulledFile); os.IsNotExist(err) {
		// Debug: list files and show git status
		files, _ := os.ReadDir(workRepo)
		t.Logf("Files in workRepo: %v", files)
		cmd = exec.Command("git", "-C", workRepo, "status")
		out, _ := cmd.CombinedOutput()
		t.Logf("Git status: %s", string(out))
		cmd = exec.Command("git", "-C", workRepo, "log", "--oneline", "-5")
		out, _ = cmd.CombinedOutput()
		t.Logf("Git log: %s", string(out))
		t.Error("Pulled file does not exist")
	}
}

func TestPull_InvalidRepo(t *testing.T) {
	err := Pull("/nonexistent/path")
	if err == nil {
		t.Error("Pull() should return error for invalid repo")
	}
}

func TestPush(t *testing.T) {
	workRepo, _ := setupRemoteTestRepo(t)

	// Make a local commit
	testFile := filepath.Join(workRepo, "push-test.txt")
	os.WriteFile(testFile, []byte("push test"), 0644)
	exec.Command("git", "-C", workRepo, "add", ".").Run()
	exec.Command("git", "-C", workRepo, "commit", "-m", "Test push").Run()

	// Push
	err := Push(workRepo)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
}

func TestPush_InvalidRepo(t *testing.T) {
	err := Push("/nonexistent/path")
	if err == nil {
		t.Error("Push() should return error for invalid repo")
	}
}

func TestGetAheadBehindCount(t *testing.T) {
	workRepo, remoteRepo := setupRemoteTestRepo(t)

	// Initially should be 0/0
	ahead, behind, err := GetAheadBehindCount(workRepo, "main")
	if err != nil {
		t.Fatalf("GetAheadBehindCount() error = %v", err)
	}
	if ahead != 0 || behind != 0 {
		t.Errorf("GetAheadBehindCount() = (%d, %d), want (0, 0)", ahead, behind)
	}

	// Make local commit (ahead)
	testFile := filepath.Join(workRepo, "local-change.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("local"), 0644))

	cmd := exec.Command("git", "-C", workRepo, "add", ".")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", workRepo, "commit", "-m", "Local commit")
	require.NoError(t, cmd.Run())

	ahead, behind, err = GetAheadBehindCount(workRepo, "main")
	if err != nil {
		t.Fatalf("GetAheadBehindCount() error = %v", err)
	}
	if ahead != 1 {
		t.Errorf("GetAheadBehindCount() ahead = %d, want 1", ahead)
	}
	if behind != 0 {
		t.Errorf("GetAheadBehindCount() behind = %d, want 0", behind)
	}

	// Push to sync
	err = Push(workRepo)
	require.NoError(t, err, "Failed to push")

	// Make remote commit (behind)
	tmpDir := t.TempDir()
	otherClone := filepath.Join(tmpDir, "other-clone")

	cmd = exec.Command("git", "clone", remoteRepo, otherClone)
	require.NoError(t, cmd.Run(), "Failed to clone")

	cmd = exec.Command("git", "-C", otherClone, "config", "user.email", "test@example.com")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", otherClone, "config", "user.name", "Test User")
	require.NoError(t, cmd.Run())

	remoteFile := filepath.Join(otherClone, "remote-change.txt")
	require.NoError(t, os.WriteFile(remoteFile, []byte("remote"), 0644))

	cmd = exec.Command("git", "-C", otherClone, "add", ".")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", otherClone, "commit", "-m", "Remote commit")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", otherClone, "push")
	require.NoError(t, cmd.Run(), "Failed to push from other clone")

	// Fetch to update refs
	err = Fetch(workRepo)
	require.NoError(t, err, "Failed to fetch")

	_, behind, err = GetAheadBehindCount(workRepo, "main")
	if err != nil {
		t.Fatalf("GetAheadBehindCount() error = %v", err)
	}
	if behind != 1 {
		t.Errorf("GetAheadBehindCount() behind = %d, want 1", behind)
	}
}

func TestGetAheadBehindCount_NoBranchTracking(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repo without remote
	exec.Command("git", "init", repoPath).Run()
	exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", repoPath, "config", "user.name", "Test User").Run()

	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", repoPath, "add", ".").Run()
	exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit").Run()

	// Should return 0, 0 for branch without tracking
	ahead, behind, err := GetAheadBehindCount(repoPath, "master")
	if err != nil {
		t.Fatalf("GetAheadBehindCount() error = %v", err)
	}
	if ahead != 0 || behind != 0 {
		t.Errorf("GetAheadBehindCount() = (%d, %d), want (0, 0) for untracked branch", ahead, behind)
	}
}

func TestListRemoteBranches_InvalidRepo(t *testing.T) {
	_, err := ListRemoteBranches("/nonexistent/path")
	if err == nil {
		t.Error("ListRemoteBranches() should return error for invalid repo")
	}
}
