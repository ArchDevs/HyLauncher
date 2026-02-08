package app

import (
	"context"
	"fmt"

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
	newsSvc  *service.NewsService
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

	launcherCfg, err := config.LoadLauncher()
	if err != nil {
		panic(fmt.Errorf("failed to load launcher config: %w", err))
	}

	a.launcherCfg = launcherCfg

	if launcherCfg.DiscordRPC {
		_ = client.Login("1465005878276128888")
	}

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
	a.newsSvc = service.NewNewsService()

	go a.discordRPC()
	go env.CreateFolders(a.instance.InstanceID)
	go a.checkUpdateSilently()
	go env.CleanupLauncher(a.instance)
}
