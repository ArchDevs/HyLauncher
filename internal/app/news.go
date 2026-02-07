package app

import "HyLauncher/internal/service"

func (a *App) GetLatestNews() (*service.NewsArticle, error) {
	return a.newsSvc.FetchLatestNews()
}

func (a *App) GetAllNews() ([]service.NewsArticle, error) {
	return a.newsSvc.FetchNews()
}
