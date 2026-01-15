package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	maxRetries      = 3
	retryDelay      = 2 * time.Second
	downloadTimeout = 30 * time.Minute
)

// DownloadWithProgress downloads a file with retries, partial download support and progress callback
func DownloadWithProgress(
	dest string,
	url string,
	stage string,
	progressWeight float64,
	callback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64),
) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			if callback != nil {
				callback(stage, 0, fmt.Sprintf("Retrying download (attempt %d/%d)...", attempt, maxRetries), "", "", 0, 0)
			}
			time.Sleep(retryDelay * time.Duration(attempt-1))
		}

		err := attemptDownload(dest, url, stage, progressWeight, callback)
		if err == nil {
			return nil
		}
		lastErr = err
		fmt.Printf("Download attempt %d failed: %v\n", attempt, err)
	}

	return fmt.Errorf("download failed after %d attempts: %w", maxRetries, lastErr)
}

func attemptDownload(
	dest string,
	url string,
	stage string,
	progressWeight float64,
	callback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64),
) error {
	// Create HTTP client
	client := &http.Client{Timeout: downloadTimeout}

	// Check if partial file exists
	var resumeFrom int64 = 0
	if stat, err := os.Stat(dest); err == nil {
		resumeFrom = stat.Size()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if resumeFrom > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	totalSize := resp.ContentLength
	if resumeFrom > 0 && resp.StatusCode == http.StatusPartialContent {
		totalSize += resumeFrom
	}

	// Open file
	flag := os.O_CREATE | os.O_WRONLY
	if resumeFrom > 0 && resp.StatusCode == http.StatusPartialContent {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
		resumeFrom = 0
	}

	out, err := os.OpenFile(dest, flag, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	// Download loop
	buffer := make([]byte, 32*1024)
	downloaded := resumeFrom
	startTime := time.Now()
	lastUpdate := startTime

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			now := time.Now()
			if callback != nil && now.Sub(lastUpdate) >= 200*time.Millisecond {
				elapsed := now.Sub(startTime).Seconds()
				speed := ""
				if elapsed > 0 {
					speed = formatSpeed(float64(downloaded-resumeFrom) / elapsed)
				}
				progress := float64(downloaded) / float64(totalSize) * 100 * progressWeight
				callback(stage, progress, "Downloading...", "", speed, downloaded, totalSize)
				lastUpdate = now
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	if callback != nil {
		callback(stage, progressWeight*100, "Download complete", "", "", downloaded, totalSize)
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
