package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GithubRelease represents a GitHub release
type GithubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// UpdateChecker handles checking for updates
type UpdateChecker struct {
	releaseURL string
}

// DefaultReleaseURL is the default URL for the GitHub releases API
const DefaultReleaseURL = "https://api.github.com/repos/foundagent/foundagent/releases/latest"

// ReleaseURL can be overridden for testing purposes (deprecated: use NewUpdateCheckerWithURL instead)
var ReleaseURL = DefaultReleaseURL

// NewUpdateChecker creates a new UpdateChecker with the default release URL
func NewUpdateChecker() *UpdateChecker {
	return &UpdateChecker{
		releaseURL: DefaultReleaseURL,
	}
}

// NewUpdateCheckerWithURL creates a new UpdateChecker with a custom release URL
func NewUpdateCheckerWithURL(url string) *UpdateChecker {
	return &UpdateChecker{
		releaseURL: url,
	}
}

// CheckForUpdate queries GitHub releases API for the latest version
func (u *UpdateChecker) CheckForUpdate(ctx context.Context) (updateAvailable bool, latestVersion string, downloadURL string, err error) {
	// Create request with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", u.releaseURL, nil)
	if err != nil {
		return false, "", "", err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", "", fmt.Errorf("failed to parse release info: %w", err)
	}

	latestVersion = strings.TrimPrefix(release.TagName, "v")
	downloadURL = release.HTMLURL

	// Compare versions (simple string comparison for now)
	currentVersion := strings.TrimPrefix(Version, "v")
	if currentVersion == "dev" || currentVersion == "unknown" {
		return false, latestVersion, downloadURL, nil
	}

	updateAvailable = latestVersion != currentVersion

	return updateAvailable, latestVersion, downloadURL, nil
}

// CheckForUpdate is a convenience function that uses the default UpdateChecker
// It respects the ReleaseURL variable for backward compatibility with tests
func CheckForUpdate(ctx context.Context) (updateAvailable bool, latestVersion string, downloadURL string, err error) {
	checker := NewUpdateCheckerWithURL(ReleaseURL)
	return checker.CheckForUpdate(ctx)
}
