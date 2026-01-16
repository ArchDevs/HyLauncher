package updater

import (
	"HyLauncher/internal/util/download"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func DownloadUpdate(ctx context.Context, url string, progress func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) (string, error) {
	fmt.Printf("Starting download from: %s\n", url)

	tmp := filepath.Join(os.TempDir(), "hylauncher-update.tmp")

	_ = os.Remove(tmp)

	if err := download.DownloadWithProgress(tmp, url, "update", 1.0, progress); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("failed to download update: %w", err)
	}

	fmt.Printf("Download complete: %s\n", tmp)
	return tmp, nil
}
