package patch

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/internal/platform"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/model"
)

func ApplyPWR(ctx context.Context, pwrFile string, request model.InstanceModel, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(request.Branch, request.BuildVersion)
	stagingDir := filepath.Join(gameDir, ".staging-temp")

	_ = os.MkdirAll(filepath.Dir(gameDir), 0755)
	_ = os.MkdirAll(stagingDir, 0755)

	butlerPath, err := GetButlerExec()
	if err != nil {
		fmt.Println("Can not get butler: %w", err)
	}

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		pwrFile,
		gameDir,
	)

	platform.HideConsoleWindow(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	reporter.Report(progress.StagePatch, 60, "Applying game patch...")

	if err := cmd.Run(); err != nil {
		return err
	}

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Process.Release()
	}

	// Clean up staging directory
	_ = os.RemoveAll(stagingDir)

	reporter.Report(progress.StagePatch, 100, "Game patched!")
	return nil
}

func DownloadPWR(ctx context.Context, branch string, targetVer int, reporter *progress.Reporter) (string, error) {
	cacheDir := env.GetCacheDir()
	_ = os.MkdirAll(cacheDir, 0755)

	osName := runtime.GOOS
	arch := runtime.GOARCH

	fileName := fmt.Sprintf("%d.pwr", targetVer)
	dest := filepath.Join(cacheDir, fileName)
	tempDest := dest + ".tmp"

	_ = os.Remove(tempDest)

	if _, err := os.Stat(dest); err == nil {
		reporter.Report(progress.StagePWR, 100, "PWR file cached")
		return dest, nil
	}

	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%s",
		osName, arch, branch, fileName)

	reporter.Report(progress.StagePWR, 0, "Downloading PWR file...")

	scaler := progress.NewScaler(reporter, progress.StagePWR, 0, 100)

	if err := download.DownloadWithReporter(dest, url, fileName, reporter, progress.StagePWR, scaler); err != nil {
		_ = os.Remove(tempDest)
		return "", err
	}

	reporter.Report(progress.StagePWR, 100, "PWR file downloaded")

	return dest, nil
}
