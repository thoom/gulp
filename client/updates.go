package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release API response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	UpdateURL      string
	HasUpdate      bool
}

// httpClient defines the interface for an HTTP client, allowing for mocks.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CheckForUpdates checks if there's a newer version available on GitHub
func CheckForUpdates(currentVersion string, timeout time.Duration) (*UpdateInfo, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	return checkForUpdatesWithClient(currentVersion, client)
}

// checkForUpdatesWithClient is the testable implementation of CheckForUpdates.
func checkForUpdatesWithClient(currentVersion string, client httpClient) (*UpdateInfo, error) {
	// Skip update check for development/snapshot versions
	if strings.Contains(currentVersion, "SNAPSHOT") || currentVersion == "" {
		return &UpdateInfo{
			CurrentVersion: currentVersion,
			LatestVersion:  "unknown",
			HasUpdate:      false,
		}, nil
	}

	// GitHub API endpoint for latest release
	apiURL := "https://api.github.com/repos/thoom/gulp/releases/latest"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header as required by GitHub API
	req.Header.Set("User-Agent", CreateUserAgent())
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch release info: HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Clean up version tags (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersionClean := strings.TrimPrefix(currentVersion, "v")

	hasUpdate := isNewerVersion(latestVersion, currentVersionClean)

	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		UpdateURL:      release.HTMLURL,
		HasUpdate:      hasUpdate,
	}, nil
}

// isNewerVersion compares two version strings and returns true if 'latest' is newer than 'current'
// This handles basic semantic versioning (X.Y.Z format)
func isNewerVersion(latest, current string) bool {
	// If versions are identical, no update needed
	if latest == current {
		return false
	}

	// If either version is empty, use simple comparison
	if latest == "" || current == "" {
		return latest > current
	}

	// Parse semantic versions
	latestParts := parseVersion(latest)
	currentParts := parseVersion(current)

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		} else if latestParts[i] < currentParts[i] {
			return false
		}
	}

	// All parts are equal, no update needed
	return false
}

// parseVersion parses a version string into [major, minor, patch] integers
// If parsing fails, returns [0, 0, 0]
func parseVersion(version string) [3]int {
	parts := strings.Split(version, ".")
	result := [3]int{0, 0, 0}

	for i := 0; i < len(parts) && i < 3; i++ {
		if num, err := strconv.Atoi(parts[i]); err == nil {
			result[i] = num
		}
	}

	return result
}
