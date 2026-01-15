package app

import "HyLauncher/internal/config"

func (a *App) SetNick(nick string) error {
	a.cfg.Nick = nick
	return config.Save(a.cfg)
}

func (a *App) GetNick() string {
	return a.cfg.Nick
}
