package app

import (
	"HyLauncher/internal/config"
	"HyLauncher/pkg/hyerrors"
)

func (a *App) SetNick(nick string) error {
	if nick == "" {
		err := hyerrors.Validation("nickname cannot be empty")
		hyerrors.Report(err)
		return err
	}

	if err := config.SaveNick(nick); err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to save nickname").
			WithContext("nick", nick)
		hyerrors.Report(appErr)
		return appErr
	}

	a.cfg.Nick = nick
	return nil
}

func (a *App) GetNick() (string, error) {
	nick, err := config.GetNick()
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to get nickname")
		hyerrors.Report(appErr)
		return "", appErr
	}

	a.cfg.Nick = nick
	return nick, nil
}

func (a *App) GetLauncherVersion() string {
	return config.Default().Version
}

func (a *App) SetLocalGameVersion(version int) error {
	if err := config.SaveLocalGameVersion(version); err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to save game version").
			WithContext("version", version)
		hyerrors.Report(appErr)
		return appErr
	}

	a.cfg.CurrentGameVersion = version
	return nil
}

func (a *App) GetLocalGameVersion() (int, error) {
	version, err := config.GetLocalGameVersion()
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to get game version")
		hyerrors.Report(appErr)
		return 0, appErr
	}

	a.cfg.CurrentGameVersion = version
	return version, nil
}

func (a *App) SetBranch(branch string) error {
	if err := config.SaveBranch(branch); err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to save branch").
			WithContext("branch", branch)
		hyerrors.Report(appErr)
		return appErr
	}

	a.cfg.Branch = branch
	return nil
}

func (a *App) GetBranch() (string, error) {
	branch, err := config.GetBranch()
	if err != nil {
		appErr := hyerrors.WrapConfig(err, "failed to get branch")
		hyerrors.Report(appErr)
		return "", appErr
	}

	a.cfg.Branch = branch
	return branch, nil
}
