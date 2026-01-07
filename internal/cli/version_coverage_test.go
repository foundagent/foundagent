package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/foundagent/foundagent/internal/version"
	"github.com/stretchr/testify/assert"
)

// TestCheckForUpdates_UpdateAvailable tests the update available path
func TestCheckForUpdates_UpdateAvailable(t *testing.T) {
	// Create mock server that returns a newer version
	mockRelease := struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}{
		TagName: "v99.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v99.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	// Save and restore original values
	origReleaseURL := version.ReleaseURL
	origVersion := version.Version
	defer func() {
		version.ReleaseURL = origReleaseURL
		version.Version = origVersion
	}()

	version.ReleaseURL = server.URL
	version.Version = "1.0.0"

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run checkForUpdates
	err := checkForUpdates()

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	_, _ = stdoutBuf.ReadFrom(rOut)
	_, _ = stderrBuf.ReadFrom(rErr)

	stdout := stdoutBuf.String()

	// Should succeed
	assert.NoError(t, err)

	// Should show update available message
	assert.Contains(t, stdout, "Update available")
	assert.Contains(t, stdout, "99.0.0")
	assert.Contains(t, stdout, "Download:")
}

// TestCheckForUpdates_UpToDate tests the up to date path
func TestCheckForUpdates_UpToDate(t *testing.T) {
	// Create mock server that returns the same version
	mockRelease := struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}{
		TagName: "v1.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v1.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	// Save and restore original values
	origReleaseURL := version.ReleaseURL
	origVersion := version.Version
	defer func() {
		version.ReleaseURL = origReleaseURL
		version.Version = origVersion
	}()

	version.ReleaseURL = server.URL
	version.Version = "1.0.0"

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run checkForUpdates
	err := checkForUpdates()

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	_, _ = stdoutBuf.ReadFrom(rOut)
	_, _ = stderrBuf.ReadFrom(rErr)

	stdout := stdoutBuf.String()

	// Should succeed
	assert.NoError(t, err)

	// Should show up to date message
	assert.Contains(t, stdout, "up to date")
}

// TestCheckForUpdates_NetworkError tests the error handling path
func TestCheckForUpdates_NetworkError(t *testing.T) {
	// Save and restore original values
	origReleaseURL := version.ReleaseURL
	defer func() {
		version.ReleaseURL = origReleaseURL
	}()

	// Point to an unreachable server
	version.ReleaseURL = "http://127.0.0.1:59998/releases/latest"

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run checkForUpdates
	err := checkForUpdates()

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	_, _ = stdoutBuf.ReadFrom(rOut)
	_, _ = stderrBuf.ReadFrom(rErr)

	stderr := stderrBuf.String()

	// Should succeed (error is a warning only)
	assert.NoError(t, err)

	// Should show warning message
	assert.Contains(t, stderr, "Warning")
	assert.Contains(t, stderr, "Failed to check for updates")
}

// TestCheckForUpdates_ServerError tests HTTP error handling
func TestCheckForUpdates_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Save and restore original values
	origReleaseURL := version.ReleaseURL
	defer func() {
		version.ReleaseURL = origReleaseURL
	}()

	version.ReleaseURL = server.URL

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run checkForUpdates
	err := checkForUpdates()

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	_, _ = stdoutBuf.ReadFrom(rOut)
	_, _ = stderrBuf.ReadFrom(rErr)

	stderr := stderrBuf.String()

	// Should succeed (error is a warning only)
	assert.NoError(t, err)

	// Should show warning message about HTTP error
	assert.True(t, strings.Contains(stderr, "Warning") || strings.Contains(stderr, "Failed"),
		"Expected warning in stderr: %q", stderr)
}

// TestCheckForUpdates_ShowsCurrentVersion tests that current version is always shown
func TestCheckForUpdates_ShowsCurrentVersion(t *testing.T) {
	// Create mock server
	mockRelease := struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}{
		TagName: "v1.0.0",
		HTMLURL: "https://github.com/foundagent/foundagent/releases/tag/v1.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	// Save and restore original values
	origReleaseURL := version.ReleaseURL
	defer func() {
		version.ReleaseURL = origReleaseURL
	}()

	version.ReleaseURL = server.URL

	// Capture output
	var stdoutBuf bytes.Buffer
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// Run checkForUpdates
	err := checkForUpdates()

	// Restore stdout
	wOut.Close()
	os.Stdout = oldStdout
	_, _ = stdoutBuf.ReadFrom(rOut)

	stdout := stdoutBuf.String()

	// Should succeed
	assert.NoError(t, err)

	// Should contain version string
	assert.Contains(t, stdout, version.String())
}
