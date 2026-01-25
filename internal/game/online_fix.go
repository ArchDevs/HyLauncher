package game

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/model"

	"github.com/anacrolix/torrent"
	"github.com/mholt/archives"
)

const (
	metadataURL    = "https://hydralinks.pages.dev/sources/onlinefix.json"
	fixPassword    = "online-fix.me"
	torrentTimeout = 30 * time.Minute
	fixArchiveName = "Hytale_Fix_Repair.rar"
	fixFolderName  = "Fix Repair"
	gameIdentifier = "Hytale"
)

type OnlineFixIndex struct {
	Name      string     `json:"name"`
	Downloads []Download `json:"downloads"`
}

type Download struct {
	Title      string   `json:"title"`
	FileSize   string   `json:"fileSize"`
	UploadDate string   `json:"uploadDate"`
	URIs       []string `json:"uris"`
}

func ApplyOnlineFixWindows(ctx context.Context, gameDir string, reporter *progress.Reporter) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("online fix is only supported on Windows")
	}

	cacheDir := filepath.Join(env.GetCacheDir(), "onlinefix")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	defer os.RemoveAll(cacheDir)

	reporter.Report(progress.StageOnlineFix, 0, "Fetching metadata...")

	magnetLink, err := fetchMagnetLink(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch magnet link: %w", err)
	}

	reporter.Report(progress.StageOnlineFix, 5, "Starting selective download...")

	fixArchivePath := filepath.Join(cacheDir, fixArchiveName)
	if err := downloadFixArchive(ctx, magnetLink, fixArchivePath, reporter); err != nil {
		return fmt.Errorf("failed to download fix: %w", err)
	}

	reporter.Report(progress.StageOnlineFix, 80, "Extracting fix...")

	if err := extractAndApply(fixArchivePath, gameDir); err != nil {
		return fmt.Errorf("failed to extract fix: %w", err)
	}

	reporter.Report(progress.StageOnlineFix, 100, "Online fix applied")
	return nil
}

func fetchMagnetLink(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	var index OnlineFixIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return "", err
	}

	for _, download := range index.Downloads {
		if strings.Contains(download.Title, gameIdentifier) {
			if len(download.URIs) == 0 {
				return "", fmt.Errorf("no magnet links found")
			}
			return download.URIs[0], nil
		}
	}

	return "", fmt.Errorf("Hytale fix not found in metadata")
}

func downloadFixArchive(ctx context.Context, magnetLink, destPath string, reporter *progress.Reporter) error {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = filepath.Dir(destPath)
	cfg.NoUpload = true
	cfg.Seed = false

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	t, err := client.AddMagnet(magnetLink)
	if err != nil {
		return err
	}

	select {
	case <-t.GotInfo():
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("timeout waiting for torrent metadata")
	}

	var targetFile *torrent.File
	for _, file := range t.Files() {
		if strings.Contains(file.Path(), fixFolderName) && strings.HasSuffix(file.Path(), fixArchiveName) {
			targetFile = file
			file.SetPriority(torrent.PiecePriorityNormal)
		} else {
			file.SetPriority(torrent.PiecePriorityNone)
		}
	}

	if targetFile == nil {
		return fmt.Errorf("fix archive not found in torrent")
	}

	targetFile.Download()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, torrentTimeout)
	defer cancel()

	completed := false
	for !completed {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("torrent download timeout")
		case <-ticker.C:
			bytesCompleted := targetFile.BytesCompleted()
			totalBytes := targetFile.Length()

			if bytesCompleted >= totalBytes {
				completed = true
				break
			}

			progressPct := 0.0
			if totalBytes > 0 {
				progressPct = float64(bytesCompleted) / float64(totalBytes) * 100
			}

			scaledProgress := 5 + (progressPct * 0.75)
			reporter.Report(progress.StageOnlineFix, scaledProgress, "Downloading fix...")
		}
	}

	reporter.Report(progress.StageOnlineFix, 78, "Finalizing download...")

	t.Drop()
	client.Close()
	time.Sleep(2 * time.Second)

	sourcePath := filepath.Join(cfg.DataDir, targetFile.Path())
	partPath := sourcePath + ".part"

	var finalSource string
	if _, err := os.Stat(sourcePath); err == nil {
		finalSource = sourcePath
	} else if _, err := os.Stat(partPath); err == nil {
		finalSource = partPath
	} else {
		return fmt.Errorf("downloaded file not found at %s or %s", sourcePath, partPath)
	}

	if err := fileutil.CopyFile(finalSource, destPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func extractAndApply(archivePath, gameDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	format := archives.Rar{
		Password: fixPassword,
	}

	return format.Extract(context.Background(), file, func(ctx context.Context, f archives.FileInfo) error {
		if f.IsDir() {
			return nil
		}

		relPath := f.NameInArchive

		if strings.Contains(relPath, "Client") && strings.HasSuffix(relPath, "HytaleClient.exe") {
			targetPath := filepath.Join(gameDir, "Client", "HytaleClient.exe")
			return extractFileFromArchive(f, targetPath)
		}

		if strings.Contains(relPath, "Server") {
			if strings.HasSuffix(relPath, "HytaleServer.jar") {
				targetPath := filepath.Join(gameDir, "Server", "HytaleServer.jar")
				return extractFileFromArchive(f, targetPath)
			}
			if strings.HasSuffix(relPath, "start-server.bat") {
				targetPath := filepath.Join(gameDir, "Server", "start-server.bat")
				return extractFileFromArchive(f, targetPath)
			}
		}

		return nil
	})
}

func extractFileFromArchive(f archives.FileInfo, targetPath string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	dstFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, rc)
	return err
}

func EnsureServerAndClientFix(ctx context.Context, requst model.InstanceModel, reporter *progress.Reporter) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	gameLatestDir := filepath.Join(
		env.GetInstance(requst.InstanceID),
		strconv.Itoa(requst.BuildVersion),
	)

	serverBat := filepath.Join(gameLatestDir, "Server", "start-server.bat")
	if _, err := os.Stat(serverBat); err == nil {
		return nil
	}

	if reporter != nil {
		reporter.Report(progress.StageOnlineFix, 0, "Applying online fix...")
	}

	return ApplyOnlineFixWindows(ctx, gameLatestDir, reporter)
}
