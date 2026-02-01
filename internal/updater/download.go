package updater

import (
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/fileutil"
	"context"
	"fmt"
	"os"
)

// Downloads latest launcher, returns path to temp file. If cant download deletes temp file
func DownloadTemp(
	ctx context.Context,
	url string,
	reporter *progress.Reporter,
) (string, error) {

	tmpPath, err := fileutil.CreateTempFile("file-update-*")
	if err != nil {
		return "", err
	}

	reporter.Report(progress.StageUpdate, 0, "Downloading launcher update...")

	scaler := progress.NewScaler(reporter, progress.StageUpdate, 0, 100)

	if err := download.DownloadWithReporter(ctx, tmpPath, url, "launcher", reporter, progress.StageUpdate, scaler); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	fmt.Printf("Download complete: %s\n", tmpPath)
	reporter.Report(progress.StageUpdate, 100, "Download complete")

	return tmpPath, nil
}
