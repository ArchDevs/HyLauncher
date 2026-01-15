package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	downloadTimeout = 30 * time.Minute
	progressTick    = 150 * time.Millisecond
)

func DownloadWithProgress(
	dest string,
	url string,
	stage string,
	progressWeight float64,
	callback func(
		stage string,
		progress float64,
		message string,
		currentFile string,
		speed string,
		downloaded, total int64,
	),
) error {
	// Create HTTP client
	client := &http.Client{
		Timeout: downloadTimeout,
	}

	// Build request
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// REQUIRED for Windows/CDN stability
	req.Header.Set("User-Agent", "HyLauncher/1.0")
	req.Header.Set("Accept-Encoding", "identity")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status: %s", resp.Status)
	}

	// Prepare output file (always overwrite)
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer out.Close()

	total := resp.ContentLength
	unknownSize := total <= 0

	buffer := make([]byte, 32*1024)
	var downloaded int64
	start := time.Now()
	lastUpdate := time.Now()

	for {
		n, err := resp.Body.Read(buffer)

		if n > 0 {
			if _, werr := out.Write(buffer[:n]); werr != nil {
				return fmt.Errorf("write failed: %w", werr)
			}
			downloaded += int64(n)
		}

		now := time.Now()
		if callback != nil && now.Sub(lastUpdate) >= progressTick {
			elapsed := now.Sub(start).Seconds()
			speed := ""
			if elapsed > 0 {
				speed = formatSpeed(float64(downloaded) / elapsed)
			}

			progress := 0.0
			if !unknownSize {
				progress = (float64(downloaded) / float64(total)) * 100 * progressWeight
			}

			callback(
				stage,
				progress,
				"Downloading...",
				"",
				speed,
				downloaded,
				total,
			)

			lastUpdate = now
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
	}

	// Final callback
	if callback != nil {
		callback(
			stage,
			progressWeight*100,
			"Download complete",
			"",
			"",
			downloaded,
			total,
		)
	}

	if runtime.GOOS == "windows" {
		_ = out.Sync()
	}

	return nil
}

func formatSpeed(bytesPerSec float64) string {
	const unit = 1024

	if bytesPerSec < unit {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	}

	div, exp := float64(unit), 0
	for n := bytesPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB/s", bytesPerSec/div, "KMGTPE"[exp])
}
