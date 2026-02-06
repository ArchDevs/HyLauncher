package app

import (
	"HyLauncher/internal/config"
	"HyLauncher/pkg/hyerrors"
)

func (a *App) SelectInstance(instanceID string) {
	err := config.UpdateLauncher(func(cfg *config.LauncherConfig) error {
		cfg.Instance = instanceID
		return nil
	})
	if err != nil {
		hyerrors.WrapConfig(err, "can not update config").WithContext("instance", instanceID)
	}
}
