package pwr

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/internal/util"
)

func ApplyPWR(ctx context.Context, pwrFile string,
	progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {

	gameLatest := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")
	stagingDir := filepath.Join(gameLatest, "staging-temp")
	_ = os.MkdirAll(stagingDir, 0755)

	butlerPath := filepath.Join(env.GetDefaultAppDir(), "tools", "butler", "butler")
	if runtime.GOOS == "windows" {
		butlerPath += ".exe"
	}

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		pwrFile,
		gameLatest,
	)

	util.HideConsoleWindow(cmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if progressCallback != nil {
		progressCallback("game", 60, "Applying game patch...", "", "", 0, 0)
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	// Retry rename on Windows if locked
	if runtime.GOOS == "windows" {
		for i := 0; i < 5; i++ {
			if err := os.Rename(stagingDir, gameLatest); err == nil {
				break
			}
			time.Sleep(2 * time.Second)
		}
	} else {
		_ = os.Rename(stagingDir, gameLatest)
	}

	if progressCallback != nil {
		progressCallback("game", 100, "Game installed successfully", "", "", 0, 0)
	}

	return nil
}
