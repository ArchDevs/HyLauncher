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

const playerUUID string = "1da855d2-6219-4d02-ad93-c4b160b073c3"

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

	fmt.Printf("Launching %s from %s with UserData at %s...\n", playerName, version, userDataDir)
	return cmd.Start()
}
