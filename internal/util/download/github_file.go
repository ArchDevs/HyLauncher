package download

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultRepoOwner = "ArchDevs"
	defaultRepoName  = "HyLauncher"
)

type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type GitHubRelease struct {
	TagName string               `json:"tag_name"`
	Name    string               `json:"name"`
	Assets  []GitHubReleaseAsset `json:"assets"`
}

func DownloadLatestReleaseAsset(
	ctx context.Context,
	assetName string,
	destPath string,
	progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64),
) error {
	// Construct GitHub API URL for our repository
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", defaultRepoOwner, defaultRepoName)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for GitHub API
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "HyLauncher")

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Decode the release information
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode GitHub release JSON: %w", err)
	}

	// Find the requested asset
	var downloadURL string
	var assetSize int64
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			assetSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("asset '%s' not found in latest release (tag: %s)", assetName, release.TagName)
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Download the file
	if progressCallback != nil {
		progressCallback("download", 0, fmt.Sprintf("Downloading %s from release %s...", assetName, release.TagName), assetName, "", 0, assetSize)
	}

	if err := DownloadWithProgress(destPath, downloadURL, "download", 1.0, progressCallback); err != nil {
		// Clean up partial download on error
		_ = os.Remove(destPath)
		return fmt.Errorf("failed to download %s: %w", assetName, err)
	}

	if progressCallback != nil {
		progressCallback("download", 100, fmt.Sprintf("Downloaded %s successfully", assetName), assetName, "", assetSize, assetSize)
	}

	return nil
}

func GetLatestReleaseInfo(ctx context.Context) (*GitHubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", defaultRepoOwner, defaultRepoName)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "HyLauncher")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release JSON: %w", err)
	}

	return &release, nil
}

func ListLatestReleaseAssets(ctx context.Context) ([]GitHubReleaseAsset, error) {
	release, err := GetLatestReleaseInfo(ctx)
	if err != nil {
		return nil, err
	}
	return release.Assets, nil
}
