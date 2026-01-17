package updater

import (
	"HyLauncher/internal/util"
	"HyLauncher/internal/util/download"
	"context"
	"fmt"
	"os"
)

// Downloads latest launcher, returns path to temp file. If cant download deletes temp file
// !! Possible misunderstanding !! DownloadUpdate downloads any file, right now
// it downloads any file, as hylauncher-update-*.tmp file.
// TODO normal naming handling
// TODO need to be update function name to: DownloadTemp
func DownloadUpdate(
	ctx context.Context,
	url string,
	progress func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64),
) (string, error) {

	tmpPath, err := util.CreateTempFile("hylauncher-update-*")
	if err != nil {
		return "", err
	}

	if err := download.DownloadWithProgress(tmpPath, url, "update", 1.0, progress); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	fmt.Printf("Download complete: %s\n", tmpPath)
	return tmpPath, nil
}
