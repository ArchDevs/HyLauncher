package patch

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/internal/platform"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
)

func ApplyPWR(ctx context.Context, pwrFile string, reporter *progress.Reporter) error {

	gameLatest := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")
	stagingDir := filepath.Join(gameLatest, "staging-temp")
	_ = os.MkdirAll(stagingDir, 0755)

	butlerPath := filepath.Join(env.GetDefaultAppDir(), "tools", "butler", "butler")
	if runtime.GOOS == "windows" {
		butlerPath += ".exe"
	}

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		pwrFile,
		gameLatest,
	)

	platform.HideConsoleWindow(cmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	reporter.Report(progress.StagePatch, 60, "Applying game patch...")

	if err := cmd.Run(); err != nil {
		return err
	}

	// Retry rename on Windows if locked
	if runtime.GOOS == "windows" {
		for i := 0; i < 5; i++ {
			if err := os.Rename(stagingDir, gameLatest); err == nil {
				break
			}
			time.Sleep(2 * time.Second)
		}
	} else {
		_ = os.Rename(stagingDir, gameLatest)
	}

	reporter.Report(progress.StagePatch, 100, "Game patched!")

	return nil
}

func DownloadPWR(ctx context.Context, versionType string, prevVer int, targetVer int, reporter *progress.Reporter) (string, error) {

	cacheDir := filepath.Join(env.GetDefaultAppDir(), "cache")
	_ = os.MkdirAll(cacheDir, 0755)

	osName := runtime.GOOS
	arch := runtime.GOARCH

	fileName := fmt.Sprintf("%d.pwr", targetVer)
	dest := filepath.Join(cacheDir, fileName)
	tempDest := dest + ".tmp"

	_ = os.Remove(tempDest)

	if _, err := os.Stat(dest); err == nil {
		reporter.Report(progress.StagePWR, 100, "PWR file cached")
		return dest, nil
	}

	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/%d/%s",
		osName, arch, versionType, prevVer, fileName)

	reporter.Report(progress.StagePWR, 0, "Downloading PWR file...")

	// Create a scaler for the download portion (0-100%)
	scaler := progress.NewScaler(reporter, progress.StagePWR, 0, 100)

	if err := download.DownloadWithReporter(dest, url, fileName, reporter, progress.StagePWR, scaler); err != nil {
		_ = os.Remove(tempDest)
		return "", err
	}

	reporter.Report(progress.StagePWR, 100, "PWR file downloaded")

	return dest, nil
}
