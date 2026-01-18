package patch

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/extract"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func InstallButler(ctx context.Context, reporter *progress.Reporter) (string, error) {
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
		reporter.Report(progress.StageButler, 100, "Butler already installed")
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
	reporter.Report(progress.StageButler, 0, "Downloading butler.zip...")

	// Create a scaler for the download portion (0-70%)
	scaler := progress.NewScaler(reporter, progress.StageButler, 0, 70)

	if err := download.DownloadWithReporter(tempZipPath, url, "butler.zip", reporter, progress.StageButler, scaler); err != nil {
		_ = os.Remove(tempZipPath)
		return "", err
	}

	// Move temp file to final destination
	if err := os.Rename(tempZipPath, zipPath); err != nil {
		_ = os.Remove(tempZipPath)
		return "", err
	}

	fmt.Println("Extracting Butler...")
	reporter.Report(progress.StageButler, 80, "Extracting butler.zip")

	if err := extract.ExtractZip(zipPath, toolsDir); err != nil {
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

	reporter.Report(progress.StageButler, 100, "Butler successfully installed!")

	return butlerPath, nil
}
