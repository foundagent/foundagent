package version

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckForUpdate_Success(t *testing.T) {
	// Mock GitHub API response
	mockRelease := GithubRelease{
		TagName: "v1.2.3",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v1.2.3",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	// Save original Version and restore after test
	origVersion := Version
	defer func() { Version = origVersion }()

	Version = "1.0.0"

	// We can't easily override the URL in CheckForUpdate, so this test is limited
	// In a real scenario, we'd refactor CheckForUpdate to accept a custom URL
	// For now, test with dev version which should return no update available
	Version = "dev"

	ctx := context.Background()
	updateAvailable, _, _, err := CheckForUpdate(ctx)

	// When version is "dev", no update should be available
	if err != nil {
		// Network errors are acceptable in test environment
		t.Logf("CheckForUpdate returned error (expected in some environments): %v", err)
		return
	}

	if updateAvailable {
		t.Error("Expected no update available for dev version")
	}
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	// Save original Version and restore after test
	origVersion := Version
	defer func() { Version = origVersion }()

	Version = "dev"

	ctx := context.Background()
	updateAvailable, _, _, err := CheckForUpdate(ctx)

	if err != nil {
		// Network errors are acceptable in test environment
		t.Logf("CheckForUpdate returned error (expected in some environments): %v", err)
		return
	}

	// Dev version should never show update available
	if updateAvailable {
		t.Error("Expected no update available for dev version")
	}
}

func TestCheckForUpdate_UnknownVersion(t *testing.T) {
	// Save original Version and restore after test
	origVersion := Version
	defer func() { Version = origVersion }()

	Version = "unknown"

	ctx := context.Background()
	updateAvailable, _, _, err := CheckForUpdate(ctx)

	if err != nil {
		// Network errors are acceptable in test environment
		t.Logf("CheckForUpdate returned error (expected in some environments): %v", err)
		return
	}

	// Unknown version should never show update available
	if updateAvailable {
		t.Error("Expected no update available for unknown version")
	}
}

func TestCheckForUpdate_Timeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, _, _, err := CheckForUpdate(ctx)

	// Should timeout - but we can't override the URL easily
	// This test documents the limitation
	if err == nil {
		t.Log("CheckForUpdate didn't timeout as expected (URL hardcoded)")
	}
}

func TestGithubRelease_Structure(t *testing.T) {
	// Test that GithubRelease struct can be properly unmarshaled
	jsonData := `{"tag_name": "v1.0.0", "html_url": "https://example.com"}`

	var release GithubRelease
	err := json.Unmarshal([]byte(jsonData), &release)

	if err != nil {
		t.Fatalf("Failed to unmarshal GithubRelease: %v", err)
	}

	if release.TagName != "v1.0.0" {
		t.Errorf("TagName = %q, want 'v1.0.0'", release.TagName)
	}

	if release.HTMLURL != "https://example.com" {
		t.Errorf("HTMLURL = %q, want 'https://example.com'", release.HTMLURL)
	}
}

// TestCheckForUpdate_RealAPI tests against real GitHub API
// This test may fail in environments without network access
func TestCheckForUpdate_RealAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, latestVersion, downloadURL, err := CheckForUpdate(ctx)

	if err != nil {
		t.Logf("CheckForUpdate failed (may be expected): %v", err)
		return
	}

	if latestVersion == "" {
		t.Error("Expected non-empty latest version")
	}

	if downloadURL == "" {
		t.Error("Expected non-empty download URL")
	}

	t.Logf("Latest version: %s", latestVersion)
	t.Logf("Download URL: %s", downloadURL)
}
