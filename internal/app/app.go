package app

import (
	"context"
	"fmt"

	"HyLauncher/internal/config"
	"HyLauncher/internal/diagnostics"
	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/internal/service"
	"HyLauncher/pkg/hyerrors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var AppVersion string = config.Default().Version

type App struct {
	ctx         context.Context
	cfg         *config.Config
	progress    *progress.Reporter
	diagnostics *diagnostics.Reporter
	branch      string
	gameSvc     *service.GameService
}

func NewApp() *App {
	return &App{
		cfg: config.New(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.progress = progress.New(ctx)

	reporter, err := diagnostics.NewReporter(
		env.GetDefaultAppDir(),
		AppVersion,
	)
	if err != nil {
		fmt.Printf("failed to initialize diagnostics: %v\n", err)
	}
	a.diagnostics = reporter

	hyerrors.RegisterHandlerFunc(func(err *hyerrors.Error) {
		runtime.EventsEmit(ctx, "error", err)
	})

	a.gameSvc = service.NewGameService(ctx, a.progress)

	branch, err := config.GetBranch()
	if err != nil {
		hyerrors.WrapConfig(err, "failed to get branch").
			WithContext("default_branch", "stable")
		branch = "stable"
	}
	a.branch = branch

	fmt.Printf("Application starting: v%s, branch=%s\n", AppVersion, branch)

	go env.CreateFolders(branch)
	go a.checkUpdateSilently()
	go env.CleanupLauncher(branch)
}

func (a *App) DownloadAndLaunch(playerName string) error {
	if err := a.validatePlayerName(playerName); err != nil {
		hyerrors.Report(hyerrors.Validation("provided invalid username"))
		return err
	}

	if err := a.gameSvc.EnsureInstalled(a.ctx, a.branch, a.progress); err != nil {
		appErr := hyerrors.WrapGame(err, "failed to install game").
			WithContext("branch", a.branch)
		hyerrors.Report(appErr)
		return appErr
	}

	if err := a.gameSvc.Launch(playerName, a.branch); err != nil {
		appErr := hyerrors.GameCritical("failed to launch game").
			WithDetails(err.Error()).
			WithContext("player", playerName).
			WithContext("branch", a.branch)
		hyerrors.Report(appErr)
		return appErr
	}

	return nil
}

func (a *App) validatePlayerName(name string) error {
	if len(name) == 0 {
		return hyerrors.Validation("please enter a nickname")
	}
	if len(name) > 16 {
		return hyerrors.Validation("nickname too long (max 16 characters)").
			WithContext("length", len(name))
	}
	return nil
}

func (a *App) GetLogs() (string, error) {
	if a.diagnostics == nil {
		return "", fmt.Errorf("diagnostics not initialized")
	}
	return a.diagnostics.GetLogs()
}

func (a *App) GetCrashReports() ([]diagnostics.CrashReport, error) {
	if a.diagnostics == nil {
		return nil, fmt.Errorf("diagnostics not initialized")
	}
	return a.diagnostics.GetCrashReports()
}
