package game

import (
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/extract"
	"HyLauncher/pkg/fileutil"
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
	if err := extract.ExtractZip(zipPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract ZIP: %w", err)
	}

	// Copy client executable
	clientSrc := filepath.Join(tempDir, "Client", "HytaleClient.exe")
	clientDst := filepath.Join(gameDir, "Client", "HytaleClient.exe")

	if err := os.MkdirAll(filepath.Dir(clientDst), 0755); err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}
	if err := fileutil.CopyFile(clientSrc, clientDst); err != nil {
		return fmt.Errorf("failed to copy client executable: %w", err)
	}

	// Copy ONLY specific server files (not the whole folder)
	serverDir := filepath.Join(gameDir, "Server")
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %w", err)
	}

	// Copy HytaleServer.jar (replace existing)
	serverJarSrc := filepath.Join(tempDir, "Server", "HytaleServer.jar")
	serverJarDst := filepath.Join(serverDir, "HytaleServer.jar")
	if err := fileutil.CopyFile(serverJarSrc, serverJarDst); err != nil {
		return fmt.Errorf("failed to copy HytaleServer.jar: %w", err)
	}

	// Copy start-server.bat (add new file)
	startBatSrc := filepath.Join(tempDir, "Server", "start-server.bat")
	startBatDst := filepath.Join(serverDir, "start-server.bat")
	if err := fileutil.CopyFile(startBatSrc, startBatDst); err != nil {
		return fmt.Errorf("failed to copy start-server.bat: %w", err)
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

func EnsureServerAndClientFix(ctx context.Context, gameDir string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Check if server exists
	serverBat := filepath.Join(gameDir, "Server", "start-server.bat")
	if _, err := os.Stat(serverBat); err == nil {
		return nil
	}

	// Server missing, download and apply online fix
	if progressCallback != nil {
		progressCallback("online-fix", 0, "Server missing, downloading online fix...", "", "", 0, 0)
	}

	if err := ApplyOnlineFixWindows(ctx, gameDir, progressCallback); err != nil {
		return fmt.Errorf("failed to apply online fix: %w", err)
	}

	return nil
}
