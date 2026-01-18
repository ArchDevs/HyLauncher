package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/game"
	"HyLauncher/internal/patch"
	"HyLauncher/pkg/hyerrors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	cfg     *config.Config
	gameCmd *exec.Cmd
}

type GameVersions struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
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

	go func() {
		fmt.Println("Starting cleanup")
		env.CleanupLauncher()
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
	if appErr, ok := err.(*hyerrors.AppError); ok {
		runtime.EventsEmit(a.ctx, "error", appErr)
	} else {
		runtime.EventsEmit(a.ctx, "error", hyerrors.NewAppError(hyerrors.ErrorTypeUnknown, err.Error(), err))
	}
}

var AppVersion string = config.Default().Version

func (a *App) GetGameVersions(channel string) []int {
	latest := patch.FindLatestVersion(channel)
	if latest == 0 {
		return []int{}
	}
	versions := make([]int, 0, latest)
	for i := latest; i >= 1; i-- {
		versions = append(versions, i)
	}
	return versions
}

func (a *App) GetVersions() GameVersions {
	channel := a.cfg.Settings.Channel
	if channel == "" {
		channel = "release"
	}
	current := patch.GetLocalVersion(channel)
	latest := patch.FindLatestVersion(channel)
	fmt.Printf("GetVersions: Channel=%s, Current=%s, Latest=%d\n", channel, current, latest)
	return GameVersions{
		Current: current,
		Latest:  strconv.Itoa(latest),
	}
}

func (a *App) DownloadAndLaunch(playerName string) error {
	// Validate nickname
	if len(playerName) == 0 {
		err := hyerrors.NewAppError(
			hyerrors.ErrorTypeValidation,
			"Please enter a nickname",
			nil,
		)
		a.emitError(err)
		return err
	}

	if len(playerName) > 16 {
		err := hyerrors.NewAppError(
			hyerrors.ErrorTypeValidation,
			"Nickname is too long (max 16 characters)",
			nil,
		)
		a.emitError(err)
		return err
	}

	channel := a.cfg.Settings.Channel
	if channel == "" {
		channel = "release"
	}
	targetVersion := a.cfg.Settings.GameVersion

	// Ensure game is installed
	if err := game.EnsureInstalled(a.ctx, channel, targetVersion, a.cfg.Settings.OnlineFix, a.progressCallback); err != nil {
		wrappedErr := hyerrors.NewAppError(hyerrors.ErrorTypeGame, "Failed to install or update game", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	// Launch the game
	a.progressCallback("launch", 100, "Launching game...", "", "", 0, 0)

	// Use the current profile's ID as the UUID to ensure persistence across name changes
	// and consistency with the config file
	playerUUID := a.cfg.CurrentProfile

	versionStr := "latest"
	if a.cfg.Settings.GameVersion != 0 {
		versionStr = strconv.Itoa(a.cfg.Settings.GameVersion)
	}

	cmd, err := game.Launch(playerName, channel, playerUUID, versionStr, a.cfg.Settings.OnlineFix)
	if err != nil {
		wrappedErr := hyerrors.NewAppError(hyerrors.ErrorTypeGame, "Failed to launch game", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	a.gameCmd = cmd
	runtime.EventsEmit(a.ctx, "game-launched", nil)

	// Monitor game process
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Printf("Game process exited with error: %v\n", err)
		}
		a.gameCmd = nil
		runtime.EventsEmit(a.ctx, "game-closed", nil)
	}()

	return nil
}

func (a *App) StopGame() {
	if a.gameCmd != nil && a.gameCmd.Process != nil {
		if err := a.gameCmd.Process.Kill(); err != nil {
			fmt.Printf("Failed to kill game process: %v\n", err)
		}
		a.gameCmd = nil
		runtime.EventsEmit(a.ctx, "game-closed", nil)
	}
}

func (a *App) GetLogs() (string, error) {
	logFile := filepath.Join(env.GetDefaultAppDir(), "logs", "errors.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
