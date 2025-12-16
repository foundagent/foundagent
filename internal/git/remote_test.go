package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupRemoteTestRepo(t *testing.T) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	
	// Create "remote" repo
	remoteRepo := filepath.Join(tmpDir, "remote.git")
	exec.Command("git", "init", "--bare", remoteRepo).Run()

	// Create local working repo
	workRepo := filepath.Join(tmpDir, "work")
	exec.Command("git", "init", workRepo).Run()
	exec.Command("git", "-C", workRepo, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", workRepo, "config", "user.name", "Test User").Run()
	
	// Create initial commit
	readmePath := filepath.Join(workRepo, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", workRepo, "add", ".").Run()
	exec.Command("git", "-C", workRepo, "commit", "-m", "Initial commit").Run()
	exec.Command("git", "-C", workRepo, "branch", "-M", "main").Run()
	
	// Push to remote
	exec.Command("git", "-C", workRepo, "remote", "add", "origin", remoteRepo).Run()
	exec.Command("git", "-C", workRepo, "push", "-u", "origin", "main").Run()

	return workRepo, remoteRepo
}

func TestGetDefaultBranch(t *testing.T) {
	_, remoteRepo := setupRemoteTestRepo(t)

	// Clone the bare repo for testing
	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "test-bare.git")
	exec.Command("git", "clone", "--bare", remoteRepo, bareRepo).Run()

	branch, err := GetDefaultBranch(bareRepo)
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
	exec.Command("git", "-C", workRepo, "checkout", "-b", "feature-1").Run()
	os.WriteFile(filepath.Join(workRepo, "f1.txt"), []byte("feature 1"), 0644)
	exec.Command("git", "-C", workRepo, "add", ".").Run()
	exec.Command("git", "-C", workRepo, "commit", "-m", "Feature 1").Run()
	exec.Command("git", "-C", workRepo, "push", "-u", "origin", "feature-1").Run()

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
	exec.Command("git", "clone", remoteRepo, otherClone).Run()
	exec.Command("git", "-C", otherClone, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", otherClone, "config", "user.name", "Test User").Run()

	// Make and push changes from other clone
	testFile := filepath.Join(otherClone, "new-file.txt")
	os.WriteFile(testFile, []byte("new content"), 0644)
	exec.Command("git", "-C", otherClone, "add", ".").Run()
	exec.Command("git", "-C", otherClone, "commit", "-m", "Add new file").Run()
	exec.Command("git", "-C", otherClone, "push").Run()

	// Pull in original repo
	err := Pull(workRepo)
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Verify new file exists
	if _, err := os.Stat(filepath.Join(workRepo, "new-file.txt")); os.IsNotExist(err) {
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
	os.WriteFile(testFile, []byte("local"), 0644)
	exec.Command("git", "-C", workRepo, "add", ".").Run()
	exec.Command("git", "-C", workRepo, "commit", "-m", "Local commit").Run()

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
	Push(workRepo)

	// Make remote commit (behind)
	tmpDir := t.TempDir()
	otherClone := filepath.Join(tmpDir, "other-clone")
	exec.Command("git", "clone", remoteRepo, otherClone).Run()
	exec.Command("git", "-C", otherClone, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", otherClone, "config", "user.name", "Test User").Run()
	
	remoteFile := filepath.Join(otherClone, "remote-change.txt")
	os.WriteFile(remoteFile, []byte("remote"), 0644)
	exec.Command("git", "-C", otherClone, "add", ".").Run()
	exec.Command("git", "-C", otherClone, "commit", "-m", "Remote commit").Run()
	exec.Command("git", "-C", otherClone, "push").Run()

	// Fetch to update refs
	Fetch(workRepo)

	ahead, behind, err = GetAheadBehindCount(workRepo, "main")
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
