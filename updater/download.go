package updater

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/util/download"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func DownloadUpdate(
	ctx context.Context,
	url string,
	progress func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64),
) (string, error) {

	tmpFile, err := os.CreateTemp(
		filepath.Join(env.GetDefaultAppDir(), "cache"),
		"hylauncher-update-*",
	)
	if err != nil {
		return "", err
	}

	tmp := tmpFile.Name()
	tmpFile.Close()

	if err := download.DownloadWithProgress(tmp, url, "update", 1.0, progress); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}

	fmt.Printf("Download complete: %s\n", tmp)
	return tmp, nil
}
