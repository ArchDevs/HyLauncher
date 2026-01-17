package game

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

const (
	onlineFixAssetName     = "online-fix.zip"
	progressWeightDownload = 0.7
	progressWeightExtract  = 0.3
)

func ApplyOnlineFixWindows(ctx context.Context, gameDir string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("online fix is only for Windows")
	}

	cacheDir := filepath.Join(gameDir, ".cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	zipPath := filepath.Join(cacheDir, onlineFixAssetName)

	// Download from GitHub releases
	if progressCallback != nil {
		progressCallback("online-fix", 0, "Downloading online-fix from GitHub...", onlineFixAssetName, "", 0, 0)
	}

	if err := download.DownloadLatestReleaseAsset(ctx, onlineFixAssetName, zipPath, wrapProgressCallback(progressCallback, progressWeightDownload)); err != nil {
		_ = os.Remove(zipPath)
		return fmt.Errorf("failed to download online-fix: %w", err)
	}

	// Extract and apply the fix
	if progressCallback != nil {
		progressCallback("online-fix", 70, "Extracting archive...", "", "", 0, 0)
	}

	if err := extractAndApplyFix(zipPath, gameDir, cacheDir); err != nil {
		return err
	}

	// Cleanup
	_ = os.Remove(zipPath)

	if progressCallback != nil {
		progressCallback("online-fix", 100, "Online fix applied successfully", "", "", 0, 0)
	}

	return nil
}

func extractAndApplyFix(zipPath, gameDir, cacheDir string) error {
	tempDir := filepath.Join(cacheDir, "temp_extract")

	// Clean and create temp directory
	if err := os.RemoveAll(tempDir); err != nil {
		return fmt.Errorf("failed to clean temp directory: %w", err)
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract zip
	if err := util.ExtractZip(zipPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract ZIP: %w", err)
	}

	// Copy client executable
	clientSrc := filepath.Join(tempDir, "Client", "HytaleClient.exe")
	clientDst := filepath.Join(gameDir, "Client", "HytaleClient.exe")

	if err := os.MkdirAll(filepath.Dir(clientDst), 0755); err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}
	if err := util.CopyFile(clientSrc, clientDst); err != nil {
		return fmt.Errorf("failed to copy client executable: %w", err)
	}

	// Copy server folder
	serverSrc := filepath.Join(tempDir, "Server")
	serverDst := filepath.Join(gameDir, "Server")

	// Remove existing server folder
	if err := os.RemoveAll(serverDst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing server folder: %w", err)
	}
	if err := util.CopyDir(serverSrc, serverDst); err != nil {
		return fmt.Errorf("failed to copy server folder: %w", err)
	}

	return nil
}

func wrapProgressCallback(callback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64), weight float64) func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64) {
	if callback == nil {
		return nil
	}
	return func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64) {
		callback(stage, progress*weight, message, currentFile, speed, downloaded, total)
	}
}

func EnsureServerAndClientFix(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	baseDir := env.GetDefaultAppDir()
	gameLatestDir := filepath.Join(baseDir, "release", "package", "game", "latest")

	// Check if server exists
	serverBat := filepath.Join(gameLatestDir, "Server", "start-server.bat")
	if _, err := os.Stat(serverBat); err == nil {
		return nil
	}

	// Server missing, download and apply online fix
	if progressCallback != nil {
		progressCallback("online-fix", 0, "Server missing, downloading online fix...", "", "", 0, 0)
	}

	if err := ApplyOnlineFixWindows(ctx, gameLatestDir, progressCallback); err != nil {
		return fmt.Errorf("failed to apply online fix: %w", err)
	}

	return nil
}
