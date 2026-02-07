package app

import (
	"context"
	"fmt"
	"regexp"
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

type App struct {
	ctx         context.Context
	launcherCfg *config.LauncherConfig
	instanceCfg *config.InstanceConfig
	progress    *progress.Reporter
	instance    model.InstanceModel

	crashSvc *service.Reporter
	gameSvc  *service.GameService
	authSvc  *service.AuthService
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.progress = progress.New(ctx)

	hyerrors.RegisterHandlerFunc(func(err *hyerrors.Error) {
		runtime.EventsEmit(a.ctx, "error", err)
	})

	if err := client.Login("1465005878276128888"); err != nil {
		fmt.Printf("failed to initialize Discord RPC: %v\n", err)
	}

	launcherCfg, err := config.LoadLauncher()
	if err != nil {
		panic(fmt.Errorf("failed to load launcher config: %w", err))
	}
	a.launcherCfg = launcherCfg

	instanceName := launcherCfg.Instance
	instanceCfg, err := config.LoadInstance(instanceName)
	if err != nil {
		hyerrors.WrapConfig(err, "failed to load instance").
			WithContext("instance", instanceName)
		_ = config.UpdateInstance(instanceName, func(cfg *config.InstanceConfig) error {
			cfg.ID = instanceName
			return nil
		})
		panic(fmt.Errorf("failed to load instance config %q: %w", instanceName, err))
	}
	a.instanceCfg = instanceCfg

	a.instance.Branch = instanceCfg.Branch
	a.instance.BuildVersion = instanceCfg.Build
	a.instance.InstanceID = instanceCfg.ID
	a.instance.InstanceName = instanceCfg.Name

	crashReporter, err := service.NewCrashReporter(
		env.GetDefaultAppDir(),
		a.launcherCfg.Version,
	)
	if err != nil {
		fmt.Printf("failed to initialize diagnostics: %v\n", err)
	} else {
		a.crashSvc = crashReporter
	}

	a.authSvc = service.NewAuthService(a.ctx)
	a.gameSvc = service.NewGameService(a.ctx, a.progress, a.authSvc)

	fmt.Printf("Application starting: v%s, branch=%s, build=%d\n",
		a.launcherCfg.Version, a.instance.Branch, a.instance.BuildVersion)

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

	if err := a.SyncInstanceState(); err != nil {
		fmt.Printf("Warning: Failed to sync instance state: %v\n", err)
	}

	installedVersion, err := a.gameSvc.EnsureInstalled(a.ctx, a.instance, a.progress)
	if err != nil {
		appErr := hyerrors.WrapGame(err, "failed to install game").
			WithContext("branch", a.instance.Branch).
			WithContext("requestedVersion", a.instance.BuildVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	if installedVersion != a.instance.BuildVersion {
		a.instance.BuildVersion = installedVersion
		if err := a.UpdateInstanceVersion(installedVersion); err != nil {
			fmt.Printf("Warning: Failed to update instance version after install: %v\n", err)
		}
	}

	if err := a.gameSvc.Launch(playerName, a.instance); err != nil {
		appErr := hyerrors.GameCritical("failed to launch game").
			WithDetails(err.Error()).
			WithContext("player", playerName).
			WithContext("branch", a.instance.Branch).
			WithContext("version", a.instance.BuildVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	return nil
}

func (a *App) validatePlayerName(name string) error {
	// 3–16 characters long, consisting only of letters, numbers, and underscores
	re := regexp.MustCompile("^[A-Za-z0-9_]{3,16}$")

	if !re.MatchString(name) {
		return hyerrors.Validation("nickname should be 3-16 characters long, consisting only of letters, numbers, and underscores").
			WithContext("length", len(name)).
			WithContext("name", name)
	}

	return nil
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
			{
				Label: "GitHub",
				Url:   "https://github.com/ArchDevs/HyLauncher",
			},
			{
				Label: "Website",
				Url:   "https://hylauncher.fun",
			},
		},
	})

	if err != nil {
		fmt.Printf("failed to set Discord activity: %v\n", err)
	}
}
