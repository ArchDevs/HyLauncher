package pwr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/internal/util"
)

func DownloadPWR(ctx context.Context, version string, prevVer int, targetVer int, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) (string, error) {
	cacheDir := filepath.Join(env.GetDefaultAppDir(), "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH

	fileName := fmt.Sprintf("%d.pwr", targetVer)

	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/%d/%s",
		osName, arch, version, prevVer, fileName)

	dest := filepath.Join(cacheDir, fileName)
	tempDest := dest + ".tmp"

	// Remove any incomplete temp file from previous session
	_ = os.Remove(tempDest)

	// Skip if already downloaded and complete
	if _, err := os.Stat(dest); err == nil {
		fmt.Println("PWR file already exists:", dest)
		if progressCallback != nil {
			progressCallback("download", 100, "PWR file cached", fileName, "", 0, 0)
		}
		return dest, nil
	}

	fmt.Println("Downloading PWR file:", url)

	// DownloadWithProgress will handle progress from 0-100
	if err := util.DownloadWithProgress(tempDest, url, "download", 1.0, progressCallback); err != nil {
		_ = os.Remove(tempDest)
		return "", err
	}

	// Move temp file to final destination atomically
	if err := os.Rename(tempDest, dest); err != nil {
		_ = os.Remove(tempDest)
		return "", err
	}

	fmt.Println("PWR downloaded to:", dest)
	return dest, nil
}
