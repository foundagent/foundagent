package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestBareRepo creates a bare repository for testing
func setupTestBareRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "test-repo.git")

	// Initialize bare repo
	cmd := exec.Command("git", "init", "--bare", bareRepo)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create bare repo: %v", err)
	}

	// Create a temporary working directory to make initial commit
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work dir: %v", err)
	}

	// Initialize and make first commit
	exec.Command("git", "init", workDir).Run()
	exec.Command("git", "-C", workDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", workDir, "config", "user.name", "Test User").Run()

	readmePath := filepath.Join(workDir, "README.md")
	os.WriteFile(readmePath, []byte("# Test Repo"), 0644)

	exec.Command("git", "-C", workDir, "add", ".").Run()
	exec.Command("git", "-C", workDir, "commit", "-m", "Initial commit").Run()
	exec.Command("git", "-C", workDir, "branch", "-M", "main").Run()
	exec.Command("git", "-C", workDir, "remote", "add", "origin", bareRepo).Run()
	exec.Command("git", "-C", workDir, "push", "-u", "origin", "main").Run()

	return bareRepo
}

func TestBranchExists(t *testing.T) {
	bareRepo := setupTestBareRepo(t)

	tests := []struct {
		name       string
		branchName string
		want       bool
		wantErr    bool
	}{
		{
			name:       "existing branch",
			branchName: "main",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "non-existing branch",
			branchName: "nonexistent",
			want:       false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BranchExists(bareRepo, tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("BranchExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BranchExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateBranch(t *testing.T) {
	bareRepo := setupTestBareRepo(t)

	tests := []struct {
		name         string
		newBranch    string
		sourceBranch string
		wantErr      bool
	}{
		{
			name:         "create from main",
			newBranch:    "feature-1",
			sourceBranch: "main",
			wantErr:      false,
		},
		{
			name:         "create from non-existent source",
			newBranch:    "feature-2",
			sourceBranch: "nonexistent",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateBranch(bareRepo, tt.newBranch, tt.sourceBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBranch() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If successful, verify branch was created
			if !tt.wantErr {
				exists, _ := BranchExists(bareRepo, tt.newBranch)
				if !exists {
					t.Errorf("Branch %s was not created", tt.newBranch)
				}
			}
		})
	}
}

func TestDeleteBranch(t *testing.T) {
	bareRepo := setupTestBareRepo(t)

	// Create a test branch and merge it so it can be deleted without force
	CreateBranch(bareRepo, "to-delete", "main")

	tests := []struct {
		name       string
		branchName string
		force      bool
		wantErr    bool
	}{
		{
			name:       "delete existing branch with force",
			branchName: "to-delete",
			force:      true, // Use force since branch may not be fully merged
			wantErr:    false,
		},
		{
			name:       "delete non-existing branch",
			branchName: "nonexistent",
			force:      false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteBranch(bareRepo, tt.branchName, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteBranch() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If successful, verify branch was deleted
			if !tt.wantErr {
				exists, _ := BranchExists(bareRepo, tt.branchName)
				if exists {
					t.Errorf("Branch %s was not deleted", tt.branchName)
				}
			}
		})
	}
}

func TestGetBranches(t *testing.T) {
	bareRepo := setupTestBareRepo(t)

	// Create additional branches
	CreateBranch(bareRepo, "feature-1", "main")
	CreateBranch(bareRepo, "feature-2", "main")

	branches, err := GetBranches(bareRepo)
	if err != nil {
		t.Fatalf("GetBranches() error = %v", err)
	}

	expectedBranches := map[string]bool{
		"main":      true,
		"feature-1": true,
		"feature-2": true,
	}

	if len(branches) != len(expectedBranches) {
		t.Errorf("GetBranches() returned %d branches, want %d", len(branches), len(expectedBranches))
	}

	for _, branch := range branches {
		if !expectedBranches[branch] {
			t.Errorf("Unexpected branch: %s", branch)
		}
	}
}

func TestIsBranchMerged(t *testing.T) {
	bareRepo := setupTestBareRepo(t)

	// Create a branch that will be merged
	CreateBranch(bareRepo, "merged-branch", "main")

	tests := []struct {
		name       string
		branch     string
		baseBranch string
		want       bool
		wantErr    bool
	}{
		{
			name:       "same branch is considered merged",
			branch:     "main",
			baseBranch: "main",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "new branch based on main is merged",
			branch:     "merged-branch",
			baseBranch: "main",
			want:       true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsBranchMerged(bareRepo, tt.branch, tt.baseBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsBranchMerged() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsBranchMerged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDetachedHead(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize regular repo
	exec.Command("git", "init", repoPath).Run()
	exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", repoPath, "config", "user.name", "Test User").Run()

	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)

	exec.Command("git", "-C", repoPath, "add", ".").Run()
	exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit").Run()

	// Test normal HEAD
	detached, err := IsDetachedHead(repoPath)
	if err != nil {
		t.Errorf("IsDetachedHead() error = %v", err)
	}
	if detached {
		t.Error("IsDetachedHead() = true, want false for normal checkout")
	}

	// Checkout a specific commit to create detached HEAD
	exec.Command("git", "-C", repoPath, "checkout", "HEAD~0").Run()

	detached, err = IsDetachedHead(repoPath)
	if err != nil {
		t.Errorf("IsDetachedHead() error = %v", err)
	}
	if !detached {
		t.Error("IsDetachedHead() = false, want true for detached HEAD")
	}
}

func TestBranchExists_InvalidRepo(t *testing.T) {
	exists, err := BranchExists("/nonexistent/path", "main")
	if err == nil && exists {
		t.Error("BranchExists() should return false or error for invalid repo")
	}
	// It's ok if it returns false without error - that's a valid response
}

func TestGetBranches_InvalidRepo(t *testing.T) {
	_, err := GetBranches("/nonexistent/path")
	if err == nil {
		t.Error("GetBranches() should return error for invalid repo")
	}
}

func TestCreateBranch_DuplicateBranch(t *testing.T) {
	bareRepo := setupTestBareRepo(t)

	// Create branch first time
	err := CreateBranch(bareRepo, "duplicate", "main")
	if err != nil {
		t.Fatalf("First CreateBranch() failed: %v", err)
	}

	// Try to create same branch again
	err = CreateBranch(bareRepo, "duplicate", "main")
	if err == nil {
		t.Error("CreateBranch() should fail when creating duplicate branch")
	}
}
