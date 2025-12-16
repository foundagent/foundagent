package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestClone(t *testing.T) {
	// Create a source repo to clone
	tmpDir := t.TempDir()
	sourceRepo := filepath.Join(tmpDir, "source")
	
	exec.Command("git", "init", sourceRepo).Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.name", "Test User").Run()
	
	readmePath := filepath.Join(sourceRepo, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", sourceRepo, "add", ".").Run()
	exec.Command("git", "-C", sourceRepo, "commit", "-m", "Initial commit").Run()

	// Clone it
	targetPath := filepath.Join(tmpDir, "cloned")
	opts := CloneOptions{
		URL:        sourceRepo,
		TargetPath: targetPath,
		Bare:       false,
		Progress:   false,
	}

	err := Clone(opts)
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}

	// Verify clone exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Cloned repository does not exist")
	}

	// Verify it's a valid git repo
	gitDir := filepath.Join(targetPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("Cloned repository is not a valid git repository")
	}
}

func TestCloneBare(t *testing.T) {
	// Create a source repo to clone
	tmpDir := t.TempDir()
	sourceRepo := filepath.Join(tmpDir, "source")
	
	exec.Command("git", "init", sourceRepo).Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.name", "Test User").Run()
	
	readmePath := filepath.Join(sourceRepo, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", sourceRepo, "add", ".").Run()
	exec.Command("git", "-C", sourceRepo, "commit", "-m", "Initial commit").Run()

	// Clone it as bare
	targetPath := filepath.Join(tmpDir, "cloned-bare.git")
	err := CloneBare(sourceRepo, targetPath, false)
	if err != nil {
		t.Fatalf("CloneBare() error = %v", err)
	}

	// Verify bare clone exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Bare cloned repository does not exist")
	}

	// Verify it's a bare repository (has refs, objects, etc but no .git subdir)
	refsDir := filepath.Join(targetPath, "refs")
	if _, err := os.Stat(refsDir); os.IsNotExist(err) {
		t.Error("Bare repository does not have refs directory")
	}

	// Should not have a .git subdirectory
	gitSubdir := filepath.Join(targetPath, ".git")
	if _, err := os.Stat(gitSubdir); !os.IsNotExist(err) {
		t.Error("Bare repository should not have .git subdirectory")
	}
}

func TestClone_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "should-fail")

	opts := CloneOptions{
		URL:        "https://github.com/nonexistent/totally-fake-repo-12345.git",
		TargetPath: targetPath,
		Bare:       false,
		Progress:   false,
	}

	err := Clone(opts)
	if err == nil {
		t.Error("Clone() should return error for invalid URL")
	}
}

func TestClone_WithProgress(t *testing.T) {
	// Create a source repo
	tmpDir := t.TempDir()
	sourceRepo := filepath.Join(tmpDir, "source")
	
	exec.Command("git", "init", sourceRepo).Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.name", "Test User").Run()
	
	readmePath := filepath.Join(sourceRepo, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", sourceRepo, "add", ".").Run()
	exec.Command("git", "-C", sourceRepo, "commit", "-m", "Initial commit").Run()

	// Clone with progress
	targetPath := filepath.Join(tmpDir, "cloned-progress")
	opts := CloneOptions{
		URL:        sourceRepo,
		TargetPath: targetPath,
		Bare:       false,
		Progress:   true,
	}

	err := Clone(opts)
	if err != nil {
		t.Fatalf("Clone() with progress error = %v", err)
	}

	// Verify clone exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Cloned repository does not exist")
	}
}

func TestCloneBare_WithProgress(t *testing.T) {
	// Create a source repo
	tmpDir := t.TempDir()
	sourceRepo := filepath.Join(tmpDir, "source")
	
	exec.Command("git", "init", sourceRepo).Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.name", "Test User").Run()
	
	readmePath := filepath.Join(sourceRepo, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", sourceRepo, "add", ".").Run()
	exec.Command("git", "-C", sourceRepo, "commit", "-m", "Initial commit").Run()

	// Clone bare with progress
	targetPath := filepath.Join(tmpDir, "cloned-bare-progress.git")
	err := CloneBare(sourceRepo, targetPath, true)
	if err != nil {
		t.Fatalf("CloneBare() with progress error = %v", err)
	}

	// Verify bare clone exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Bare cloned repository does not exist")
	}
}

func TestClone_EmptyRepo(t *testing.T) {
	// Create an empty source repo
	tmpDir := t.TempDir()
	sourceRepo := filepath.Join(tmpDir, "empty-source")
	exec.Command("git", "init", sourceRepo).Run()

	// Try to clone empty repo
	targetPath := filepath.Join(tmpDir, "cloned-empty")
	opts := CloneOptions{
		URL:        sourceRepo,
		TargetPath: targetPath,
		Bare:       false,
		Progress:   false,
	}

	// Cloning an empty repo should work but produce warnings
	err := Clone(opts)
	// Some git versions allow this, some don't - just verify the function handles it
	if err == nil {
		// If it succeeded, verify the directory exists
		if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
			t.Error("Clone succeeded but target directory doesn't exist")
		}
	}
}

func TestCloneBare_ExistingDirectory(t *testing.T) {
	// Create a source repo
	tmpDir := t.TempDir()
	sourceRepo := filepath.Join(tmpDir, "source")
	
	exec.Command("git", "init", sourceRepo).Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", sourceRepo, "config", "user.name", "Test User").Run()
	
	readmePath := filepath.Join(sourceRepo, "README.md")
	os.WriteFile(readmePath, []byte("# Test"), 0644)
	exec.Command("git", "-C", sourceRepo, "add", ".").Run()
	exec.Command("git", "-C", sourceRepo, "commit", "-m", "Initial commit").Run()

	// Create target directory
	targetPath := filepath.Join(tmpDir, "existing.git")
	os.MkdirAll(targetPath, 0755)
	os.WriteFile(filepath.Join(targetPath, "somefile.txt"), []byte("existing"), 0644)

	// Try to clone into existing directory
	err := CloneBare(sourceRepo, targetPath, false)
	if err == nil {
		t.Error("CloneBare() should fail when target directory already exists with content")
	}
}
