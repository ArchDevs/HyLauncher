package pwr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
)

type ProgressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)

func DownloadPWR(ctx context.Context, version, fileName string, progressCallback ProgressCallback) (string, error) {
	cacheDir := filepath.Join(env.GetDefaultAppDir(), "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH

	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%s",
		osName, arch, version, fileName)

	dest := filepath.Join(cacheDir, fileName)

	// Skip if already downloaded
	if _, err := os.Stat(dest); err == nil {
		fmt.Println("PWR file already exists:", dest)
		if progressCallback != nil {
			progressCallback("game", 40, "PWR file cached", fileName, "", 0, 0)
		}
		return dest, nil
	}

	fmt.Println("Downloading PWR file:", url)
	if err := downloadFile(dest, url, progressCallback); err != nil {
		return "", err
	}

	fmt.Println("PWR downloaded to:", dest)
	return dest, nil
}

// downloadFile with progress reporting
func downloadFile(dest, url string, progressCallback ProgressCallback) error {
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

					if progressCallback != nil {
						// Map 0-100% download to 0-40% overall game progress
						overallProgress := percent * 0.4
						progressCallback("game", overallProgress, "Downloading game files...", filepath.Base(dest), speed, downloaded, total)
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
