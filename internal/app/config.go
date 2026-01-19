package app

import (
	"HyLauncher/internal/config"
	"fmt"
)

func (a *App) SetNick(nick string) error {
	a.cfg.Nick = nick
	return config.Save(a.cfg)
}

func (a *App) GetNick() string {
	return a.cfg.Nick
}

func (a *App) GetLauncherVersion() string {
	return config.Default().Version
}

func (a *App) SetLocalGameVersion(version int) error {
	a.cfg.CurrentGameVersion = version
	return config.Save(a.cfg)
}

func (a *App) GetLocalGameVersion() int {
	return a.cfg.CurrentGameVersion
}

func (a *App) SetBranch(branch string) error {
	err := config.SaveBranch(branch)
	if err != nil {
		return fmt.Errorf("Warning: failed to save branch")
	}
	return nil
}

func (a *App) GetBranch() (string, error) {
	branch, err := config.GetBranch()
	if err != nil {
		return "", fmt.Errorf("Warning: failed to save branch")
	}
	return branch, nil
}
