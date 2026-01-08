package version

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckForUpdate_MockServer_Success tests successful API response
func TestCheckForUpdate_MockServer_Success(t *testing.T) {
	mockRelease := GithubRelease{
		TagName: "v2.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	// Save and restore original version
	origVersion := Version
	defer func() {
		Version = origVersion
	}()

	Version = "1.0.0"

	ctx := context.Background()
	checker := NewUpdateCheckerWithURL(server.URL)
	updateAvailable, latestVersion, downloadURL, err := checker.CheckForUpdate(ctx)

	require.NoError(t, err)
	assert.True(t, updateAvailable)
	assert.Equal(t, "2.0.0", latestVersion)
	assert.Contains(t, downloadURL, "v2.0.0")
}

// TestCheckForUpdate_MockServer_SameVersion tests when current version matches latest
func TestCheckForUpdate_MockServer_SameVersion(t *testing.T) {
	mockRelease := GithubRelease{
		TagName: "v1.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v1.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	origVersion := Version
	origReleaseURL := ReleaseURL
	defer func() {
		Version = origVersion
		ReleaseURL = origReleaseURL
	}()

	Version = "1.0.0"
	ReleaseURL = server.URL

	ctx := context.Background()
	updateAvailable, latestVersion, _, err := CheckForUpdate(ctx)

	require.NoError(t, err)
	assert.False(t, updateAvailable)
	assert.Equal(t, "1.0.0", latestVersion)
}

// TestCheckForUpdate_MockServer_DevVersion tests dev version handling
func TestCheckForUpdate_MockServer_DevVersion(t *testing.T) {
	mockRelease := GithubRelease{
		TagName: "v2.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	origVersion := Version
	origReleaseURL := ReleaseURL
	defer func() {
		Version = origVersion
		ReleaseURL = origReleaseURL
	}()

	Version = "dev"
	ReleaseURL = server.URL

	ctx := context.Background()
	updateAvailable, latestVersion, _, err := CheckForUpdate(ctx)

	require.NoError(t, err)
	assert.False(t, updateAvailable) // dev should never show update
	assert.Equal(t, "2.0.0", latestVersion)
}

// TestCheckForUpdate_MockServer_UnknownVersion tests unknown version handling
func TestCheckForUpdate_MockServer_UnknownVersion(t *testing.T) {
	mockRelease := GithubRelease{
		TagName: "v2.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	origVersion := Version
	origReleaseURL := ReleaseURL
	defer func() {
		Version = origVersion
		ReleaseURL = origReleaseURL
	}()

	Version = "unknown"
	ReleaseURL = server.URL

	ctx := context.Background()
	updateAvailable, _, _, err := CheckForUpdate(ctx)

	require.NoError(t, err)
	assert.False(t, updateAvailable) // unknown should never show update
}

// TestCheckForUpdate_MockServer_HTTPError tests non-200 response
func TestCheckForUpdate_MockServer_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = server.URL

	ctx := context.Background()
	_, _, _, err := CheckForUpdate(ctx)

	// Should error for non-200
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
}

// TestCheckForUpdate_MockServer_InvalidJSON tests malformed JSON response
func TestCheckForUpdate_MockServer_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json {{{"))
	}))
	defer server.Close()

	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = server.URL

	ctx := context.Background()
	_, _, _, err := CheckForUpdate(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

// TestCheckForUpdate_MockServer_ServerError tests server error response
func TestCheckForUpdate_MockServer_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = server.URL

	ctx := context.Background()
	_, _, _, err := CheckForUpdate(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

// TestCheckForUpdate_MockServer_EmptyResponse tests empty response body
func TestCheckForUpdate_MockServer_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body
	}))
	defer server.Close()

	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = server.URL

	ctx := context.Background()
	_, _, _, err := CheckForUpdate(ctx)

	assert.Error(t, err) // JSON decode should fail on empty body
}

// TestCheckForUpdate_MockServer_TagWithoutV tests tag without v prefix
func TestCheckForUpdate_MockServer_TagWithoutV(t *testing.T) {
	mockRelease := GithubRelease{
		TagName: "2.0.0", // No 'v' prefix
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	origVersion := Version
	origReleaseURL := ReleaseURL
	defer func() {
		Version = origVersion
		ReleaseURL = origReleaseURL
	}()

	Version = "1.0.0"
	ReleaseURL = server.URL

	ctx := context.Background()
	updateAvailable, latestVersion, _, err := CheckForUpdate(ctx)

	require.NoError(t, err)
	assert.True(t, updateAvailable)
	assert.Equal(t, "2.0.0", latestVersion)
}

// TestCheckForUpdate_MockServer_ContextTimeout tests timeout handling
func TestCheckForUpdate_MockServer_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, _, _, err := CheckForUpdate(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestCheckForUpdate_MockServer_ContextCancelled tests cancelled context
func TestCheckForUpdate_MockServer_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, _, err := CheckForUpdate(ctx)

	assert.Error(t, err)
}

// TestCheckForUpdate_InvalidURL tests invalid URL handling
func TestCheckForUpdate_InvalidURL(t *testing.T) {
	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = "://invalid-url"

	ctx := context.Background()
	_, _, _, err := CheckForUpdate(ctx)

	assert.Error(t, err)
}

// TestCheckForUpdate_UnreachableServer tests connection refused
func TestCheckForUpdate_UnreachableServer(t *testing.T) {
	origReleaseURL := ReleaseURL
	defer func() { ReleaseURL = origReleaseURL }()
	ReleaseURL = "http://127.0.0.1:59999/releases/latest"

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, _, _, err := CheckForUpdate(ctx)

	assert.Error(t, err)
}

// TestGithubRelease_FullJSONFields tests all JSON fields
func TestGithubRelease_FullJSONFields(t *testing.T) {
	// Test with additional fields that might be in GitHub API response
	jsonData := `{
		"tag_name": "v3.0.0",
		"html_url": "https://github.com/foundagent/foundagent/releases/tag/v3.0.0",
		"name": "Release v3.0.0",
		"draft": false,
		"prerelease": false
	}`

	var release GithubRelease
	err := json.Unmarshal([]byte(jsonData), &release)
	require.NoError(t, err)

	assert.Equal(t, "v3.0.0", release.TagName)
	assert.Equal(t, "https://github.com/foundagent/foundagent/releases/tag/v3.0.0", release.HTMLURL)
}

// TestVersionPrefixHandling tests various version prefix scenarios
func TestVersionPrefixHandling(t *testing.T) {
	tests := []struct {
		name      string
		tagName   string
		expectVer string
	}{
		{"with v prefix", "v1.2.3", "1.2.3"},
		{"without v prefix", "1.2.3", "1.2.3"},
		{"with V prefix", "V1.2.3", "V1.2.3"}, // Only lowercase v is stripped
		{"alpha version", "v1.0.0-alpha", "1.0.0-alpha"},
		{"beta version", "v2.0.0-beta.1", "2.0.0-beta.1"},
		{"rc version", "v3.0.0-rc1", "3.0.0-rc1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.TrimPrefix(tt.tagName, "v")
			assert.Equal(t, tt.expectVer, result)
		})
	}
}

// TestCheckForUpdate_MockServer_VersionWithVPrefix tests version comparison with v prefix
func TestCheckForUpdate_MockServer_VersionWithVPrefix(t *testing.T) {
	mockRelease := GithubRelease{
		TagName: "v2.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	origVersion := Version
	origReleaseURL := ReleaseURL
	defer func() {
		Version = origVersion
		ReleaseURL = origReleaseURL
	}()

	// Version with v prefix should also work
	Version = "v1.0.0"
	ReleaseURL = server.URL

	ctx := context.Background()
	updateAvailable, latestVersion, _, err := CheckForUpdate(ctx)

	require.NoError(t, err)
	assert.True(t, updateAvailable)
	assert.Equal(t, "2.0.0", latestVersion)
}

// TestReleaseURL_DefaultValue tests that ReleaseURL has correct default value
func TestReleaseURL_DefaultValue(t *testing.T) {
	assert.Contains(t, ReleaseURL, "github.com")
	assert.Contains(t, ReleaseURL, "releases/latest")
}
