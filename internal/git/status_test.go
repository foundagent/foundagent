package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestWorktree(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repo
	exec.Command("git", "init", repoPath).Run()
	exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", repoPath, "config", "user.name", "Test User").Run()

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", repoPath, "add", ".").Run()
	exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit").Run()

	return repoPath
}

func TestHasUncommittedChanges(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Initially should be clean
	hasChanges, err := HasUncommittedChanges(repoPath)
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error = %v", err)
	}
	if hasChanges {
		t.Error("HasUncommittedChanges() = true, want false for clean repo")
	}

	// Create a change
	testFile := filepath.Join(repoPath, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	hasChanges, err = HasUncommittedChanges(repoPath)
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error = %v", err)
	}
	if !hasChanges {
		t.Error("HasUncommittedChanges() = false, want true after creating file")
	}
}

func TestIsClean(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Initially should be clean
	isClean, err := IsClean(repoPath)
	if err != nil {
		t.Fatalf("IsClean() error = %v", err)
	}
	if !isClean {
		t.Error("IsClean() = false, want true for clean repo")
	}

	// Modify a file
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Modified"), 0644)

	isClean, err = IsClean(repoPath)
	if err != nil {
		t.Fatalf("IsClean() error = %v", err)
	}
	if isClean {
		t.Error("IsClean() = true, want false after modifying file")
	}
}

func TestHasUntrackedFiles(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Initially no untracked files
	hasUntracked, err := HasUntrackedFiles(repoPath)
	if err != nil {
		t.Fatalf("HasUntrackedFiles() error = %v", err)
	}
	if hasUntracked {
		t.Error("HasUntrackedFiles() = true, want false for repo without untracked files")
	}

	// Create untracked file
	testFile := filepath.Join(repoPath, "untracked.txt")
	os.WriteFile(testFile, []byte("untracked"), 0644)

	hasUntracked, err = HasUntrackedFiles(repoPath)
	if err != nil {
		t.Fatalf("HasUntrackedFiles() error = %v", err)
	}
	if !hasUntracked {
		t.Error("HasUntrackedFiles() = false, want true after creating untracked file")
	}
}

func TestGetModifiedFiles(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Initially no modified files
	files, err := GetModifiedFiles(repoPath)
	if err != nil {
		t.Fatalf("GetModifiedFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("GetModifiedFiles() returned %d files, want 0", len(files))
	}

	// Modify README
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Modified"), 0644)
	exec.Command("git", "-C", repoPath, "add", "README.md").Run()

	files, err = GetModifiedFiles(repoPath)
	if err != nil {
		t.Fatalf("GetModifiedFiles() error = %v", err)
	}
	if len(files) != 1 {
		t.Errorf("GetModifiedFiles() returned %d files, want 1", len(files))
	}
	if len(files) > 0 && files[0] != "README.md" {
		t.Errorf("GetModifiedFiles() returned %s, want README.md", files[0])
	}
}

func TestGetUntrackedFiles(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Initially no untracked files
	files, err := GetUntrackedFiles(repoPath)
	if err != nil {
		t.Fatalf("GetUntrackedFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("GetUntrackedFiles() returned %d files, want 0", len(files))
	}

	// Create untracked files
	testFile1 := filepath.Join(repoPath, "untracked1.txt")
	testFile2 := filepath.Join(repoPath, "untracked2.txt")
	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	files, err = GetUntrackedFiles(repoPath)
	if err != nil {
		t.Fatalf("GetUntrackedFiles() error = %v", err)
	}
	if len(files) != 2 {
		t.Errorf("GetUntrackedFiles() returned %d files, want 2", len(files))
	}
}

func TestHasConflicts(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// No conflicts initially
	hasConflicts, err := HasConflicts(repoPath)
	if err != nil {
		t.Fatalf("HasConflicts() error = %v", err)
	}
	if hasConflicts {
		t.Error("HasConflicts() = true, want false for repo without conflicts")
	}
}

func TestHasUncommittedChanges_InvalidPath(t *testing.T) {
	_, err := HasUncommittedChanges("/nonexistent/path")
	if err == nil {
		t.Error("HasUncommittedChanges() should return error for invalid path")
	}
}

func TestIsClean_InvalidPath(t *testing.T) {
	_, err := IsClean("/nonexistent/path")
	if err == nil {
		t.Error("IsClean() should return error for invalid path")
	}
}

func TestHasUntrackedFiles_InvalidPath(t *testing.T) {
	_, err := HasUntrackedFiles("/nonexistent/path")
	if err == nil {
		t.Error("HasUntrackedFiles() should return error for invalid path")
	}
}

func TestHasConflicts_InvalidPath(t *testing.T) {
	_, err := HasConflicts("/nonexistent/path")
	if err == nil {
		t.Error("HasConflicts() should return error for invalid path")
	}
}

func TestGetModifiedFiles_InvalidPath(t *testing.T) {
	_, err := GetModifiedFiles("/nonexistent/path")
	if err == nil {
		t.Error("GetModifiedFiles() should return error for invalid path")
	}
}

func TestGetUntrackedFiles_InvalidPath(t *testing.T) {
	_, err := GetUntrackedFiles("/nonexistent/path")
	if err == nil {
		t.Error("GetUntrackedFiles() should return error for invalid path")
	}
}

func TestModifiedFilesWithMultipleChanges(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Create new files, add and commit them first
	for i := 1; i <= 3; i++ {
		fileName := "file" + string(rune('0'+i)) + ".txt"
		testFile := filepath.Join(repoPath, fileName)
		os.WriteFile(testFile, []byte("initial"), 0644)
	}
	exec.Command("git", "-C", repoPath, "add", ".").Run()
	exec.Command("git", "-C", repoPath, "commit", "-m", "Add files").Run()

	// Now modify them
	for i := 1; i <= 3; i++ {
		fileName := "file" + string(rune('0'+i)) + ".txt"
		testFile := filepath.Join(repoPath, fileName)
		os.WriteFile(testFile, []byte("modified"), 0644)
	}

	files, err := GetModifiedFiles(repoPath)
	if err != nil {
		t.Fatalf("GetModifiedFiles() error = %v", err)
	}

	// Should have multiple modified files
	if len(files) < 3 {
		t.Errorf("GetModifiedFiles() returned %d files, expected at least 3", len(files))
	}
}
