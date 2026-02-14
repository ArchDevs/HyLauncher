package updater

import (
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const versionJSONAsset = "version.json"

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
			Portable Asset `json:"portable"`
			Setup    Asset `json:"setup"`
			Helper   Asset `json:"helper"`
		} `json:"amd64"`
	} `json:"windows"`
	Darwin struct {
		Arm64 struct {
			Launcher Asset `json:"launcher"`
			Helper   Asset `json:"helper"`
		} `json:"arm64"`
		Amd64 struct {
			Launcher Asset `json:"launcher"`
			Helper   Asset `json:"helper"`
		} `json:"amd64"`
	} `json:"darwin"`
}

type Asset struct {
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`
}

// Checks if there is any new launcher update, returns Asset: url to download, sha256 hash
func CheckUpdate(ctx context.Context, current string) (*Asset, string, error) {
	info, err := fetchUpdateInfo(ctx)
	if err != nil {
		return nil, "", err
	}

	currentClean := strings.TrimPrefix(strings.TrimSpace(current), "v")
	latestClean := strings.TrimPrefix(strings.TrimSpace(info.Version), "v")

	logger.Info("Checking for updates", "current", current, "latest", info.Version)

	if currentClean == latestClean {
		logger.Info("Already on latest version")
		return nil, "", nil
	}

	var asset *Asset
	switch runtime.GOOS {
	case "windows":
		asset = &info.Windows.Amd64.Portable
		logger.Info("Update available", "platform", "windows", "current", current, "latest", info.Version)
	case "darwin":
		if runtime.GOARCH == "arm64" {
			asset = &info.Darwin.Arm64.Launcher
			logger.Info("Update available", "platform", "darwin-arm64", "current", current, "latest", info.Version)
		} else {
			asset = &info.Darwin.Amd64.Launcher
			logger.Info("Update available", "platform", "darwin-amd64", "current", current, "latest", info.Version)
		}
	default:
		asset = &info.Linux.Amd64.Launcher
		logger.Info("Update available", "platform", "linux", "current", current, "latest", info.Version)
	}

	if asset.URL == "" {
		return nil, "", fmt.Errorf("no download URL found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return asset, info.Version, nil
}

// Get update-helper asset/info
func GetHelperAsset(ctx context.Context) (*Asset, error) {
	info, err := fetchUpdateInfo(ctx)
	if err != nil {
		return nil, err
	}

	var asset *Asset
	switch runtime.GOOS {
	case "windows":
		asset = &info.Windows.Amd64.Helper
	case "darwin":
		if runtime.GOARCH == "arm64" {
			asset = &info.Darwin.Arm64.Helper
		} else {
			asset = &info.Darwin.Amd64.Helper
		}
	default:
		asset = &info.Linux.Amd64.Helper
	}

	if asset.URL == "" {
		return nil, fmt.Errorf("no helper URL found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return asset, nil
}

// Creates temp version json, downloads actual version json, reads actual info, returns atest update info
func fetchUpdateInfo(ctx context.Context) (*UpdateInfo, error) {
	tempFile, err := fileutil.CreateTempFile("version-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// Download version.json
	if err := download.DownloadLatestReleaseAsset(ctx, versionJSONAsset, tempFile, progress.StageUpdate, nil, nil); err != nil {
		return nil, fmt.Errorf("failed to download version info: %w", err)
	}

	// Open version.json temp
	f, err := os.Open(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open version file: %w", err)
	}
	defer f.Close()

	// Read info from downloaded version.json
	var info UpdateInfo
	if err := json.NewDecoder(f).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to parse version info: %w", err)
	}

	return &info, nil
}
