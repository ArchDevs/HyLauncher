package app

import "HyLauncher/internal/service"

func (a *App) GetServers() ([]service.ServerWithUrls, error) {
	return a.serversSvc.FetchServers()
}
