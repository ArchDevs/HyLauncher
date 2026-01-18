package patch

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/internal/platform"
	"HyLauncher/pkg/download"
)

func ApplyPWR(ctx context.Context, channel string, pwrFile string, installDirName string,
	progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {

	gameInstallDir := filepath.Join(env.GetDefaultAppDir(), channel, "package", "game", installDirName)
	stagingDir := filepath.Join(env.GetDefaultAppDir(), channel, "package", "game", "staging-temp")

	// Create parent directory
	_ = os.MkdirAll(filepath.Dir(gameInstallDir), 0755)

	// Create target directory explicitly (butler requires it to exist)
	if err := os.MkdirAll(gameInstallDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Clean up any previous staging directory
	_ = os.RemoveAll(stagingDir)
	_ = os.MkdirAll(stagingDir, 0755)

	butlerPath := filepath.Join(env.GetDefaultAppDir(), "tools", "butler", "butler")
	if runtime.GOOS == "windows" {
		butlerPath += ".exe"
	}

	// Verify butler exists
	if _, err := os.Stat(butlerPath); err != nil {
		return fmt.Errorf("butler tool not found at %s: %w", butlerPath, err)
	}

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		pwrFile,
		gameInstallDir,
	)

	platform.HideConsoleWindow(cmd)

	// Open log file for this operation
	logDir := filepath.Join(env.GetDefaultAppDir(), "logs")
	_ = os.MkdirAll(logDir, 0755)
	logFile, err := os.Create(filepath.Join(logDir, "butler_apply.log"))
	if err == nil {
		defer logFile.Close()
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		fmt.Fprintf(logFile, "Starting butler apply for %s to %s\n", pwrFile, gameInstallDir)
	} else {
		// Fallback if log file fails
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if progressCallback != nil {
		progressCallback("game", 60, "Applying game patch (this may take a while)...", "", "", 0, 0)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("butler apply failed (check logs/butler_apply.log): %w", err)
	}

	_ = os.RemoveAll(stagingDir)

	if progressCallback != nil {
		progressCallback("game", 100, "Game installed successfully", "", "", 0, 0)
	}

	return nil
}

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
