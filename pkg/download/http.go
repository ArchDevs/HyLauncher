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
	"strconv"
	"strings"
	"time"

	"HyLauncher/internal/progress"
)

const (
	maxRetries     = 5
	baseRetryDelay = 3 * time.Second
	downloadLimit  = 45 * time.Minute
	readTimeout    = 60 * time.Second
)

var debugDownload = true

func DownloadWithReporter(
	ctx context.Context,
	dest string,
	url string,
	fileName string,
	reporter *progress.Reporter,
	stage progress.Stage,
	scaler *progress.Scaler,
) error {
	// Allow caller to cancel
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Check context before retry
		select {
		case <-ctx.Done():
			return fmt.Errorf("download canceled: %w", ctx.Err())
		default:
		}

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

			// Interruptible sleep
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return fmt.Errorf("download canceled during retry: %w", ctx.Err())
			case <-timer.C:
			}
		}

		err := attemptDownload(ctx, dest, url, fileName, reporter, stage, scaler)
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
	ctx context.Context,
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

	// Trust the caller's context for overall timeout; no extra wrapping here.
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

	if debugDownload {
		fmt.Printf(
			"Download debug: status=%d resume=%v length=%d accept-ranges=%q content-range=%q\n",
			resp.StatusCode,
			resumeFrom > 0,
			resp.ContentLength,
			resp.Header.Get("Accept-Ranges"),
			resp.Header.Get("Content-Range"),
		)
	}

	// Resume safety checks
	resumeValid := false
	if resumeFrom > 0 {
		// Normalize Accept-Ranges check
		acceptRanges := strings.ToLower(resp.Header.Get("Accept-Ranges"))
		if resp.StatusCode == http.StatusPartialContent && acceptRanges == "bytes" {
			// Parse Content-Range to verify server honored our request
			contentRange := resp.Header.Get("Content-Range")
			if contentRange != "" {
				// Format: "bytes START-END/TOTAL" or "bytes START-END/*"
				if strings.HasPrefix(contentRange, "bytes ") {
					parts := strings.Split(contentRange[6:], "/")
					if len(parts) == 2 {
						rangePart := strings.Split(parts[0], "-")
						if len(rangePart) == 2 {
							start, parseErr := strconv.ParseInt(rangePart[0], 10, 64)
							if parseErr == nil && start == resumeFrom {
								resumeValid = true
							} else if parseErr == nil {
								reportWarning(reporter, scaler, stage,
									fmt.Sprintf("Server returned wrong range (expected %d, got %d); restarting", resumeFrom, start))
							}
						}
					}
				}
			}
		}

		if !resumeValid {
			reportWarning(reporter, scaler, stage, "Server refused range request; restarting download")
			_ = os.Remove(partPath)
			resumeFrom = 0
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("bad HTTP status: %s", resp.Status)
	}

	// Handle unknown content length
	total := resp.ContentLength
	unknownSize := false
	if total < 0 {
		total = 0
		unknownSize = true
	} else if resumeFrom > 0 && resumeValid {
		total += resumeFrom
	}

	// Check if existing .part exceeds reported total
	if !unknownSize && resumeFrom > total {
		reportWarning(reporter, scaler, stage,
			fmt.Sprintf("Partial file (%d bytes) exceeds total (%d bytes); truncating", resumeFrom, total))
		_ = os.Remove(partPath)
		resumeFrom = 0
	}

	// Check disk space
	if total > 0 {
		if err := checkDiskSpace(dest, total-resumeFrom); err != nil {
			reportWarning(reporter, scaler, stage, fmt.Sprintf("Disk space warning: %v", err))
		}
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resumeFrom > 0 && resumeValid {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
		resumeFrom = 0
	}

	out, err := os.OpenFile(partPath, flags, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	// Wrap body with timeout reader
	bodyReader := &timeoutReader{
		r:          resp.Body,
		timeout:    readTimeout,
		lastReadAt: time.Now(),
	}

	buf := make([]byte, 64*1024)
	downloaded := resumeFrom
	lastUpdate := time.Now()
	lastBytes := downloaded

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := bodyReader.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}

			downloaded += int64(n)
			now := time.Now()

			if now.Sub(lastUpdate) >= 200*time.Millisecond {
				speed := float64(downloaded-lastBytes) / now.Sub(lastUpdate).Seconds()
				progressPct := 0.0

				if unknownSize {
					// Show progress without percentage
					if scaler != nil {
						scaler.ReportDownload(stage, 0, "Downloading (unknown size)...", fileName, formatSpeed(speed), downloaded, 0)
					} else if reporter != nil {
						reporter.ReportDownload(stage, 0, "Downloading (unknown size)...", fileName, formatSpeed(speed), downloaded, 0)
					}
				} else {
					if total > 0 {
						progressPct = float64(downloaded) / float64(total) * 100
					}
					if scaler != nil {
						scaler.ReportDownload(stage, progressPct, "Downloading...", fileName, formatSpeed(speed), downloaded, total)
					} else if reporter != nil {
						reporter.ReportDownload(stage, progressPct, "Downloading...", fileName, formatSpeed(speed), downloaded, total)
					}
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

	if !unknownSize && total > 0 && downloaded != total {
		return fmt.Errorf("incomplete download: got %d bytes, expected %d", downloaded, total)
	}

	if err := out.Sync(); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	if err := renameWithRetry(partPath, dest); err != nil {
		return err
	}

	finalMsg := "Download complete"
	finalPct := 100.0
	if unknownSize {
		finalMsg = fmt.Sprintf("Download complete (%s)", formatBytes(downloaded))
		finalPct = 0
	}

	if scaler != nil {
		scaler.ReportDownload(stage, finalPct, finalMsg, fileName, "", downloaded, total)
	} else if reporter != nil {
		reporter.ReportDownload(stage, finalPct, finalMsg, fileName, "", downloaded, total)
	}

	return nil
}

// timeoutReader wraps an io.Reader and enforces read timeout
type timeoutReader struct {
	r          io.Reader
	timeout    time.Duration
	lastReadAt time.Time
}

func (tr *timeoutReader) Read(p []byte) (n int, err error) {
	if time.Since(tr.lastReadAt) > tr.timeout {
		return 0, fmt.Errorf("read timeout: no data received for %v", tr.timeout)
	}

	n, err = tr.r.Read(p)
	if n > 0 {
		tr.lastReadAt = time.Now()
	}
	return n, err
}

func isRetryable(err error) bool {
	// Don't retry context cancellations – check first so nothing below
	// accidentally short-circuits a DeadlineExceeded that wraps a net.Error.
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "tls handshake")
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
		ResponseHeaderTimeout: 30 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Proxy: http.ProxyFromEnvironment,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   downloadLimit,
	}
}

func renameWithRetry(oldPath, newPath string) error {
	if runtime.GOOS == "windows" {
		// Remove destination first on Windows
		_ = os.Remove(newPath)
	}

	maxAttempts := 3
	for i := 0; i < maxAttempts; i++ {
		err := os.Rename(oldPath, newPath)
		if err == nil {
			return nil
		}

		// Only retry on Windows for specific errors
		if runtime.GOOS == "windows" && i < maxAttempts-1 {
			time.Sleep(time.Second)
			continue
		}

		return err
	}

	return fmt.Errorf("rename failed after %d attempts", maxAttempts)
}

func reportWarning(reporter *progress.Reporter, scaler *progress.Scaler, stage progress.Stage, msg string) {
	fmt.Println("Warning:", msg)
	if scaler != nil {
		scaler.Report(stage, 0, msg)
	} else if reporter != nil {
		reporter.Report(stage, 0, msg)
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

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
