package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func Download(url string, progress func(int64, int64)) (string, error) {
	fmt.Printf("Starting download from: %s\n", url)

	client := &http.Client{
		Timeout: 30 * time.Minute, // Long timeout for large files
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	tmp := filepath.Join(os.TempDir(), "hylauncher-update.tmp")
	fmt.Printf("Downloading to: %s\n", tmp)

	// Remove any existing temp file
	_ = os.Remove(tmp)

	out, err := os.Create(tmp)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	total := resp.ContentLength
	var downloaded int64

	fmt.Printf("Total size: %d bytes (%.2f MB)\n", total, float64(total)/1024/1024)

	buf := make([]byte, 32*1024)
	lastUpdate := time.Now()

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("failed to write to file: %w", writeErr)
			}
			downloaded += int64(n)

			// Update progress every 100ms
			if progress != nil && time.Since(lastUpdate) >= 100*time.Millisecond {
				progress(downloaded, total)
				lastUpdate = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("download error: %w", err)
		}
	}

	// Final progress update
	if progress != nil {
		progress(downloaded, total)
	}

	fmt.Printf("Download complete: %d bytes\n", downloaded)
	return tmp, nil
}
