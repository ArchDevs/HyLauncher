package butler

import (
	"HyLauncher/internal/env"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func InstallButler(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) (string, error) {
	toolsDir := filepath.Join(env.GetDefaultAppDir(), "tools", "butler")
	zipPath := filepath.Join(toolsDir, "butler.zip")
	var butlerPath string
	if runtime.GOOS == "windows" {
		butlerPath = filepath.Join(toolsDir, "butler.exe")
	} else {
		butlerPath = filepath.Join(toolsDir, "butler")
	}

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

	if err := downloadFile(zipPath, url, progressCallback); err != nil {
		return "", err
	}

	fmt.Println("Extracting Butler...")
	if progressCallback != nil {
		progressCallback("butler", 80, "Extracting Butler...", "butler.zip", "", 0, 0)
	}

	if err := unzip(zipPath, toolsDir); err != nil {
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

// Download butler with progress
func downloadFile(dest, url string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)
	start := time.Now()
	lastUpdate := time.Now()

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := out.Write(buf[:n]); wErr != nil {
				return wErr
			}
			downloaded += int64(n)

			// Update progress every 200ms
			if time.Since(lastUpdate) > 200*time.Millisecond {
				if total > 0 {
					percent := float64(downloaded) / float64(total) * 100
					elapsed := time.Since(start).Seconds()
					speed := ""
					if elapsed > 0 {
						mbps := float64(downloaded) / 1024 / 1024 / elapsed
						speed = fmt.Sprintf("%.2f MB/s", mbps)
					}

					// Map to 0-70% of butler stage
					butlerProgress := percent * 0.7

					if progressCallback != nil {
						progressCallback("butler", butlerProgress, "Downloading Butler...", filepath.Base(dest), speed, downloaded, total)
					}

					fmt.Printf("\r%.1f%% downloaded (%.2f MB/s)", percent, float64(downloaded)/1024/1024/elapsed)
				}
				lastUpdate = time.Now()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	fmt.Println()
	return nil
}
