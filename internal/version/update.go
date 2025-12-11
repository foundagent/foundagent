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

// CheckForUpdate queries GitHub releases API for the latest version
func CheckForUpdate(ctx context.Context) (updateAvailable bool, latestVersion string, downloadURL string, err error) {
	const releaseURL = "https://api.github.com/repos/foundagent/foundagent/releases/latest"

	// Create request with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
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
