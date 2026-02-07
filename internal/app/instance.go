package app

import (
	"fmt"

	"HyLauncher/internal/config"
	"HyLauncher/pkg/hyerrors"
)

func (a *App) SelectInstance(instanceID string) error {
	err := config.UpdateLauncher(func(cfg *config.LauncherConfig) error {
		cfg.Instance = instanceID
		return nil
	})
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to select instance").
			WithContext("instance", instanceID)
		hyerrors.Report(appErr)
		return appErr
	}

	instanceCfg, err := config.LoadInstance(instanceID)
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to load selected instance").
			WithContext("instance", instanceID)
		hyerrors.Report(appErr)
		return appErr
	}

	a.instanceCfg = instanceCfg
	a.instance.InstanceID = instanceCfg.ID
	a.instance.InstanceName = instanceCfg.Name
	a.instance.Branch = instanceCfg.Branch
	a.instance.BuildVersion = instanceCfg.Build

	fmt.Printf("[SelectInstance] Switched to instance %s\n", instanceID)
	return nil
}

func (a *App) UpdateInstanceVersion(buildVersion string) error {
	err := config.UpdateInstance(a.instance.InstanceID, func(cfg *config.InstanceConfig) error {
		cfg.Build = buildVersion
		return nil
	})
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to update instance version").
			WithContext("instance", a.instance.InstanceID).
			WithContext("buildVersion", buildVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	a.instance.BuildVersion = buildVersion
	a.instanceCfg.Build = buildVersion

	fmt.Printf("[UpdateInstanceVersion] Updated to version %s\n", buildVersion)
	return nil
}

func (a *App) UpdateInstanceBranch(branch string) error {
	if branch != "release" && branch != "pre-release" {
		return hyerrors.Validation("invalid branch type").
			WithContext("branch", branch).
			WithDetails("branch must be either 'release' or 'pre-release'")
	}

	err := config.UpdateInstance(a.instance.InstanceID, func(cfg *config.InstanceConfig) error {
		cfg.Branch = branch
		return nil
	})
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to update instance branch").
			WithContext("instanceID", a.instance.InstanceID).
			WithContext("branch", branch)
		hyerrors.Report(appErr)
		return appErr
	}

	a.instance.Branch = branch
	a.instanceCfg.Branch = branch

	fmt.Printf("[UpdateInstanceBranch] Updated to branch %s\n", branch)
	return nil
}

func (a *App) GetInstanceInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":      a.instance.InstanceID,
		"name":    a.instance.InstanceName,
		"branch":  a.instance.Branch,
		"version": a.instance.BuildVersion,
	}
}

func (a *App) UpdateInstanceName(name string) error {
	if name == "" {
		return hyerrors.Validation("instance name cannot be empty")
	}

	err := config.UpdateInstance(a.instance.InstanceID, func(cfg *config.InstanceConfig) error {
		cfg.Name = name
		return nil
	})
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to update instance name").
			WithContext("instanceID", a.instance.InstanceID).
			WithContext("name", name)
		hyerrors.Report(appErr)
		return appErr
	}

	a.instance.InstanceName = name
	a.instanceCfg.Name = name

	fmt.Printf("[UpdateInstanceName] Updated name to %s\n", name)
	return nil
}

func (a *App) SyncInstanceState() error {
	instanceCfg, err := config.LoadInstance(a.instance.InstanceID)
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to sync instance state").
			WithContext("instance", a.instance.InstanceID)
		hyerrors.Report(appErr)
		return appErr
	}

	a.instanceCfg = instanceCfg
	a.instance.InstanceName = instanceCfg.Name
	a.instance.Branch = instanceCfg.Branch
	a.instance.BuildVersion = instanceCfg.Build

	fmt.Printf("[SyncInstanceState] Synced instance state: branch=%s, version=%s\n",
		a.instance.Branch, a.instance.BuildVersion)
	return nil
}

func (a *App) ValidateInstance() error {
	if a.instance.InstanceID == "" {
		return hyerrors.Validation("instance ID is empty")
	}

	if a.instance.Branch != "release" && a.instance.Branch != "pre-release" {
		return hyerrors.Validation("invalid branch in instance config").
			WithContext("branch", a.instance.Branch)
	}

	if a.instance.BuildVersion == "" {
		return hyerrors.Validation("build version is empty").
			WithContext("version", a.instance.BuildVersion)
	}

	return nil
}
