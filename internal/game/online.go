package game

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/util"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nwaples/rardecode"
)

const (
	onlineFixURL      = "https://uploads.online-fix.me:2053/uploads/Hytale/Fix%20Repair/Hytale_Fix_Repair_V3.rar"
	onlineFixPassword = "online-fix.me"
)

func ApplyOnlineFixWindows(ctx context.Context, gameDir string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("online fix is only for Windows")
	}

	cacheDir := filepath.Join(gameDir, ".cache")
	_ = os.MkdirAll(cacheDir, 0755)

	rarPath := filepath.Join(cacheDir, "online_fix.rar")
	tmpRAR := rarPath + ".tmp"

	if progressCallback != nil {
		progressCallback("online-fix", 0, "Downloading online-fix...", "Hytale_Fix_Repair_V3.rar", "", 0, 0)
	}

	// Скачиваем архив
	if err := util.DownloadWithProgress(tmpRAR, onlineFixURL, "online-fix", 0.6, progressCallback); err != nil {
		_ = os.Remove(tmpRAR)
		return err
	}
	if err := os.Rename(tmpRAR, rarPath); err != nil {
		return err
	}

	// Открываем RAR
	file, err := os.Open(rarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	r, err := rardecode.NewReader(file, onlineFixPassword)
	if err != nil {
		return fmt.Errorf("failed to open RAR: %w", err)
	}

	// Временная папка для распаковки
	tempDir := filepath.Join(cacheDir, "temp_extract")
	_ = os.RemoveAll(tempDir)
	_ = os.MkdirAll(tempDir, 0755)

	if progressCallback != nil {
		progressCallback("online-fix", 30, "Extracting archive...", "", "", 0, 0)
	}

	for {
		hdr, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Игнорируем папки
		if strings.HasSuffix(hdr.Name, "/") {
			continue
		}

		outPath := filepath.Join(tempDir, filepath.Base(hdr.Name))
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, r); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}

	// 1. Заменяем HytaleClient.exe
	clientPath := filepath.Join(gameDir, "Client", "HytaleClient.exe")
	tempClient := clientPath + ".tmp"

	extractedClient := filepath.Join(tempDir, "HytaleClient.exe")
	if _, err := os.Stat(extractedClient); err != nil {
		return fmt.Errorf("extracted client not found in archive")
	}

	if err := util.СopyFile(extractedClient, tempClient); err != nil {
		return err
	}
	if err := os.Rename(tempClient, clientPath); err != nil {
		return err
	}

	if progressCallback != nil {
		progressCallback("online-fix", 70, "Client replaced", "HytaleClient.exe", "", 0, 0)
	}

	// 2. Заменяем Server
	serverDir := filepath.Join(gameDir, "Server")
	tempServer := filepath.Join(tempDir, "Server")
	if _, err := os.Stat(tempServer); err != nil {
		return fmt.Errorf("extracted Server folder not found in archive")
	}

	if err := os.RemoveAll(serverDir); err != nil {
		return err
	}

	if err := util.СopyDir(tempServer, serverDir); err != nil {
		return err
	}

	if progressCallback != nil {
		progressCallback("online-fix", 100, "Server replaced", "Server folder", "", 0, 0)
	}

	// Очистка
	_ = os.RemoveAll(tempDir)
	_ = os.Remove(rarPath)

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
