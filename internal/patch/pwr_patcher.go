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
)

func ApplyPWR(ctx context.Context, pwrFile string, sigFile string, branch string, version int, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(branch, version)
	// Keep staging dir OUTSIDE game directory to avoid verification issues
	stagingDir := filepath.Join(env.GetCacheDir(), "staging-temp")

	// Clean up any previous incomplete staging state
	_ = os.RemoveAll(stagingDir)

	_ = os.MkdirAll(gameDir, 0755)
	_ = os.MkdirAll(stagingDir, 0755)

	butlerPath, err := GetButlerExec()
	if err != nil {
		return fmt.Errorf("cannot get butler: %w", err)
	}

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		"--signature", sigFile,
		pwrFile,
		gameDir,
	)

	platform.HideConsoleWindow(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if reporter != nil {
		reporter.Report(progress.StagePatch, 60, "Applying game patch...")
	}

	if err := cmd.Run(); err != nil {
		// Clean up staging on failure
		_ = os.RemoveAll(stagingDir)
		return err
	}

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Process.Release()
	}

	// Clean up staging directory
	_ = os.RemoveAll(stagingDir)

	if reporter != nil {
		reporter.Report(progress.StagePatch, 80, "Game patched!")
	}
	return nil
}

func DownloadPWR(ctx context.Context, branch string, targetVer int, reporter *progress.Reporter) (pwrPath string, sigPath string, err error) {
	cacheDir := env.GetCacheDir()
	_ = os.MkdirAll(cacheDir, 0755)

	osName := runtime.GOOS
	arch := runtime.GOARCH

	pwrFileName := fmt.Sprintf("%d.pwr", targetVer)
	sigFileName := fmt.Sprintf("%d.pwr.sig", targetVer)

	pwrDest := filepath.Join(cacheDir, pwrFileName)
	sigDest := filepath.Join(cacheDir, sigFileName)

	baseURL := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0", osName, arch, branch)

	// Check if both files are already cached
	_, pwrErr := os.Stat(pwrDest)
	_, sigErr := os.Stat(sigDest)
	if pwrErr == nil && sigErr == nil {
		reporter.Report(progress.StagePWR, 100, "PWR files cached")
		return pwrDest, sigDest, nil
	}

	// Download PWR file
	reporter.Report(progress.StagePWR, 0, "Downloading PWR file...")
	pwrURL := fmt.Sprintf("%s/%s", baseURL, pwrFileName)
	scaler := progress.NewScaler(reporter, progress.StagePWR, 0, 70)

	if err := download.DownloadWithReporter(ctx, pwrDest, pwrURL, pwrFileName, reporter, progress.StagePWR, scaler); err != nil {
		_ = os.Remove(pwrDest + ".tmp")
		return "", "", err
	}

	// Download signature file
	reporter.Report(progress.StagePWR, 70, "Downloading signature file...")
	sigURL := fmt.Sprintf("%s/%s", baseURL, sigFileName)
	sigScaler := progress.NewScaler(reporter, progress.StagePWR, 70, 100)

	if err := download.DownloadWithReporter(ctx, sigDest, sigURL, sigFileName, reporter, progress.StagePWR, sigScaler); err != nil {
		_ = os.Remove(sigDest + ".tmp")
		return "", "", err
	}

	reporter.Report(progress.StagePWR, 100, "PWR files downloaded")

	return pwrDest, sigDest, nil
}
