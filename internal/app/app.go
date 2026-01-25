package app

import (
	"context"
	"fmt"
	"time"

	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/internal/service"
	"HyLauncher/pkg/hyerrors"
	"HyLauncher/pkg/model"

	"github.com/hugolgst/rich-go/client"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var AppVersion string = config.LauncherDefault().Version

type App struct {
	ctx         context.Context
	launcherCfg *config.LauncherConfig
	instanceCfg *config.InstanceConfig
	progress    *progress.Reporter
	instance    model.InstanceModel

	crashSvc *service.Reporter
	gameSvc  *service.GameService
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.progress = progress.New(ctx)

	hyerrors.RegisterHandlerFunc(func(err *hyerrors.Error) {
		runtime.EventsEmit(ctx, "error", err)
	})

	err := client.Login("1465005878276128888")
	if err != nil {
		panic(err)
	}

	launcherCfg, err := config.LoadLauncher()
	if err != nil {
		panic(err) // launcher config is critical
	}
	a.launcherCfg = launcherCfg

	instanceName := launcherCfg.Instance
	instanceCfg, err := config.LoadInstance(instanceName)
	if err != nil {
		panic(err)
	}
	a.instanceCfg = instanceCfg

	instance, err := config.LoadInstance(instanceName)
	if err != nil {
		hyerrors.WrapConfig(err, "failed to get instance").
			WithContext("default_instance", "default")
		config.UpdateInstance(instanceName, func(cfg *config.InstanceConfig) error {
			cfg.ID = instanceName
			return nil
		})
	}

	crashReporter, err := service.NewCrashReporter(
		env.GetDefaultAppDir(),
		AppVersion,
	)
	if err != nil {
		fmt.Printf("failed to initialize diagnostics: %v\n", err)
	}

	a.instance.Branch = a.instanceCfg.Branch
	a.instance.BuildVersion = instance.Build
	a.instance.InstanceID = instance.ID
	a.instance.InstanceName = instance.Name

	a.crashSvc = crashReporter
	a.gameSvc = service.NewGameService(ctx, a.progress)

	fmt.Printf("Application starting: v%s, branch=%s\n", AppVersion, a.instance.Branch)

	go a.discordRPC()
	go env.CreateFolders(a.instance.InstanceID)
	go a.checkUpdateSilently()
	go env.CleanupLauncher(a.instance)
}

func (a *App) DownloadAndLaunch(playerName string) error {
	if err := a.validatePlayerName(playerName); err != nil {
		hyerrors.Report(hyerrors.Validation("provided invalid username"))
		return err
	}

	if err := a.gameSvc.EnsureInstalled(a.ctx, a.instance, a.progress); err != nil {
		appErr := hyerrors.WrapGame(err, "failed to install game").
			WithContext("branch", a.instance.Branch)
		hyerrors.Report(appErr)
		return appErr
	}

	if err := a.gameSvc.Launch(playerName, a.instance); err != nil {
		appErr := hyerrors.GameCritical("failed to launch game").
			WithDetails(err.Error()).
			WithContext("player", playerName).
			WithContext("branch", a.instance.Branch)
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
	if a.crashSvc == nil {
		return "", fmt.Errorf("diagnostics not initialized")
	}
	return a.crashSvc.GetLogs()
}

func (a *App) GetCrashReports() ([]service.CrashReport, error) {
	if a.crashSvc == nil {
		return nil, fmt.Errorf("diagnostics not initialized")
	}
	return a.crashSvc.GetCrashReports()
}

func (a *App) discordRPC() {
	now := time.Now()

	err := client.SetActivity(client.Activity{
		State:   "Idle",
		Details: "The best Hytale launcher",
		Timestamps: &client.Timestamps{
			Start: &now,
		},
		Buttons: []*client.Button{
			&client.Button{
				Label: "GitHub",
				Url:   "https://github.com/ArchDevs/HyLauncher",
			},
			&client.Button{
				Label: "Website",
				Url:   "https://hylauncher.fun",
			},
		},
	})

	if err != nil {
		fmt.Println("Error occured, %w", err)
	}
}
