package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/internal/service"
	"HyLauncher/pkg/hyerrors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var AppVersion string = config.Default().Version

type App struct {
	ctx      context.Context
	cfg      *config.Config
	progress *progress.Reporter
	branch   string

	gameSvc *service.GameService
}

func NewApp() *App {
	cfg := config.New()
	return &App{
		cfg: cfg,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.progress = progress.New(ctx)

	a.gameSvc = service.NewGameService(
		ctx,
		a.progress,
	)

	branch, err := config.GetBranch()
	if err != nil {
		fmt.Printf("Game version grabbed: %s\n", err)
	}

	a.branch = branch

	fmt.Println("Application starting up...")
	fmt.Printf("Current launcher version: %s\n", AppVersion)
	fmt.Printf("Game version picker: %s\n", branch)

	go func() {
		fmt.Println("Creating folders...")
		env.CreateFolders(branch)
	}()

	// Check for launcher updates in background
	go func() {
		fmt.Println("Starting background update check...")
		a.checkUpdateSilently()
	}()

	go func() {
		fmt.Println("Starting cleanup")
		env.CleanupLauncher(branch)
	}()
}

// handleError creates an AppError, emits it to frontend, and returns it
func (a *App) handleError(errType hyerrors.ErrorType, userMsg string, err error) error {
	appErr := hyerrors.NewAppError(errType, userMsg, err)
	a.emitError(appErr)
	return appErr
}

// emitError sends structured errors to frontend
func (a *App) emitError(err error) {
	if appErr, ok := err.(*hyerrors.AppError); ok {
		runtime.EventsEmit(a.ctx, "error", appErr)
	} else {
		runtime.EventsEmit(a.ctx, "error", hyerrors.NewAppError(
			hyerrors.ErrorTypeUnknown,
			err.Error(),
			err,
		))
	}
}

func (a *App) DownloadAndLaunch(playerName string) error {
	// Validate nickname
	if len(playerName) == 0 {
		return a.handleError(
			hyerrors.ErrorTypeValidation,
			"Please enter a nickname",
			nil,
		)
	}

	if len(playerName) > 16 {
		return a.handleError(
			hyerrors.ErrorTypeValidation,
			"Nickname is too long (max 16 characters)",
			nil,
		)
	}

	// Ensure game is installed
	if err := a.gameSvc.EnsureInstalled(a.ctx, a.branch, a.progress); err != nil {
		return a.handleError(
			hyerrors.ErrorTypeGame,
			"Failed to install or update game",
			err,
		)
	}

	// Launch the game
	if err := a.gameSvc.Launch(playerName, a.branch); err != nil {
		return a.handleError(
			hyerrors.ErrorTypeGame,
			"Failed to launch game",
			err,
		)
	}

	return nil
}

func (a *App) GetLogs() (string, error) {
	logFile := filepath.Join(env.GetDefaultAppDir(), "logs", "errors.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
