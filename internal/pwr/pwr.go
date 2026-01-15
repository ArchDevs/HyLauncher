package pwr

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"HyLauncher/internal/env"
	"HyLauncher/internal/pwr/butler"
	"HyLauncher/internal/util"
)

func ApplyPWR(ctx context.Context, pwrFile string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	gameLatest := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")

	butlerPath, err := butler.InstallButler(ctx, progressCallback)
	if err != nil {
		return err
	}

	stagingDir := filepath.Join(gameLatest, "staging-temp")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		pwrFile,
		gameLatest,
	)

	util.HideConsoleWindow(cmd)

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	cmd.Stdout = os.Stdout

	// Run the command **only once**
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("butler apply failed: %s: %w", errBuf.String(), err)
	}

	fmt.Println("Applying .pwr file...")
	if progressCallback != nil {
		progressCallback("game", 100, "Game installed successfully", "", "", 0, 0)
	}

	_ = os.RemoveAll(stagingDir)
	fmt.Println("Game extracted successfully")

	return nil
}
