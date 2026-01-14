package pwr

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/pwr/butler"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func InstallGame(ctx context.Context, version, fileName string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	gameLatest := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")

	// Check if Hytale client already exists
	gameClient := "HytaleClient"
	if os.PathSeparator == '\\' {
		gameClient += ".exe"
	}
	clientPath := filepath.Join(gameLatest, "Client", gameClient)
	if _, err := os.Stat(clientPath); err == nil {
		fmt.Println("Game already installed, skipping download.")
		if progressCallback != nil {
			progressCallback("game", 100, "Game already installed", "", "", 0, 0)
		}
		return nil
	}

	// Download .pwr if needed
	if progressCallback != nil {
		progressCallback("game", 0, "Downloading game files...", fileName, "", 0, 0)
	}

	pwrPath, err := DownloadPWR(ctx, version, fileName, progressCallback)
	if err != nil {
		return err
	}

	// Apply .pwr using Butler
	if progressCallback != nil {
		progressCallback("game", 50, "Extracting game files...", "", "", 0, 0)
	}

	return ApplyPWR(ctx, pwrPath, progressCallback)
}

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

	// Hide console window on Windows - handled in platform-specific file
	hideConsoleWindow(cmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Applying .pwr file...")
	if progressCallback != nil {
		progressCallback("game", 60, "Applying game patch...", "", "", 0, 0)
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	_ = os.RemoveAll(stagingDir)
	fmt.Println("Game extracted successfully")

	if progressCallback != nil {
		progressCallback("game", 100, "Game installed successfully", "", "", 0, 0)
	}

	return nil
}
