package download

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"HyLauncher/internal/progress"
)

const (
	maxRetries     = 5
	baseRetryDelay = 3 * time.Second
	downloadLimit  = 45 * time.Minute
)

// DownloadWithReporter is a reliable, tolerant downloader
func DownloadWithReporter(
	dest string,
	url string,
	fileName string,
	reporter *progress.Reporter,
	stage progress.Stage,
	scaler *progress.Scaler,
) error {

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			delay := baseRetryDelay * time.Duration(1<<(attempt-2))
			if delay > 60*time.Second {
				delay = 60 * time.Second
			}

			msg := fmt.Sprintf("Retrying download (%d/%d)...", attempt, maxRetries)
			if scaler != nil {
				scaler.Report(stage, 0, msg)
			} else if reporter != nil {
				reporter.Report(stage, 0, msg)
			}

			time.Sleep(delay)
		}

		err := attemptDownload(dest, url, fileName, reporter, stage, scaler)
		if err == nil {
			return nil
		}

		lastErr = err
		fmt.Println("Download failed:", err)

		// Windows AV needs a little time
		if runtime.GOOS == "windows" {
			time.Sleep(2 * time.Second)
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", maxRetries, lastErr)
}

func attemptDownload(
	dest string,
	url string,
	fileName string,
	reporter *progress.Reporter,
	stage progress.Stage,
	scaler *progress.Scaler,
) error {

	client := createSafeClient()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	partPath := dest + ".part"

	var resumeFrom int64
	if st, err := os.Stat(partPath); err == nil {
		resumeFrom = st.Size()
	}

	ctx, cancel := context.WithTimeout(context.Background(), downloadLimit)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	// Diagnose
	fmt.Printf(
		"Download debug: status=%d resume=%v length=%d accept-ranges=%q\n",
		resp.StatusCode,
		resumeFrom > 0,
		resp.ContentLength,
		resp.Header.Get("Accept-Ranges"),
	)

	// Resume safety checks
	if resumeFrom > 0 {
		if resp.StatusCode != http.StatusPartialContent ||
			resp.Header.Get("Accept-Ranges") != "bytes" {

			_ = os.Remove(partPath)
			resumeFrom = 0
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("bad HTTP status: %s", resp.Status)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resumeFrom > 0 && resp.StatusCode == http.StatusPartialContent {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	out, err := os.OpenFile(partPath, flags, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	if resumeFrom > 0 && total > 0 {
		total += resumeFrom
	}

	buf := make([]byte, 64*1024)
	downloaded := resumeFrom
	lastUpdate := time.Now()
	lastBytes := downloaded

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}

			downloaded += int64(n)
			now := time.Now()

			if now.Sub(lastUpdate) >= 200*time.Millisecond {
				speed := float64(downloaded-lastBytes) / now.Sub(lastUpdate).Seconds()
				progressPct := 0.0
				if total > 0 {
					progressPct = float64(downloaded) / float64(total) * 100
				}

				if scaler != nil {
					scaler.ReportDownload(stage, progressPct, "Downloading...", fileName, formatSpeed(speed), downloaded, total)
				} else if reporter != nil {
					reporter.ReportDownload(stage, progressPct, "Downloading...", fileName, formatSpeed(speed), downloaded, total)
				}

				lastUpdate = now
				lastBytes = downloaded
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}

			if isRetryable(err) {
				return err
			}

			return fmt.Errorf("read error: %w", err)
		}
	}

	if err := out.Sync(); err != nil {
		return err
	}
	out.Close()

	if runtime.GOOS == "windows" {
		_ = os.Remove(dest)
	}

	if err := os.Rename(partPath, dest); err != nil {
		return err
	}

	if scaler != nil {
		scaler.ReportDownload(stage, 100, "Download complete", fileName, "", downloaded, total)
	} else if reporter != nil {
		reporter.ReportDownload(stage, 100, "Download complete", fileName, "", downloaded, total)
	}

	return nil
}

func isRetryable(err error) bool {
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	msg := err.Error()
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "timeout")
}

func createSafeClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   20 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Proxy: http.ProxyFromEnvironment,
	}

	return &http.Client{
		Transport: transport,
	}
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
