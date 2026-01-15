package game

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/util"
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const githubRepoAPI = "https://api.github.com/repos/ArchDevs/HyLauncher/releases/latest"

func ApplyOnlineFixWindows(ctx context.Context, gameDir string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("online fix is only for Windows")
	}

	resp, err := http.Get(githubRepoAPI)
	if err != nil {
		return fmt.Errorf("failed to query GitHub API: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode GitHub release JSON: %w", err)
	}

	var zipURL string
	for _, a := range release.Assets {
		if a.Name == "online-fix.zip" {
			zipURL = a.BrowserDownloadURL
			break
		}
	}
	if zipURL == "" {
		return fmt.Errorf("online-fix.zip not found in latest release")
	}

	cacheDir := filepath.Join(gameDir, ".cache")
	_ = os.MkdirAll(cacheDir, 0755)

	zipPath := filepath.Join(cacheDir, "online_fix.zip.tmp")
	finalZipPath := filepath.Join(cacheDir, "online_fix.zip")

	if progressCallback != nil {
		progressCallback("online-fix", 0, "Downloading online-fix from GitHub...", "online-fix.zip", "", 0, 0)
	}

	if err := util.DownloadWithProgress(zipPath, zipURL, "online-fix", 0.6, progressCallback); err != nil {
		_ = os.Remove(zipPath)
		return err
	}
	if err := os.Rename(zipPath, finalZipPath); err != nil {
		return err
	}

	if progressCallback != nil {
		progressCallback("online-fix", 30, "Extracting archive...", "", "", 0, 0)
	}

	r, err := zip.OpenReader(finalZipPath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP: %w", err)
	}
	defer r.Close()

	tempDir := filepath.Join(cacheDir, "temp_extract")
	_ = os.RemoveAll(tempDir)
	_ = os.MkdirAll(tempDir, 0755)

	for _, f := range r.File {
		outPath := filepath.Join(tempDir, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(outPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		outFile, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return err
		}
		rc.Close()
		outFile.Close()
	}

	clientSrc := filepath.Join(tempDir, "Client", "HytaleClient.exe")
	clientDst := filepath.Join(gameDir, "Client", "HytaleClient.exe")
	_ = os.MkdirAll(filepath.Dir(clientDst), 0755)
	if err := util.СopyFile(clientSrc, clientDst); err != nil {
		return fmt.Errorf("failed to copy client: %w", err)
	}

	serverSrc := filepath.Join(tempDir, "Server")
	serverDst := filepath.Join(gameDir, "Server")
	if err := os.RemoveAll(serverDst); err != nil {
		return err
	}

	if err := util.СopyDir(serverSrc, serverDst); err != nil {
		return fmt.Errorf("failed to copy server folder: %w", err)
	}
	_ = os.RemoveAll(tempDir)
	_ = os.Remove(finalZipPath)

	if progressCallback != nil {
		progressCallback("online-fix", 100, "Online fix applied", "", "", 0, 0)
	}

	return nil
}

func EnsureServerAndClientFix(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	if runtime.GOOS != "windows" {
		return nil // Онлайн-фикс нужен только для Windows
	}

	baseDir := env.GetDefaultAppDir()
	gameLatestDir := filepath.Join(baseDir, "release", "package", "game", "latest")

	serverBat := filepath.Join(gameLatestDir, "Server", "start-server.bat")
	if _, err := os.Stat(serverBat); os.IsNotExist(err) {
		if progressCallback != nil {
			progressCallback("online-fix", 0, "Server missing, downloading online fix...", "", "", 0, 0)
		}

		if err := ApplyOnlineFixWindows(ctx, gameLatestDir, progressCallback); err != nil {
			return fmt.Errorf("failed to apply online fix: %w", err)
		}

		if progressCallback != nil {
			progressCallback("online-fix", 100, "Online fix applied", "", "", 0, 0)
		}
	}

	return nil
}
