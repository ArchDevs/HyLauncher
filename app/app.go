package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/game"
	"HyLauncher/internal/pwr"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	cfg *config.Config
}

type ProgressUpdate struct {
	Stage       string  `json:"stage"`
	Progress    float64 `json:"progress"`
	Message     string  `json:"message"`
	CurrentFile string  `json:"currentFile"`
	Speed       string  `json:"speed"`
	Downloaded  int64   `json:"downloaded"`
	Total       int64   `json:"total"`
}

func NewApp() *App {
	cfg, _ := config.Load()
	return &App{cfg: cfg}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	fmt.Println("Application starting up...")
	fmt.Printf("Current launcher version: %s\n", AppVersion)

	// Check for launcher updates in background
	go func() {
		fmt.Println("Starting background update check...")
		a.checkUpdateSilently()
	}()
}

func (a *App) progressCallback(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64) {
	runtime.EventsEmit(a.ctx, "progress-update", ProgressUpdate{
		Stage:       stage,
		Progress:    progress,
		Message:     message,
		CurrentFile: currentFile,
		Speed:       speed,
		Downloaded:  downloaded,
		Total:       total,
	})
}

// emitError sends structured errors to frontend
func (a *App) emitError(err error) {
	if appErr, ok := err.(*AppError); ok {
		runtime.EventsEmit(a.ctx, "error", appErr)
	} else {
		runtime.EventsEmit(a.ctx, "error", NewAppError(ErrorTypeUnknown, err.Error(), err))
	}
}

const AppVersion string = "0.6"

func (a *App) GetVersions() (currentVersion string, latestVersion string) {
	current := pwr.GetLocalVersion()
	latest := pwr.FindLatestVersion("release")
	return current, strconv.Itoa(latest)
}

func (a *App) DownloadAndLaunch(playerName string) error {
	// Validate nickname
	if len(playerName) == 0 {
		err := ValidationError("Please enter a nickname")
		a.emitError(err)
		return err
	}

	if len(playerName) > 16 {
		err := ValidationError("Nickname is too long (max 16 characters)")
		a.emitError(err)
		return err
	}

	// Ensure game is installed
	if err := game.EnsureInstalled(a.ctx, a.progressCallback); err != nil {
		wrappedErr := GameError("Failed to install or update game", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	// Launch the game
	a.progressCallback("launch", 100, "Launching game...", "", "", 0, 0)

	if err := game.Launch(playerName, "latest"); err != nil {
		wrappedErr := GameError("Failed to launch game", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	return nil
}

// GetLogs returns recent error logs for debugging
func (a *App) GetLogs() (string, error) {
	logFile := filepath.Join(env.GetDefaultAppDir(), "logs", "errors.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
