package butler

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/util"
	"HyLauncher/internal/util/download"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func InstallButler(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) (string, error) {
	toolsDir := filepath.Join(env.GetDefaultAppDir(), "tools", "butler")
	zipPath := filepath.Join(toolsDir, "butler.zip")
	tempZipPath := zipPath + ".tmp"

	if _, err := os.Stat(toolsDir); os.IsNotExist(err) {
		os.MkdirAll(toolsDir, 0755)
	}

	var butlerPath string
	if runtime.GOOS == "windows" {
		butlerPath = filepath.Join(toolsDir, "butler.exe")
	} else {
		butlerPath = filepath.Join(toolsDir, "butler")
	}

	// Remove any incomplete temp file from previous session
	_ = os.Remove(tempZipPath)

	// If binary already exists, skip
	if _, err := os.Stat(butlerPath); err == nil {
		if progressCallback != nil {
			progressCallback("butler", 100, "Butler already installed", "", "", 0, 0)
		}
		return butlerPath, nil
	}

	// Determine download URL
	var url string
	switch runtime.GOOS {
	case "windows":
		url = "https://broth.itch.zone/butler/windows-amd64/LATEST/archive/default"
	case "darwin":
		url = "https://broth.itch.zone/butler/darwin-amd64/LATEST/archive/default"
	case "linux":
		url = "https://broth.itch.zone/butler/linux-amd64/LATEST/archive/default"
	default:
		return "", fmt.Errorf("unsupported OS")
	}

	fmt.Println("Downloading Butler...")
	if progressCallback != nil {
		progressCallback("butler", 0, "Downloading Butler...", "butler.zip", "", 0, 0)
	}

	if err := download.DownloadWithProgress(tempZipPath, url, "butler", 0.7, progressCallback); err != nil {
		_ = os.Remove(tempZipPath)
		return "", err
	}

	// Move temp file to final destination
	if err := os.Rename(tempZipPath, zipPath); err != nil {
		_ = os.Remove(tempZipPath)
		return "", err
	}

	fmt.Println("Extracting Butler...")
	if progressCallback != nil {
		progressCallback("butler", 80, "Extracting Butler...", "butler.zip", "", 0, 0)
	}

	if err := util.ExtractZip(zipPath, toolsDir); err != nil {
		return "", err
	}

	// Make executable on unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(butlerPath, 0755); err != nil {
			return "", err
		}
	}

	// Cleanup zip
	_ = os.Remove(zipPath)

	if progressCallback != nil {
		progressCallback("butler", 100, "Butler installed", "", "", 0, 0)
	}

	return butlerPath, nil
}
