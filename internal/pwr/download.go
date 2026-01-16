package pwr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/internal/util/download"
)

func DownloadPWR(ctx context.Context, versionType string, prevVer int, targetVer int,
	progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) (string, error) {

	cacheDir := filepath.Join(env.GetDefaultAppDir(), "cache")
	_ = os.MkdirAll(cacheDir, 0755)

	osName := runtime.GOOS
	arch := runtime.GOARCH

	fileName := fmt.Sprintf("%d.pwr", targetVer)
	dest := filepath.Join(cacheDir, fileName)
	tempDest := dest + ".tmp"

	_ = os.Remove(tempDest)

	if _, err := os.Stat(dest); err == nil {
		if progressCallback != nil {
			progressCallback("game", 40, "PWR file cached", fileName, "", 0, 0)
		}
		return dest, nil
	}

	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/%d/%s",
		osName, arch, versionType, prevVer, fileName)

	if err := download.DownloadWithProgress(tempDest, url, "game", 0.4, progressCallback); err != nil {
		_ = os.Remove(tempDest)
		return "", err
	}

	if err := os.Rename(tempDest, dest); err != nil {
		_ = os.Remove(tempDest)
		return "", err
	}

	return dest, nil
}
