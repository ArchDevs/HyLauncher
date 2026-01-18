package game

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/internal/java"
)

func Launch(playerName string, channel string, playerUUID string, version string, enableOnlineFix bool) (*exec.Cmd, error) {
	baseDir := env.GetDefaultAppDir()
	gameDir := filepath.Join(baseDir, channel, "package", "game", version)
	userDataDir := filepath.Join(baseDir, "UserData")

	if enableOnlineFix {
		if err := EnsureServerAndClientFix(context.Background(), gameDir, nil); err != nil {
			return nil, err
		}
	}

	gameClient := "HytaleClient"
	if runtime.GOOS == "windows" {
		gameClient += ".exe"
	}

	clientPath := filepath.Join(gameDir, "Client", gameClient)
	// Check if client executable exists
	if _, err := os.Stat(clientPath); err != nil {
		return nil, fmt.Errorf("game executable not found at %s: %w", clientPath, err)
	}

	javaBin, err := java.GetJavaExec()
	if err != nil {
		return nil, err
	}

	_ = os.MkdirAll(userDataDir, 0755)

	cmd := exec.Command(clientPath,
		"--app-dir", gameDir,
		"--user-dir", userDataDir,
		"--java-exec", javaBin,
		"--auth-mode", "offline",
		"--uuid", playerUUID,
		"--name", playerName,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	setSDLVideoDriver(cmd)

	fmt.Printf(
		"Launching %s (%s - %s) with UUID %s\n",
		playerName,
		channel,
		version,
		playerUUID,
	)

	return cmd, cmd.Start()
}
