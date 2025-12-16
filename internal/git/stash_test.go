package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStash(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Modify existing tracked file (README.md)
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Modified"), 0644)

	// Stash changes
	err := Stash(repoPath)
	if err != nil {
		t.Fatalf("Stash() error = %v", err)
	}

	// Verify worktree is clean
	isClean, _ := IsClean(repoPath)
	if !isClean {
		t.Error("Worktree should be clean after stash")
	}

	// Verify stash exists
	hasStash, _ := HasStash(repoPath)
	if !hasStash {
		t.Error("Stash should exist after stashing changes")
	}
}

func TestStash_NoChanges(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Stash with no changes (should not error)
	err := Stash(repoPath)
	if err != nil {
		t.Errorf("Stash() error = %v, want nil for clean worktree", err)
	}
}

func TestStashPop(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Modify existing tracked file
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Modified"), 0644)
	Stash(repoPath)

	// Verify clean state
	isClean, _ := IsClean(repoPath)
	if !isClean {
		t.Fatal("Worktree should be clean after stash")
	}

	// Pop stash
	err := StashPop(repoPath)
	if err != nil {
		t.Fatalf("StashPop() error = %v", err)
	}

	// Verify changes are back
	hasChanges, _ := HasUncommittedChanges(repoPath)
	if !hasChanges {
		t.Error("Changes should be restored after stash pop")
	}

	// Verify stash is empty
	hasStash, _ := HasStash(repoPath)
	if hasStash {
		t.Error("Stash should be empty after pop")
	}
}

func TestHasStash(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Initially no stash
	hasStash, err := HasStash(repoPath)
	if err != nil {
		t.Fatalf("HasStash() error = %v", err)
	}
	if hasStash {
		t.Error("HasStash() = true, want false for repo without stash")
	}

	// Modify existing tracked file and stash
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Modified"), 0644)
	Stash(repoPath)

	// Now should have stash
	hasStash, err = HasStash(repoPath)
	if err != nil {
		t.Fatalf("HasStash() error = %v", err)
	}
	if !hasStash {
		t.Error("HasStash() = false, want true after stashing")
	}
}

func TestStash_InvalidPath(t *testing.T) {
	err := Stash("/nonexistent/path")
	if err == nil {
		t.Error("Stash() should return error for invalid path")
	}
}

func TestStashPop_InvalidPath(t *testing.T) {
	err := StashPop("/nonexistent/path")
	if err == nil {
		t.Error("StashPop() should return error for invalid path")
	}
}

func TestStashPop_EmptyStash(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Try to pop with empty stash
	err := StashPop(repoPath)
	if err == nil {
		t.Error("StashPop() should return error when stash is empty")
	}
}

func TestHasStash_InvalidPath(t *testing.T) {
	_, err := HasStash("/nonexistent/path")
	if err == nil {
		t.Error("HasStash() should return error for invalid path")
	}
}

func TestStash_MultipleStashes(t *testing.T) {
	repoPath := setupTestWorktree(t)

	// Modify and stash first change
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Change 1"), 0644)
	Stash(repoPath)

	// Modify and stash second change
	os.WriteFile(readmePath, []byte("# Change 2"), 0644)
	Stash(repoPath)

	// Should have stash
	hasStash, err := HasStash(repoPath)
	if err != nil {
		t.Fatalf("HasStash() error = %v", err)
	}
	if !hasStash {
		t.Error("HasStash() = false, want true with multiple stashes")
	}

	// Worktree should be clean
	isClean, _ := IsClean(repoPath)
	if !isClean {
		t.Error("Worktree should be clean after stashing")
	}
}
