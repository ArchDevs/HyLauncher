package game

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/internal/java"
)

func Launch(playerName string, version string) error {
	baseDir := env.GetDefaultAppDir()

	gameDir := filepath.Join(baseDir, "release", "package", "game", version)
	userDataDir := filepath.Join(baseDir, "UserData")

	gameClient := "HytaleClient"
	if runtime.GOOS == "windows" {
		gameClient += ".exe"
	}

	clientPath := filepath.Join(gameDir, "Client", gameClient)
	javaBin := java.GetJavaExec()

	_ = os.MkdirAll(userDataDir, 0755)

	playerUUID := OfflineUUID(playerName).String()

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
		"Launching %s (%s) with UUID %s\n",
		playerName,
		version,
		playerUUID,
	)

	return cmd.Start()
}
