package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"HyLauncher/internal/env"
	"HyLauncher/internal/java"
	"HyLauncher/internal/pwr"
	"HyLauncher/internal/pwr/butler"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// ProgressUpdate represents a progress event
type ProgressUpdate struct {
	Stage       string  `json:"stage"`
	Progress    float64 `json:"progress"`
	Message     string  `json:"message"`
	CurrentFile string  `json:"currentFile"`
	Speed       string  `json:"speed"`
	Downloaded  int64   `json:"downloaded"`
	Total       int64   `json:"total"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	env.CreateFolders()
}

func (a *App) emitProgress(update ProgressUpdate) {
	runtime.EventsEmit(a.ctx, "progress-update", update)
}

// progressCallback is a wrapper that converts the callback signature to ProgressUpdate
func (a *App) progressCallback(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64) {
	a.emitProgress(ProgressUpdate{
		Stage:       stage,
		Progress:    progress,
		Message:     message,
		CurrentFile: currentFile,
		Speed:       speed,
		Downloaded:  downloaded,
		Total:       total,
	})
}

func (a *App) DownloadAndLaunch(playerName string) error {
	// JRE Download
	a.emitProgress(ProgressUpdate{
		Stage:    "jre",
		Progress: 0,
		Message:  "Checking JRE...",
	})

	if err := java.DownloadJRE(a.ctx, a.progressCallback); err != nil {
		return err
	}

	a.emitProgress(ProgressUpdate{
		Stage:    "jre",
		Progress: 100,
		Message:  "JRE ready",
	})

	javaBin := java.GetJavaExec()

	// Butler Installation
	a.emitProgress(ProgressUpdate{
		Stage:    "butler",
		Progress: 0,
		Message:  "Checking Butler...",
	})

	if _, err := butler.InstallButler(a.ctx, a.progressCallback); err != nil {
		return err
	}

	a.emitProgress(ProgressUpdate{
		Stage:    "butler",
		Progress: 100,
		Message:  "Butler ready",
	})

	version := "release"
	pwrFile := "1.pwr"

	gameLatest := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")
	gameClient := "HytaleClient"
	if os.PathSeparator == '\\' {
		gameClient += ".exe"
	}
	clientPath := filepath.Join(gameLatest, "Client", gameClient)

	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		// Game Installation
		a.emitProgress(ProgressUpdate{
			Stage:    "game",
			Progress: 0,
			Message:  "Installing game...",
		})

		if err := pwr.InstallGame(a.ctx, version, pwrFile, a.progressCallback); err != nil {
			return err
		}

		a.emitProgress(ProgressUpdate{
			Stage:    "game",
			Progress: 100,
			Message:  "Game installed",
		})
	} else {
		fmt.Println("Game already installed, skipping download.")
		a.emitProgress(ProgressUpdate{
			Stage:    "game",
			Progress: 100,
			Message:  "Game already installed",
		})
	}

	// Launch
	a.emitProgress(ProgressUpdate{
		Stage:    "launch",
		Progress: 100,
		Message:  "Launching game...",
	})

	uuid := uuid.NewString()
	cmd := exec.Command(clientPath,
		"--app-dir", gameLatest,
		"--java-exec", javaBin,
		"--auth-mode", "offline",
		"--uuid", uuid,
		"--name", playerName,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println("Launching game...")
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}
