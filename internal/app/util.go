package app

import (
	"regexp"

	"HyLauncher/internal/service"
	"HyLauncher/pkg/hyerrors"
)

func (a *App) GetLogs() (string, error) {
	if a.crashSvc == nil {
		return "", hyerrors.Internal("diagnostics not initialized")
	}
	return a.crashSvc.GetLogs()
}

func (a *App) GetCrashReports() ([]service.CrashReport, error) {
	if a.crashSvc == nil {
		return nil, hyerrors.Internal("diagnostics not initialized")
	}
	return a.crashSvc.GetCrashReports()
}

func (a *App) validatePlayerName(name string) error {
	re := regexp.MustCompile("^[A-Za-z0-9_]{3,16}$")

	if !re.MatchString(name) {
		return hyerrors.Validation("nickname should be 3-16 characters long, consisting only of letters, numbers, and underscores").
			WithContext("length", len(name)).
			WithContext("name", name)
	}

	return nil
}
