package app

import (
	"HyLauncher/internal/patch"
	"HyLauncher/pkg/hyerrors"
)

// GetAvailableGameVersions exposes all available game versions for the current instance branch.
// It is intended to be called from the frontend via Wails bindings.
func (a *App) GetAvailableGameVersions() ([]int, error) {
	if a.instance.Branch == "" {
		err := hyerrors.Internal("instance branch is not configured")
		hyerrors.Report(err)
		return nil, err
	}

	versions, err := patch.ListAvailableVersions(a.instance.Branch)
	if err != nil {
		appErr := hyerrors.WrapNetwork(err, "failed to fetch available game versions").
			WithContext("branch", a.instance.Branch)
		hyerrors.Report(appErr)
		return nil, appErr
	}

	return versions, nil
}
