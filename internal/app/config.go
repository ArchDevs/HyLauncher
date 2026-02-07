package app

import (
	"HyLauncher/internal/config"
	"HyLauncher/pkg/hyerrors"
)

func (a *App) SetNick(nick, instanceID string) error {
	if nick == "" {
		err := hyerrors.Validation("nickname cannot be empty")
		hyerrors.Report(err)
		return err
	}

	err := config.UpdateLauncher(func(cfg *config.LauncherConfig) error {
		cfg.Nick = nick
		return nil
	})

	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to save nickname").
			WithContext("nick", nick)
		hyerrors.Report(appErr)
		return appErr
	}

	a.launcherCfg.Nick = nick
	return nil
}

func (a *App) GetNick() (string, error) {
	cfg, err := config.LoadLauncher()
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to get nickname")
		hyerrors.Report(appErr)
		return "", appErr
	}

	a.launcherCfg.Nick = cfg.Nick
	return cfg.Nick, nil
}

func (a *App) GetLauncherVersion() string {
	return config.LauncherDefault().Version
}

func (a *App) SetLocalGameVersion(version string, instanceID string) error {
	if version == "" {
		err := hyerrors.Validation("game version cannot be empty")
		hyerrors.Report(err)
		return err
	}

	err := config.UpdateInstance(instanceID, func(cfg *config.InstanceConfig) error {
		cfg.Build = version
		return nil
	})

	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to save game version").
			WithContext("version", version).
			WithContext("instance", instanceID)
		hyerrors.Report(appErr)
		return appErr
	}

	a.instanceCfg.Build = version
	return nil
}

func (a *App) GetLocalGameVersion(instanceID string) (string, error) {
	cfg, err := config.LoadInstance(instanceID)
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to get game version").
			WithContext("instance", instanceID)
		hyerrors.Report(appErr)
		return "", appErr
	}

	a.instanceCfg.Build = cfg.Build
	return cfg.Build, nil
}
