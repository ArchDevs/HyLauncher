package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type UpdateInfo struct {
	Version string `json:"version"`
	Linux   struct {
		Amd64 struct {
			Launcher Asset `json:"launcher"`
			Helper   Asset `json:"helper"`
		} `json:"amd64"`
	} `json:"linux"`
	Windows struct {
		Amd64 struct {
			Launcher Asset `json:"launcher"`
			Helper   Asset `json:"helper"`
		} `json:"amd64"`
	} `json:"windows"`
}

type Asset struct {
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`
}

func CheckUpdate(current string) (*Asset, string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(
		"https://github.com/ArchDevs/HyLauncher/releases/latest/download/version.json",
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("update check failed with status: %d", resp.StatusCode)
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, "", fmt.Errorf("failed to parse update info: %w", err)
	}

	fmt.Printf("Current version: %s, Latest version: %s\n", current, info.Version)

	// Clean version strings (remove 'v' prefix if present)
	currentClean := strings.TrimPrefix(strings.TrimSpace(current), "v")
	latestClean := strings.TrimPrefix(strings.TrimSpace(info.Version), "v")

	// If versions match, no update needed
	if currentClean == latestClean {
		fmt.Println("Already on latest version")
		return nil, "", nil
	}

	// Get the appropriate asset for the platform
	var asset *Asset
	if runtime.GOOS == "windows" {
		asset = &info.Windows.Amd64.Launcher
		fmt.Printf("Update available for Windows: %s -> %s\n", current, info.Version)
	} else {
		asset = &info.Linux.Amd64.Launcher
		fmt.Printf("Update available for Linux: %s -> %s\n", current, info.Version)
	}

	// Validate asset has URL
	if asset.URL == "" {
		return nil, "", fmt.Errorf("no download URL found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return asset, info.Version, nil
}
